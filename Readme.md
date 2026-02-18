# joy2mouse

This tool emulates a mouse by converting gamepad events.

Note, it is only tested with USB connected 8BitDo gamepads, that don't require any configuration keypresses.

## Usage

```bash
# Show available devices
joy2mouse -list
```

```bash
# Start emulation with the given configuration
joy2mouse -config myconfig.json
```

## Configuration

Example configuration:

```json
{
    "input": "/dev/input/event17",
    "rate": 60,
    "sensitivity": 4000,
    "deadzone": 100,
    "invert_y": false,
    "invert_x": false,
    "key_mapping": {
        "BTN_TL": "BTN_LEFT",
        "BTN_TR": "BTN_RIGHT"
    }
}
```

Options:

| Key      | Description  |
| -------- | ----------- |
| `input` | Path to the Linux device (e.g., /dev/input/event0) |
| `rate` | Update interval for mouse movement per second |
| `sensitivity` | Scale to use for mouse sensitivity. Lower means slower movement |
| `deadzone` | Deadzone threshold for stopping movement |
| `invert_x` | Invert the X-axis |
| `invert_y` | Invert the Y-axis |
| `key_mapping` | List of input to output keys |

Note, when no input device is configured a list of available devices is shown to select from.

# Development

Packages are created with [nfpm](https://nfpm.goreleaser.com/)

```
VERSION=1.0.0 make package
```
