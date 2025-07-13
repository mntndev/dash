package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mntndev/dash/pkg/config"
)

type HAProvider interface {
	GetHAClient() *HomeAssistantClient
}

type HomeAssistantClient struct {
	config        *config.HomeAssistantConfig
	conn          *websocket.Conn
	connected     bool
	authenticated bool
	msgID         int
	callbacks     map[int]func(HAMessage)
	mu            sync.RWMutex
	writeMu       sync.Mutex
	ctx           context.Context
	cancel        context.CancelFunc
	eventChan     chan HAEvent
	authChan      chan bool
	*SubscriptionManager
}

type HAMessage struct {
	ID      int                    `json:"id,omitempty"`
	Type    string                 `json:"type"`
	Success bool                   `json:"success,omitempty"`
	Result  interface{}            `json:"result,omitempty"`
	Error   *HAError               `json:"error,omitempty"`
	Event   *HAEvent               `json:"event,omitempty"`
	Data    map[string]interface{} `json:",inline"`
}

type HAError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type HAEvent struct {
	EventType string                 `json:"event_type"`
	Data      map[string]interface{} `json:"data"`
	Origin    string                 `json:"origin"`
	TimeFired time.Time              `json:"time_fired"`
}

type HAEntityState struct {
	EntityID    string                 `json:"entity_id"`
	State       string                 `json:"state"`
	Attributes  map[string]interface{} `json:"attributes"`
	LastChanged time.Time              `json:"last_changed"`
	LastUpdated time.Time              `json:"last_updated"`
}

type StateChangeEvent struct {
	EntityID string         `json:"entity_id"`
	NewState *HAEntityState `json:"new_state"`
	OldState *HAEntityState `json:"old_state"`
}

type SubscriptionManager struct {
	haClient      *HomeAssistantClient
	subscriptions map[string][]chan StateChangeEvent
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewHomeAssistantClient(cfg *config.HomeAssistantConfig) *HomeAssistantClient {
	ctx, cancel := context.WithCancel(context.Background())
	client := &HomeAssistantClient{
		config:    cfg,
		callbacks: make(map[int]func(HAMessage)),
		ctx:       ctx,
		cancel:    cancel,
		eventChan: make(chan HAEvent, 100),
		authChan:  make(chan bool, 1),
	}
	client.SubscriptionManager = NewSubscriptionManager(client)
	return client
}

func (ha *HomeAssistantClient) Connect() error {
	u, err := url.Parse(ha.config.URL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	ha.conn = conn
	ha.connected = true

	go ha.readMessages()
	select {
	case success := <-ha.authChan:
		if !success {
			return fmt.Errorf("authentication failed")
		}
		return nil
	case <-time.After(10 * time.Second):
		return fmt.Errorf("authentication timeout")
	}
}

func (ha *HomeAssistantClient) authenticate(msg HAMessage) error {
	if msg.Type != "auth_required" {
		return fmt.Errorf("expected auth_required, got %s", msg.Type)
	}

	authMsg := map[string]interface{}{
		"type":         "auth",
		"access_token": ha.config.Token,
	}

	ha.writeMu.Lock()
	err := ha.conn.WriteJSON(authMsg)
	ha.writeMu.Unlock()

	if err != nil {
		return fmt.Errorf("failed to send auth message: %w", err)
	}

	return nil
}

func (ha *HomeAssistantClient) readMessages() {
	defer func() {
		ha.mu.Lock()
		ha.connected = false
		ha.authenticated = false
		ha.mu.Unlock()
		if ha.conn != nil {
			if err := ha.conn.Close(); err != nil {
				log.Printf("Failed to close WebSocket connection: %v", err)
			}
		}
	}()

	for {
		select {
		case <-ha.ctx.Done():
			return
		default:
			var msg HAMessage
			if err := ha.conn.ReadJSON(&msg); err != nil {
				if !ha.authenticated {
					select {
					case ha.authChan <- false:
					default:
					}
				}
				return
			}

			ha.handleMessage(msg)
		}
	}
}

func (ha *HomeAssistantClient) handleMessage(msg HAMessage) {
	if msg.Type == "auth_required" {
		if err := ha.authenticate(msg); err != nil {
			select {
			case ha.authChan <- false:
			default:
			}
		}
		return
	}

	if msg.Type == "auth_ok" {
		ha.mu.Lock()
		ha.authenticated = true
		ha.mu.Unlock()
		select {
		case ha.authChan <- true:
		default:
		}
		return
	}

	if msg.Type == "auth_invalid" {
		select {
		case ha.authChan <- false:
		default:
		}
		return
	}

	if msg.ID > 0 {
		ha.mu.RLock()
		callback, exists := ha.callbacks[msg.ID]
		ha.mu.RUnlock()

		if exists {
			callback(msg)
			ha.mu.Lock()
			delete(ha.callbacks, msg.ID)
			ha.mu.Unlock()
		}
	}

	if msg.Type == "event" && msg.Event != nil {
		select {
		case ha.eventChan <- *msg.Event:
		default:
			// Event channel full, drop event silently
		}
	}
}

func (ha *HomeAssistantClient) sendMessage(msgType string, data map[string]interface{}) (int, error) {
	ha.mu.RLock()
	connected := ha.connected
	authenticated := ha.authenticated
	ha.mu.RUnlock()

	if !connected {
		return 0, fmt.Errorf("not connected")
	}

	if !authenticated {
		return 0, fmt.Errorf("not authenticated")
	}

	ha.mu.Lock()
	ha.msgID++
	id := ha.msgID
	ha.mu.Unlock()

	msg := map[string]interface{}{
		"id":   id,
		"type": msgType,
	}

	for k, v := range data {
		msg[k] = v
	}

	ha.writeMu.Lock()
	err := ha.conn.WriteJSON(msg)
	ha.writeMu.Unlock()

	if err != nil {
		return 0, fmt.Errorf("failed to send message: %w", err)
	}

	return id, nil
}

func (ha *HomeAssistantClient) GetStates() ([]HAEntityState, error) {
	id, err := ha.sendMessage("get_states", nil)
	if err != nil {
		return nil, err
	}

	resultChan := make(chan HAMessage, 1)
	ha.mu.Lock()
	ha.callbacks[id] = func(msg HAMessage) {
		resultChan <- msg
	}
	ha.mu.Unlock()

	select {
	case msg := <-resultChan:
		if !msg.Success {
			return nil, fmt.Errorf("get_states failed: %v", msg.Error)
		}

		var states []HAEntityState
		statesData, err := json.Marshal(msg.Result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal states data: %w", err)
		}
		if err := json.Unmarshal(statesData, &states); err != nil {
			return nil, fmt.Errorf("failed to parse states: %w", err)
		}

		return states, nil
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("timeout waiting for states")
	}
}

func (ha *HomeAssistantClient) CallService(domain, service string, data map[string]interface{}) error {
	id, err := ha.sendMessage("call_service", map[string]interface{}{
		"domain":       domain,
		"service":      service,
		"service_data": data,
	})
	if err != nil {
		return err
	}

	resultChan := make(chan HAMessage, 1)
	ha.mu.Lock()
	ha.callbacks[id] = func(msg HAMessage) {
		resultChan <- msg
	}
	ha.mu.Unlock()

	select {
	case msg := <-resultChan:
		if !msg.Success {
			return fmt.Errorf("service call failed: %v", msg.Error)
		}
		return nil
	case <-time.After(10 * time.Second):
		return fmt.Errorf("timeout waiting for service call")
	}
}

func (ha *HomeAssistantClient) SubscribeEvents(eventType string) error {
	_, err := ha.sendMessage("subscribe_events", map[string]interface{}{
		"event_type": eventType,
	})
	return err
}

func (ha *HomeAssistantClient) GetEventChannel() <-chan HAEvent {
	return ha.eventChan
}

func (ha *HomeAssistantClient) Close() error {
	ha.cancel()
	if ha.SubscriptionManager != nil {
		ha.SubscriptionManager.Close()
	}
	if ha.conn != nil {
		return ha.conn.Close()
	}
	return nil
}

func (ha *HomeAssistantClient) IsConnected() bool {
	ha.mu.RLock()
	defer ha.mu.RUnlock()
	return ha.connected && ha.authenticated
}

func NewSubscriptionManager(haClient *HomeAssistantClient) *SubscriptionManager {
	ctx, cancel := context.WithCancel(context.Background())
	sm := &SubscriptionManager{
		haClient:      haClient,
		subscriptions: make(map[string][]chan StateChangeEvent),
		ctx:           ctx,
		cancel:        cancel,
	}
	go sm.eventProcessor()
	return sm
}

func (sm *SubscriptionManager) Subscribe(entityID string) (<-chan StateChangeEvent, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	ch := make(chan StateChangeEvent, 10)

	if _, exists := sm.subscriptions[entityID]; !exists {
		if err := sm.haClient.SubscribeEvents("state_changed"); err != nil {
			close(ch)
			return nil, fmt.Errorf("failed to subscribe to state_changed events: %w", err)
		}
	}

	sm.subscriptions[entityID] = append(sm.subscriptions[entityID], ch)
	return ch, nil
}

func (sm *SubscriptionManager) Unsubscribe(entityID string, ch <-chan StateChangeEvent) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if subs, exists := sm.subscriptions[entityID]; exists {
		for i, sub := range subs {
			if sub == ch {
				close(sub)
				sm.subscriptions[entityID] = append(subs[:i], subs[i+1:]...)
				break
			}
		}

		if len(sm.subscriptions[entityID]) == 0 {
			delete(sm.subscriptions, entityID)
		}
	}
}

func (sm *SubscriptionManager) eventProcessor() {
	eventChan := sm.haClient.GetEventChannel()

	for {
		select {
		case <-sm.ctx.Done():
			sm.closeAllSubscriptions()
			return
		case event := <-eventChan:
			if event.EventType == "state_changed" {
				sm.processStateChangeEvent(event)
			}
		}
	}
}

func (sm *SubscriptionManager) processStateChangeEvent(event HAEvent) {
	data := event.Data
	entityID, ok := data["entity_id"].(string)
	if !ok {
		return
	}

	var newState, oldState *HAEntityState

	if newStateData, ok := data["new_state"]; ok && newStateData != nil {
		newStateBytes, err := json.Marshal(newStateData)
		if err != nil {
			log.Printf("Failed to marshal new state data: %v", err)
		} else {
			newState = &HAEntityState{}
			if err := json.Unmarshal(newStateBytes, newState); err != nil {
				log.Printf("Failed to unmarshal new state: %v", err)
				newState = nil
			}
		}
	}

	if oldStateData, ok := data["old_state"]; ok && oldStateData != nil {
		oldStateBytes, err := json.Marshal(oldStateData)
		if err != nil {
			log.Printf("Failed to marshal old state data: %v", err)
		} else {
			oldState = &HAEntityState{}
			if err := json.Unmarshal(oldStateBytes, oldState); err != nil {
				log.Printf("Failed to unmarshal old state: %v", err)
				oldState = nil
			}
		}
	}

	stateEvent := StateChangeEvent{
		EntityID: entityID,
		NewState: newState,
		OldState: oldState,
	}

	sm.mu.RLock()
	subscribers := sm.subscriptions[entityID]
	sm.mu.RUnlock()

	for _, ch := range subscribers {
		select {
		case ch <- stateEvent:
		default:
		}
	}
}

func (sm *SubscriptionManager) closeAllSubscriptions() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for entityID, subs := range sm.subscriptions {
		for _, ch := range subs {
			close(ch)
		}
		delete(sm.subscriptions, entityID)
	}
}

func (sm *SubscriptionManager) Close() {
	sm.cancel()
}
