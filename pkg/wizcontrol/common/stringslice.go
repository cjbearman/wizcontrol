package common

import "go.yaml.in/yaml/v3"

// StringSlice can be used in YAML contexts to unmarshal either a single string
// or a slice of strings
type StringSlice []string

func (s *StringSlice) UnmarshalYAML(value *yaml.Node) error {
	var multi []string
	if err := value.Decode(&multi); err == nil {
		*s = multi
		return nil
	}

	var single string
	if err := value.Decode(&single); err != nil {
		return err
	}

	*s = []string{single}
	return nil
}
