package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/holoplot/go-evdev"
)

type EvdevMap map[evdev.EvCode]evdev.EvCode

func (m *EvdevMap) GetInputCodes() []evdev.EvCode {
	var codes = make([]evdev.EvCode, 0, len(*m))

	for in := range *m {
		codes = append(codes, in)
	}

	return codes
}

func (m *EvdevMap) GetOutputCodes() []evdev.EvCode {
	var codes = make([]evdev.EvCode, 0, len(*m))

	for _, out := range *m {
		codes = append(codes, out)
	}

	return codes
}

// UnmarshalJSON implements custom unmarshaling for IntMap
func (m *EvdevMap) UnmarshalJSON(data []byte) error {
	var stringMap map[string]string

	errUnmarsh := json.Unmarshal(data, &stringMap)

	if errUnmarsh != nil {
		return fmt.Errorf("failed to load key mapping: %w", errUnmarsh)
	}

	*m = make(map[evdev.EvCode]evdev.EvCode, len(stringMap))

	for inputKey, outputKey := range stringMap {
		inputCode, okInput := evdev.KEYFromString[inputKey]

		if !okInput {
			fmt.Printf("Warning: Unknown input key %s\n", outputKey)
		}

		outputCode, okOutput := evdev.KEYFromString[outputKey]

		if !okOutput {
			fmt.Printf("Warning: Unknown output key %s\n", outputKey)
		}

		(*m)[inputCode] = outputCode
	}

	return nil
}

type Config struct {
	InputPath   string   `json:"input"`
	Rate        int      `json:"rate"`
	Deadzone    int32    `json:"deadzone"`
	Sensitivity int32    `json:"sensitivity"`
	InvertY     bool     `json:"invert_y"`
	InvertX     bool     `json:"invert_x"`
	KeyMapping  EvdevMap `json:"key_mapping"`
}

// DefaultConfig returns a Config with default values
func defaultConfig() *Config {
	return &Config{
		InputPath:   "",
		Rate:        60,
		Deadzone:    1000,
		Sensitivity: 4000,
		InvertY:     false,
		InvertX:     false,
		KeyMapping: EvdevMap{
			evdev.BTN_TL: evdev.BTN_LEFT,
			evdev.BTN_TR: evdev.BTN_RIGHT,
		},
	}
}

// LoadConfig reads and parses the JSON config file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)

	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := defaultConfig()

	errUnmarsh := json.Unmarshal(data, config)

	if errUnmarsh != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", errUnmarsh)
	}

	return config, nil
}
