package widgets

import (
	"context"
	"fmt"
	"time"

	"github.com/mntndev/dash/pkg/integrations"
)

type HABaseWidget struct {
	*BaseWidget
	EntityID string
	haProvider integrations.HAProvider
}

type HAEntityWidget struct {
	*HABaseWidget
}

type HAButtonWidget struct {
	*HABaseWidget
	Service  string
	Domain   string
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

type HAButtonData struct {
	EntityID string `json:"entity_id"`
	Service  string `json:"service"`
	Domain   string `json:"domain"`
	Label    string `json:"label"`
}

func (hab *HABaseWidget) fetchEntityState() (*HAEntityData, error) {
	haClient := hab.haProvider.GetHAClient()
	if haClient == nil || !haClient.IsConnected() {
		return nil, fmt.Errorf("Home Assistant client not connected")
	}

	states, err := haClient.GetStates()
	if err != nil {
		return nil, fmt.Errorf("failed to get states: %w", err)
	}

	for _, state := range states {
		if state.EntityID == hab.EntityID {
			return &HAEntityData{
				EntityID:    state.EntityID,
				State:       state.State,
				Attributes:  state.Attributes,
				LastChanged: state.LastChanged,
				LastUpdated: state.LastUpdated,
			}, nil
		}
	}

	return nil, fmt.Errorf("entity %s not found", hab.EntityID)
}

func CreateHAEntityWidget(config map[string]interface{}, haProvider integrations.HAProvider) (Widget, error) {
	var haConfig HAEntityConfig
	parser := NewConfigParser()
	if err := parser.ParseConfig(config, &haConfig); err != nil {
		return nil, fmt.Errorf("invalid HA entity configuration: %w", err)
	}

	widget := &HAEntityWidget{
		HABaseWidget: &HABaseWidget{
			BaseWidget: &BaseWidget{
				ID:       generateWidgetID(),
				Type:     "home_assistant.entity",
				Config:   config,
			},
			EntityID: haConfig.EntityID,
			haProvider: haProvider,
		},
	}

	return widget, nil
}

func CreateHASwitchWidget(config map[string]interface{}, haProvider integrations.HAProvider) (Widget, error) {
	var haConfig HAEntityConfig
	parser := NewConfigParser()
	if err := parser.ParseConfig(config, &haConfig); err != nil {
		return nil, fmt.Errorf("invalid HA switch configuration: %w", err)
	}

	widget := &HASwitchWidget{
		HABaseWidget: &HABaseWidget{
			BaseWidget: &BaseWidget{
				ID:       generateWidgetID(),
				Type:     "home_assistant.switch",
				Config:   config,
			},
			EntityID: haConfig.EntityID,
			haProvider: haProvider,
		},
	}

	return widget, nil
}

func CreateHALightWidget(config map[string]interface{}, haProvider integrations.HAProvider) (Widget, error) {
	var haConfig HAEntityConfig
	parser := NewConfigParser()
	if err := parser.ParseConfig(config, &haConfig); err != nil {
		return nil, fmt.Errorf("invalid HA light configuration: %w", err)
	}

	widget := &HALightWidget{
		HABaseWidget: &HABaseWidget{
			BaseWidget: &BaseWidget{
				ID:       generateWidgetID(),
				Type:     "home_assistant.light",
				Config:   config,
			},
			EntityID: haConfig.EntityID,
			haProvider: haProvider,
		},
	}

	return widget, nil
}

func CreateHAButtonWidget(config map[string]interface{}, haProvider integrations.HAProvider) (Widget, error) {
	var buttonConfig HAButtonConfig
	parser := NewConfigParser()
	if err := parser.ParseConfig(config, &buttonConfig); err != nil {
		return nil, fmt.Errorf("invalid HA button configuration: %w", err)
	}

	widget := &HAButtonWidget{
		HABaseWidget: &HABaseWidget{
			BaseWidget: &BaseWidget{
				ID:       generateWidgetID(),
				Type:     "home_assistant.button",
				Config:   config,
			},
			EntityID: buttonConfig.EntityID,
			haProvider: haProvider,
		},
		Service:  buttonConfig.Service,
		Domain:   buttonConfig.Domain,
	}

	widget.Data = &HAButtonData{
		EntityID: buttonConfig.EntityID,
		Service:  buttonConfig.Service,
		Domain:   buttonConfig.Domain,
		Label:    buttonConfig.Label,
	}

	return widget, nil
}

func (w *HAEntityWidget) Update(ctx context.Context) error {
	entityData, err := w.HABaseWidget.fetchEntityState()
	if err != nil {
		return err
	}
	w.Data = entityData
	w.LastUpdate = time.Now()
	return nil
}

func (w *HASwitchWidget) Update(ctx context.Context) error {
	entityData, err := w.HABaseWidget.fetchEntityState()
	if err != nil {
		return err
	}
	w.Data = entityData
	w.LastUpdate = time.Now()
	return nil
}

func (w *HALightWidget) Update(ctx context.Context) error {
	entityData, err := w.HABaseWidget.fetchEntityState()
	if err != nil {
		return err
	}
	w.Data = entityData
	w.LastUpdate = time.Now()
	return nil
}

func (w *HAButtonWidget) Update(ctx context.Context) error {
	w.LastUpdate = time.Now()
	return nil
}


func (w *HASwitchWidget) Trigger() error {
	haClient := w.HABaseWidget.haProvider.GetHAClient()
	if haClient == nil || !haClient.IsConnected() {
		return fmt.Errorf("Home Assistant client not connected")
	}

	serviceData := map[string]interface{}{
		"entity_id": w.EntityID,
	}

	return haClient.CallService("switch", "toggle", serviceData)
}

func (w *HALightWidget) Trigger() error {
	haClient := w.HABaseWidget.haProvider.GetHAClient()
	if haClient == nil || !haClient.IsConnected() {
		return fmt.Errorf("Home Assistant client not connected")
	}

	serviceData := map[string]interface{}{
		"entity_id": w.EntityID,
	}

	return haClient.CallService("light", "toggle", serviceData)
}

func (w *HALightWidget) SetBrightness(brightness int) error {
	haClient := w.HABaseWidget.haProvider.GetHAClient()
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
	haClient := w.HABaseWidget.haProvider.GetHAClient()
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

