package widgets

import (
	"context"
	"fmt"
	"time"

	"github.com/mntndev/dash/pkg/integrations"
)

type HAEntityWidget struct {
	*BaseWidget
	EntityID string
	HAClient *integrations.HomeAssistantClient
}

type HAButtonWidget struct {
	*BaseWidget
	EntityID string
	Service  string
	Domain   string
	HAClient *integrations.HomeAssistantClient
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

func CreateHAEntityWidget(config map[string]interface{}) (Widget, error) {
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
	}

	return widget, nil
}

func CreateHAButtonWidget(config map[string]interface{}) (Widget, error) {
	entityID, ok := config["entity_id"].(string)
	if !ok {
		return nil, fmt.Errorf("entity_id is required for home_assistant.button widget")
	}

	service, ok := config["service"].(string)
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
		Service:  service,
		Domain:   domain,
	}

	widget.Data = &HAButtonData{
		EntityID: entityID,
		Service:  service,
		Domain:   domain,
		Label:    getStringConfig(config, "label", "Button"),
	}

	return widget, nil
}

func (w *HAEntityWidget) Update(ctx context.Context) error {
	if w.HAClient == nil || !w.HAClient.IsConnected() {
		return fmt.Errorf("Home Assistant client not connected")
	}

	states, err := w.HAClient.GetStates()
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

func (w *HAEntityWidget) SetHAClient(client *integrations.HomeAssistantClient) {
	w.HAClient = client
}

func (w *HAButtonWidget) SetHAClient(client *integrations.HomeAssistantClient) {
	w.HAClient = client
}

func (w *HAButtonWidget) Trigger() error {
	if w.HAClient == nil || !w.HAClient.IsConnected() {
		return fmt.Errorf("Home Assistant client not connected")
	}

	serviceData := map[string]interface{}{
		"entity_id": w.EntityID,
	}

	return w.HAClient.CallService(w.Domain, w.Service, serviceData)
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