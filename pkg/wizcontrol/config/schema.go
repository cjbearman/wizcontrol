package config

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

//go:embed schema.json
var schema []byte

func loadAndValidate(filename string) ([]byte, error) {
	raw, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var body interface{}
	if err := yaml.Unmarshal(raw, &body); err != nil {
		return nil, err
	}

	schemaLoader := gojsonschema.NewBytesLoader(schema)
	documentLoader := gojsonschema.NewGoLoader(body)
	results, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return nil, err
	}
	if !results.Valid() {
		return nil, fmt.Errorf("configvalidation failed: %v", results.Errors())
	}

	return raw, nil
}
