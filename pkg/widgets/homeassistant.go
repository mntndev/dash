package widgets

import (
	"context"
	"fmt"
	"time"
)

type HAEntityWidget struct {
	*BaseWidget
	EntityID string
	service  ServiceProvider
}

type HAButtonWidget struct {
	*BaseWidget
	EntityID string
	Service  string
	Domain   string
	service  ServiceProvider
}

type HASwitchWidget struct {
	*BaseWidget
	EntityID string
	service  ServiceProvider
}

type HALightWidget struct {
	*BaseWidget
	EntityID string
	service  ServiceProvider
}

type HAEntityData struct {
	EntityID    string                 `json:"entity_id"`
	State       string                 `json:"state"`
	Attributes  map[string]interface{} `json:"attributes"`
	LastChanged time.Time              `json:"last_changed"`
	LastUpdated time.Time              `json:"last_updated"`
}

type HAButtonData struct {
	EntityID string `json:"entity_id"`
	Service  string `json:"service"`
	Domain   string `json:"domain"`
	Label    string `json:"label"`
}

func CreateHAEntityWidget(config map[string]interface{}, service ServiceProvider) (Widget, error) {
	entityID, ok := config["entity_id"].(string)
	if !ok {
		return nil, fmt.Errorf("entity_id is required for home_assistant.entity widget")
	}

	widget := &HAEntityWidget{
		BaseWidget: &BaseWidget{
			ID:       generateWidgetID(),
			Type:     "home_assistant.entity",
			Config:   config,
		},
		EntityID: entityID,
		service:  service,
	}

	return widget, nil
}

func CreateHASwitchWidget(config map[string]interface{}, service ServiceProvider) (Widget, error) {
	entityID, ok := config["entity_id"].(string)
	if !ok {
		return nil, fmt.Errorf("entity_id is required for home_assistant.switch widget")
	}

	widget := &HASwitchWidget{
		BaseWidget: &BaseWidget{
			ID:       generateWidgetID(),
			Type:     "home_assistant.switch",
			Config:   config,
		},
		EntityID: entityID,
		service:  service,
	}

	return widget, nil
}

func CreateHALightWidget(config map[string]interface{}, service ServiceProvider) (Widget, error) {
	entityID, ok := config["entity_id"].(string)
	if !ok {
		return nil, fmt.Errorf("entity_id is required for home_assistant.light widget")
	}

	widget := &HALightWidget{
		BaseWidget: &BaseWidget{
			ID:       generateWidgetID(),
			Type:     "home_assistant.light",
			Config:   config,
		},
		EntityID: entityID,
		service:  service,
	}

	return widget, nil
}

func CreateHAButtonWidget(config map[string]interface{}, service ServiceProvider) (Widget, error) {
	entityID, ok := config["entity_id"].(string)
	if !ok {
		return nil, fmt.Errorf("entity_id is required for home_assistant.button widget")
	}

	serviceName, ok := config["service"].(string)
	if !ok {
		return nil, fmt.Errorf("service is required for home_assistant.button widget")
	}

	domain, ok := config["domain"].(string)
	if !ok {
		return nil, fmt.Errorf("domain is required for home_assistant.button widget")
	}

	widget := &HAButtonWidget{
		BaseWidget: &BaseWidget{
			ID:       generateWidgetID(),
			Type:     "home_assistant.button",
			Config:   config,
		},
		EntityID: entityID,
		Service:  serviceName,
		Domain:   domain,
		service:  service,
	}

	widget.Data = &HAButtonData{
		EntityID: entityID,
		Service:  serviceName,
		Domain:   domain,
		Label:    getStringConfig(config, "label", "Button"),
	}

	return widget, nil
}

func (w *HAEntityWidget) Update(ctx context.Context) error {
	haClient := w.service.GetHAClient()
	if haClient == nil || !haClient.IsConnected() {
		return fmt.Errorf("Home Assistant client not connected")
	}

	states, err := haClient.GetStates()
	if err != nil {
		return fmt.Errorf("failed to get states: %w", err)
	}

	for _, state := range states {
		if state.EntityID == w.EntityID {
			w.Data = &HAEntityData{
				EntityID:    state.EntityID,
				State:       state.State,
				Attributes:  state.Attributes,
				LastChanged: state.LastChanged,
				LastUpdated: state.LastUpdated,
			}
			w.LastUpdate = time.Now()
			return nil
		}
	}

	return fmt.Errorf("entity %s not found", w.EntityID)
}

func (w *HASwitchWidget) Update(ctx context.Context) error {
	haClient := w.service.GetHAClient()
	if haClient == nil || !haClient.IsConnected() {
		return fmt.Errorf("Home Assistant client not connected")
	}

	states, err := haClient.GetStates()
	if err != nil {
		return fmt.Errorf("failed to get states: %w", err)
	}

	for _, state := range states {
		if state.EntityID == w.EntityID {
			w.Data = &HAEntityData{
				EntityID:    state.EntityID,
				State:       state.State,
				Attributes:  state.Attributes,
				LastChanged: state.LastChanged,
				LastUpdated: state.LastUpdated,
			}
			w.LastUpdate = time.Now()
			return nil
		}
	}

	return fmt.Errorf("entity %s not found", w.EntityID)
}

func (w *HALightWidget) Update(ctx context.Context) error {
	haClient := w.service.GetHAClient()
	if haClient == nil || !haClient.IsConnected() {
		return fmt.Errorf("Home Assistant client not connected")
	}

	states, err := haClient.GetStates()
	if err != nil {
		return fmt.Errorf("failed to get states: %w", err)
	}

	for _, state := range states {
		if state.EntityID == w.EntityID {
			w.Data = &HAEntityData{
				EntityID:    state.EntityID,
				State:       state.State,
				Attributes:  state.Attributes,
				LastChanged: state.LastChanged,
				LastUpdated: state.LastUpdated,
			}
			w.LastUpdate = time.Now()
			return nil
		}
	}

	return fmt.Errorf("entity %s not found", w.EntityID)
}

func (w *HAButtonWidget) Update(ctx context.Context) error {
	w.LastUpdate = time.Now()
	return nil
}

// SetHAClient methods removed - clients are now injected during widget creation

func (w *HASwitchWidget) Trigger() error {
	haClient := w.service.GetHAClient()
	if haClient == nil || !haClient.IsConnected() {
		return fmt.Errorf("Home Assistant client not connected")
	}

	serviceData := map[string]interface{}{
		"entity_id": w.EntityID,
	}

	return haClient.CallService("switch", "toggle", serviceData)
}

func (w *HALightWidget) Trigger() error {
	haClient := w.service.GetHAClient()
	if haClient == nil || !haClient.IsConnected() {
		return fmt.Errorf("Home Assistant client not connected")
	}

	serviceData := map[string]interface{}{
		"entity_id": w.EntityID,
	}

	return haClient.CallService("light", "toggle", serviceData)
}

func (w *HALightWidget) SetBrightness(brightness int) error {
	haClient := w.service.GetHAClient()
	if haClient == nil || !haClient.IsConnected() {
		return fmt.Errorf("Home Assistant client not connected")
	}

	serviceData := map[string]interface{}{
		"entity_id":  w.EntityID,
		"brightness": brightness,
	}

	return haClient.CallService("light", "turn_on", serviceData)
}

func (w *HAButtonWidget) Trigger() error {
	haClient := w.service.GetHAClient()
	if haClient == nil || !haClient.IsConnected() {
		return fmt.Errorf("Home Assistant client not connected")
	}

	serviceData := map[string]interface{}{
		"entity_id": w.EntityID,
	}

	return haClient.CallService(w.Domain, w.Service, serviceData)
}

func generateWidgetID() string {
	return fmt.Sprintf("widget_%d", time.Now().UnixNano())
}

func getStringConfig(config map[string]interface{}, key, defaultValue string) string {
	if val, ok := config[key].(string); ok {
		return val
	}
	return defaultValue
}