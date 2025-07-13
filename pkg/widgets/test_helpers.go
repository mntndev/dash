package widgets

import (
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

// Helper function to convert map to ast.Node for tests.
func configToNode(config map[string]interface{}) ast.Node {
	if config == nil {
		return nil
	}
	node, err := yaml.ValueToNode(config)
	if err != nil {
		panic(err) // Should not happen in tests
	}
	return node
}