package widgets

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mntndev/dash/pkg/integrations"
)

type HABaseWidget struct {
	*BaseWidget
	EntityID     string
	haProvider   integrations.HAProvider
	provider     Provider
	subscription <-chan integrations.StateChangeEvent
	cancelSub    context.CancelFunc
}

type HAEntityWidget struct {
	*HABaseWidget
}

type HAButtonWidget struct {
	*HABaseWidget
	Service string
	Domain  string
}

type HASwitchWidget struct {
	*HABaseWidget
}

type HALightWidget struct {
	*HABaseWidget
}

type HAEntityData struct {
	EntityID    string                 `json:"entity_id"`
	State       string                 `json:"state"`
	Attributes  map[string]interface{} `json:"attributes"`
	LastChanged time.Time              `json:"last_changed"`
	LastUpdated time.Time              `json:"last_updated"`
}

func (hab *HABaseWidget) setDataAndEmit(data interface{}) {
	hab.Data = data
	hab.LastUpdate = time.Now()
	if hab.provider != nil {
		hab.provider.Emit("widget_data_update", map[string]interface{}{
			"widget_id": hab.ID,
			"data":      data,
		})
	}
}

type HAButtonData struct {
	EntityID string `json:"entity_id"`
	Service  string `json:"service"`
	Domain   string `json:"domain"`
	Label    string `json:"label"`
}

func (hab *HABaseWidget) startSubscription(ctx context.Context) error {
	haClient := hab.haProvider.GetHAClient()
	if haClient == nil || !haClient.IsConnected() {
		// HA client not connected yet, start a goroutine to wait for connection
		go hab.waitForConnectionAndSubscribe(ctx)
		return nil
	}

	return hab.setupSubscription(ctx)
}

func (hab *HABaseWidget) waitForConnectionAndSubscribe(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			haClient := hab.haProvider.GetHAClient()
			if haClient != nil && haClient.IsConnected() {
				if err := hab.setupSubscription(ctx); err != nil {
					// Log error but don't fail - widget can try again
					fmt.Printf("Failed to setup HA subscription for %s: %v\n", hab.EntityID, err)
				} else {
					fmt.Printf("HA widget %s successfully connected and subscribed\n", hab.EntityID)
				}
				return
			}
		}
	}
}

func (hab *HABaseWidget) setupSubscription(ctx context.Context) error {
	haClient := hab.haProvider.GetHAClient()
	if haClient == nil || !haClient.IsConnected() {
		return fmt.Errorf("home Assistant client not connected")
	}

	subscription, err := haClient.Subscribe(hab.EntityID)
	if err != nil {
		return fmt.Errorf("failed to subscribe to entity %s: %w", hab.EntityID, err)
	}

	hab.subscription = subscription

	ctx, cancel := context.WithCancel(ctx)
	hab.cancelSub = cancel

	go hab.processStateChanges(ctx)

	return hab.fetchInitialState()
}

func (hab *HABaseWidget) processStateChanges(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-hab.subscription:
			log.Printf("%v", event)
			if event.NewState != nil {
				data := &HAEntityData{
					EntityID:    event.NewState.EntityID,
					State:       event.NewState.State,
					Attributes:  event.NewState.Attributes,
					LastChanged: event.NewState.LastChanged,
					LastUpdated: event.NewState.LastUpdated,
				}
				hab.setDataAndEmit(data)
			}
		}
	}
}

func (hab *HABaseWidget) fetchInitialState() error {
	haClient := hab.haProvider.GetHAClient()
	if haClient == nil || !haClient.IsConnected() {
		return fmt.Errorf("home Assistant client not connected")
	}

	states, err := haClient.GetStates()
	if err != nil {
		return fmt.Errorf("failed to get states: %w", err)
	}

	for _, state := range states {
		if state.EntityID == hab.EntityID {
			data := &HAEntityData{
				EntityID:    state.EntityID,
				State:       state.State,
				Attributes:  state.Attributes,
				LastChanged: state.LastChanged,
				LastUpdated: state.LastUpdated,
			}
			hab.setDataAndEmit(data)
			return nil
		}
	}

	return fmt.Errorf("entity %s not found", hab.EntityID)
}

func (hab *HABaseWidget) stopSubscription() {
	if hab.cancelSub != nil {
		hab.cancelSub()
	}
	if hab.subscription != nil {
		haClient := hab.haProvider.GetHAClient()
		if haClient != nil {
			haClient.Unsubscribe(hab.EntityID, hab.subscription)
		}
	}
}

func CreateHAEntityWidget(id string, config map[string]interface{}, children []Widget, provider Provider) (Widget, error) {
	var haConfig HAEntityConfig
	parser := NewConfigParser()
	if err := parser.ParseConfig(config, &haConfig); err != nil {
		return nil, fmt.Errorf("invalid HA entity configuration: %w", err)
	}

	widget := &HAEntityWidget{
		HABaseWidget: &HABaseWidget{
			BaseWidget: &BaseWidget{
				ID:       id,
				Type:     "home_assistant.entity",
				Config:   config,
				Children: children,
			},
			EntityID:   haConfig.EntityID,
			haProvider: provider,
			provider:   provider,
		},
	}

	return widget, nil
}

func CreateHASwitchWidget(id string, config map[string]interface{}, children []Widget, provider Provider) (Widget, error) {
	var haConfig HAEntityConfig
	parser := NewConfigParser()
	if err := parser.ParseConfig(config, &haConfig); err != nil {
		return nil, fmt.Errorf("invalid HA switch configuration: %w", err)
	}

	widget := &HASwitchWidget{
		HABaseWidget: &HABaseWidget{
			BaseWidget: &BaseWidget{
				ID:       id,
				Type:     "home_assistant.switch",
				Config:   config,
				Children: children,
			},
			EntityID:   haConfig.EntityID,
			haProvider: provider,
			provider:   provider,
		},
	}

	return widget, nil
}

func CreateHALightWidget(id string, config map[string]interface{}, children []Widget, provider Provider) (Widget, error) {
	var haConfig HAEntityConfig
	parser := NewConfigParser()
	if err := parser.ParseConfig(config, &haConfig); err != nil {
		return nil, fmt.Errorf("invalid HA light configuration: %w", err)
	}

	widget := &HALightWidget{
		HABaseWidget: &HABaseWidget{
			BaseWidget: &BaseWidget{
				ID:       id,
				Type:     "home_assistant.light",
				Config:   config,
				Children: children,
			},
			EntityID:   haConfig.EntityID,
			haProvider: provider,
			provider:   provider,
		},
	}

	return widget, nil
}

func CreateHAButtonWidget(id string, config map[string]interface{}, children []Widget, provider Provider) (Widget, error) {
	var buttonConfig HAButtonConfig
	parser := NewConfigParser()
	if err := parser.ParseConfig(config, &buttonConfig); err != nil {
		return nil, fmt.Errorf("invalid HA button configuration: %w", err)
	}

	widget := &HAButtonWidget{
		HABaseWidget: &HABaseWidget{
			BaseWidget: &BaseWidget{
				ID:       id,
				Type:     "home_assistant.button",
				Config:   config,
				Children: children,
			},
			EntityID:   buttonConfig.EntityID,
			haProvider: provider,
			provider:   provider,
		},
		Service: buttonConfig.Service,
		Domain:  buttonConfig.Domain,
	}

	widget.Data = &HAButtonData{
		EntityID: buttonConfig.EntityID,
		Service:  buttonConfig.Service,
		Domain:   buttonConfig.Domain,
		Label:    buttonConfig.Label,
	}

	return widget, nil
}

func (w *HAEntityWidget) Init(ctx context.Context) error {
	w.LastUpdate = time.Now()
	// Initialize with empty data to avoid null
	w.Data = &HAEntityData{
		EntityID:    w.EntityID,
		State:       "unknown",
		Attributes:  make(map[string]interface{}),
		LastChanged: time.Now(),
		LastUpdated: time.Now(),
	}

	// Try to fetch initial state asynchronously with a short delay to avoid blocking
	go func() {
		time.Sleep(100 * time.Millisecond) // Brief delay to let HA client fully initialize
		if err := w.fetchInitialState(); err != nil {
			fmt.Printf("Failed to fetch initial state for %s: %v\n", w.EntityID, err)
		}
	}()

	// Start subscription asynchronously to avoid blocking widget initialization
	go func() {
		if err := w.startSubscription(ctx); err != nil {
			fmt.Printf("Failed to start subscription for HA widget %s: %v\n", w.EntityID, err)
		}
	}()
	return nil
}

func (w *HASwitchWidget) Init(ctx context.Context) error {
	w.LastUpdate = time.Now()
	// Initialize with empty data to avoid null
	w.Data = &HAEntityData{
		EntityID:    w.EntityID,
		State:       "unknown",
		Attributes:  make(map[string]interface{}),
		LastChanged: time.Now(),
		LastUpdated: time.Now(),
	}

	// Try to fetch initial state asynchronously with a short delay to avoid blocking
	go func() {
		time.Sleep(100 * time.Millisecond) // Brief delay to let HA client fully initialize
		if err := w.fetchInitialState(); err != nil {
			fmt.Printf("Failed to fetch initial state for %s: %v\n", w.EntityID, err)
		}
	}()
	// Start subscription asynchronously to avoid blocking widget initialization
	go func() {
		if err := w.startSubscription(ctx); err != nil {
			fmt.Printf("Failed to start subscription for HA widget %s: %v\n", w.EntityID, err)
		}
	}()
	return nil
}

func (w *HALightWidget) Init(ctx context.Context) error {
	w.LastUpdate = time.Now()
	// Initialize with empty data to avoid null
	w.Data = &HAEntityData{
		EntityID:    w.EntityID,
		State:       "unknown",
		Attributes:  make(map[string]interface{}),
		LastChanged: time.Now(),
		LastUpdated: time.Now(),
	}

	// Try to fetch initial state asynchronously with a short delay to avoid blocking
	go func() {
		time.Sleep(100 * time.Millisecond) // Brief delay to let HA client fully initialize
		if err := w.fetchInitialState(); err != nil {
			fmt.Printf("Failed to fetch initial state for %s: %v\n", w.EntityID, err)
		}
	}()

	// Start subscription asynchronously to avoid blocking widget initialization
	go func() {
		if err := w.startSubscription(ctx); err != nil {
			fmt.Printf("Failed to start subscription for HA widget %s: %v\n", w.EntityID, err)
		}
	}()
	return nil
}

func (w *HAButtonWidget) Init(ctx context.Context) error {
	w.LastUpdate = time.Now()
	return nil
}

func (w *HASwitchWidget) Trigger() error {
	haClient := w.haProvider.GetHAClient()
	if haClient == nil || !haClient.IsConnected() {
		return fmt.Errorf("home Assistant client not connected")
	}

	serviceData := map[string]interface{}{
		"entity_id": w.EntityID,
	}

	return haClient.CallService("switch", "toggle", serviceData)
}

func (w *HALightWidget) Trigger() error {
	haClient := w.haProvider.GetHAClient()
	if haClient == nil || !haClient.IsConnected() {
		return fmt.Errorf("home Assistant client not connected")
	}

	serviceData := map[string]interface{}{
		"entity_id": w.EntityID,
	}

	return haClient.CallService("light", "toggle", serviceData)
}

func (w *HALightWidget) SetBrightness(brightness int) error {
	haClient := w.haProvider.GetHAClient()
	if haClient == nil || !haClient.IsConnected() {
		return fmt.Errorf("home Assistant client not connected")
	}

	serviceData := map[string]interface{}{
		"entity_id":  w.EntityID,
		"brightness": brightness,
	}

	return haClient.CallService("light", "turn_on", serviceData)
}

func (w *HAButtonWidget) Trigger() error {
	haClient := w.haProvider.GetHAClient()
	if haClient == nil || !haClient.IsConnected() {
		return fmt.Errorf("home Assistant client not connected")
	}

	serviceData := map[string]interface{}{
		"entity_id": w.EntityID,
	}

	return haClient.CallService(w.Domain, w.Service, serviceData)
}

func (w *HAEntityWidget) Close() error {
	w.stopSubscription()
	return nil
}

func (w *HASwitchWidget) Close() error {
	w.stopSubscription()
	return nil
}

func (w *HALightWidget) Close() error {
	w.stopSubscription()
	return nil
}

func (w *HAButtonWidget) Close() error {
	return nil
}
