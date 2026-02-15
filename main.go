package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/holoplot/go-evdev"
)

const (
	unknownKey = "UNKNOWN"
)

type Mouse struct {
	InputDevice   *evdev.InputDevice
	Configuration *Config
	active        bool  // Track whether continuous movement is active
	velocityX     int32 // Current X velocity
	velocityY     int32 // Current Y velocity
}

func NewMouse(configuration *Config) (*Mouse, error) {
	outputs := configuration.KeyMapping.GetOutputCodes()

	dev, err := evdev.CreateDevice(
		"mouse-emulation",
		evdev.InputID{
			BusType: 0x03,
			Vendor:  0x4711,
			Product: 0x0816,
			Version: 1,
		},
		map[evdev.EvType][]evdev.EvCode{
			evdev.EV_KEY: outputs,
			evdev.EV_REL: {
				evdev.REL_X,
				evdev.REL_Y,
				evdev.REL_WHEEL,
				evdev.REL_HWHEEL,
			},
		},
	)

	m := &Mouse{
		Configuration: configuration,
		InputDevice:   dev,
	}

	return m, err
}

func (m *Mouse) handleEvent(inputEvent *evdev.InputEvent) {
	switch inputEvent.Type {
	case evdev.EV_KEY:
		m.handleKeyEvent(inputEvent)
	case evdev.EV_ABS:
		switch inputEvent.Code {
		case evdev.ABS_X:
			m.velocityX = inputEvent.Value
		case evdev.ABS_Y:
			m.velocityY = inputEvent.Value
		case evdev.ABS_RX:
			m.velocityX = inputEvent.Value
		case evdev.ABS_RY:
			m.velocityY = inputEvent.Value
		}

		if abs(m.velocityX) > m.Configuration.Deadzone || abs(m.velocityY) > m.Configuration.Deadzone {
			m.active = true
		} else {
			m.active = false
			m.velocityX = 0
			m.velocityY = 0
		}
	}
}

func (m *Mouse) handleKeyEvent(inputEvent *evdev.InputEvent) {
	outputCode, ok := m.Configuration.KeyMapping[inputEvent.Code]

	if !ok {
		// We only want to print this message on the button press
		if inputEvent.Value == 0 {
			return
		}

		keyName, okName := evdev.KEYToString[inputEvent.Code]

		if !okName {
			keyName = unknownKey
		}

		fmt.Printf("Info: Key '%d' (%s) not configured\n", inputEvent.Code, keyName)
		return
	}

	//nolint: errcheck
	m.InputDevice.WriteOne(&evdev.InputEvent{
		Type:  evdev.EV_KEY,
		Code:  outputCode,
		Value: inputEvent.Value,
	})

	//nolint: errcheck
	m.InputDevice.WriteOne(&evdev.InputEvent{
		Type:  evdev.EV_SYN,
		Code:  evdev.SYN_REPORT,
		Value: 0,
	})
}

func (m *Mouse) updateMovement() {
	if !m.active {
		return
	}

	moveX := m.velocityX / m.Configuration.Sensitivity

	if m.Configuration.InvertX {
		moveX *= -1
	}

	moveY := m.velocityY / m.Configuration.Sensitivity

	if m.Configuration.InvertY {
		moveY *= -1
	}

	if moveX != 0 {
		//nolint: errcheck
		m.InputDevice.WriteOne(&evdev.InputEvent{
			Type:  evdev.EV_REL,
			Code:  evdev.REL_X,
			Value: moveX,
		})
	}

	if moveY != 0 {
		//nolint: errcheck
		m.InputDevice.WriteOne(&evdev.InputEvent{
			Type:  evdev.EV_REL,
			Code:  evdev.REL_Y,
			Value: moveY,
		})
	}

	//nolint: errcheck
	m.InputDevice.WriteOne(&evdev.InputEvent{
		Type:  evdev.EV_SYN,
		Code:  evdev.SYN_REPORT,
		Value: 0,
	})
}

func getDevices() (map[int]evdev.InputPath, error) {
	devicePaths, err := evdev.ListDevicePaths()

	if err != nil {
		return nil, err
	}

	devices := make(map[int]evdev.InputPath, len(devicePaths))

	for idx, d := range devicePaths {
		devices[idx] = d
	}

	return devices, nil
}

func printDevices(devices map[int]evdev.InputPath) {
	keys := make([]int, 0, len(devices))

	for k := range devices {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	for idx := range keys {
		d := devices[idx]
		fmt.Printf("[%d] %s:\t%s\n", idx, d.Path, d.Name)
	}
}

func printDeviceSelection(devices map[int]evdev.InputPath) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Available devices:\n")

	printDevices(devices)

	fmt.Printf("Select the device event number [0-%d]:\n", len(devices)-1)

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	index, err := strconv.Atoi(input)

	if err != nil {
		return "", errors.New("invalid input")
	}

	dev, ok := devices[index]

	if !ok {
		return "", errors.New("no such device")
	}

	return dev.Path, nil
}

func main() {
	var (
		cliConfigPath  string
		cliListDevices bool
	)

	flag.StringVar(&cliConfigPath, "config", "./config.json", "Path to the configuration file")
	flag.BoolVar(&cliListDevices, "list-devices", false, "List available devices and exit")

	flag.Parse()

	devices, errDevs := getDevices()

	if errDevs != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not get devices. %v", errDevs)
		os.Exit(1)
	}

	if cliListDevices {
		printDevices(devices)
		os.Exit(0)
	}

	config, errConfig := LoadConfig(cliConfigPath)

	if errConfig != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not read configuration. %v", errConfig)
		os.Exit(1)
	}

	if config.InputPath == "" {
		path, errSelect := printDeviceSelection(devices)

		if errSelect != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid selection %v\n", errSelect)
			os.Exit(1)
		}

		config.InputPath = path
	}

	inputDevice, err := evdev.Open(config.InputPath)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot read %s: %v\n", os.Args[1], err)
		os.Exit(1)
	}

	mouse, err := NewMouse(config)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create virtual mouse: %v\n", err)
		os.Exit(1)
	}

	//nolint: errcheck
	defer mouse.InputDevice.Close()

	// Channel to handle shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	events := make(chan evdev.InputEvent, 64)

	// Reading events from the input
	go func() {
		for {
			ev, err := inputDevice.ReadOne()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Could not read event: %v\n", err)
				return
			}
			events <- *ev
		}
	}()

	// Create ticker for continuous movement updates
	updateInterval := time.Second / time.Duration(config.Rate)
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	// Process events via the channel
	for {
		select {
		case <-sigChan:
			fmt.Printf("\nShutting down...\n")
			return
		case ev := <-events:
			// Handle the incoming event (e.g. clicks or movement changes)
			mouse.handleEvent(&ev)

		case <-ticker.C:
			// Handle continuous mouse movement based on the rate
			mouse.updateMovement()
		}
	}
}

// abs returns the absolute value of an int32
// Small helper function for detecting the deadzone
func abs(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}
