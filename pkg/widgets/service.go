package widgets

import "github.com/mntndev/dash/pkg/integrations"

// ServiceProvider provides access to integration clients
// This interface breaks circular dependencies between widgets and dashboard service
type ServiceProvider interface {
	GetHAClient() *integrations.HomeAssistantClient
	GetDexcomClient() *integrations.DexcomClient
}