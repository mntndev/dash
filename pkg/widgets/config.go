package widgets

import (
	"fmt"
	"reflect"
)

// ClockConfig defines configuration for clock widgets.
type ClockConfig struct {
	Format string `json:"format" default:"15:04:05"`
}

type HAEntityConfig struct {
	EntityID string `json:"entity_id" validate:"required"`
}

type HAButtonConfig struct {
	EntityID string `json:"entity_id" validate:"required"`
	Service  string `json:"service" validate:"required"`
	Domain   string `json:"domain" validate:"required"`
	Label    string `json:"label" default:"Button"`
}

type SplitConfig struct {
	Sizes []float64 `json:"sizes"`
}

type DexcomConfig struct {
	// No specific configuration for now
}

type GrowConfig struct {
	// No specific configuration
}

// ConfigParser provides utilities for parsing widget configurations.
type ConfigParser struct{}

func NewConfigParser() *ConfigParser {
	return &ConfigParser{}
}

// ParseConfig converts a generic map to a typed configuration struct.
func (cp *ConfigParser) ParseConfig(config map[string]interface{}, target interface{}) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	targetValue = targetValue.Elem()
	targetType := targetValue.Type()

	for i := 0; i < targetValue.NumField(); i++ {
		field := targetValue.Field(i)
		fieldType := targetType.Field(i)

		jsonTag := fieldType.Tag.Get("json")
		defaultTag := fieldType.Tag.Get("default")
		validateTag := fieldType.Tag.Get("validate")

		if jsonTag == "" {
			continue
		}

		// Get value from config map
		value, exists := config[jsonTag]

		// Handle required validation
		if validateTag == "required" && (!exists || value == nil) {
			return fmt.Errorf("required field '%s' is missing", jsonTag)
		}

		// Use default value if not provided
		if !exists || value == nil {
			if defaultTag != "" {
				value = defaultTag
			} else {
				continue
			}
		}

		// Set the field value
		if err := cp.setFieldValue(field, value); err != nil {
			return fmt.Errorf("error setting field '%s': %w", jsonTag, err)
		}
	}

	return nil
}

func (cp *ConfigParser) setFieldValue(field reflect.Value, value interface{}) error {
	if !field.CanSet() {
		return fmt.Errorf("field cannot be set")
	}

	switch field.Kind() {
	case reflect.String:
		if str, ok := value.(string); ok {
			field.SetString(str)
		} else {
			return fmt.Errorf("expected string, got %T", value)
		}
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.Float64 {
			if slice, ok := value.([]interface{}); ok {
				floatSlice := make([]float64, len(slice))
				for i, v := range slice {
					if f, ok := v.(float64); ok {
						floatSlice[i] = f
					} else {
						return fmt.Errorf("expected float64 in slice, got %T", v)
					}
				}
				field.Set(reflect.ValueOf(floatSlice))
			} else {
				return fmt.Errorf("expected []interface{} for float64 slice, got %T", value)
			}
		}
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}

	return nil
}
