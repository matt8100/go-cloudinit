package cloudinit

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type configAlias Config

// MarshalYAML implements yaml.Marshaler.
func (c Config) MarshalYAML() (any, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return configAlias(c), nil
}

// Render validates the config and returns a complete #cloud-config document.
func (c Config) Render() ([]byte, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	out, err := safeYAMLMarshal(configAlias(c))
	if err != nil {
		return nil, err
	}
	return []byte(Header + string(out)), nil
}

// String renders the config and returns it as a string.
func (c Config) String() (string, error) {
	out, err := c.Render()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// TrimmedYAML returns YAML text with surrounding whitespace removed.
func TrimmedYAML(data []byte) string {
	return strings.TrimSpace(string(data))
}

func safeYAMLMarshal(value any) (out []byte, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("yaml marshal panic: %v", recovered)
		}
	}()
	return yaml.Marshal(value)
}
