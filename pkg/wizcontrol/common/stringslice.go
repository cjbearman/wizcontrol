package common

import (
	"bytes"
	"encoding/json"
	"fmt"

	"go.yaml.in/yaml/v3"
)

// StringSlice can be used in YAML contexts to unmarshal either a single string
// or a slice of strings
type StringSlice []string

func (s *StringSlice) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var single string
		if err := value.Decode(&single); err != nil {
			return err
		}
		*s = []string{single}
		return nil

	case yaml.SequenceNode:
		var multi []string
		if err := value.Decode(&multi); err != nil {
			return err
		}
		*s = multi
		return nil

	default:
		return fmt.Errorf("line %d: cannot unmarshal %s into StringSlice", value.Line, value.ShortTag())
	}
}

func (s *StringSlice) UnmarshalJSON(data []byte) error {
	// Trim leading whitespace to check the first character
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil
	}

	// If it starts with '[', unmarshal as a slice of strings
	if trimmed[0] == '[' {
		var slice []string
		if err := json.Unmarshal(trimmed, &slice); err != nil {
			return err
		}
		*s = slice
		return nil
	}

	// Otherwise, unmarshal as a single string and wrap it in a slice
	var single string
	if err := json.Unmarshal(trimmed, &single); err != nil {
		return err
	}
	*s = []string{single}
	return nil
}
