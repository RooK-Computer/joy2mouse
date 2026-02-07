import argparse
import sys
import threading
import time

from evdev import InputDevice, ecodes, UInput


AXIS_X = ecodes.ABS_RX
AXIS_Y = ecodes.ABS_RY

BTN_LEFT = ecodes.BTN_LEFT
BTN_RIGHT = ecodes.BTN_RIGHT

# Global variables for the position
last_x = 0
last_y = 0


def debug_print(debug=False, msg=''):
    """
    Print debug messages flag is set.
    """
    if debug:
        print(msg)


def commandline(args):

    parser = argparse.ArgumentParser(prog='joy2mouse.py')

    parser.add_argument('--debug',
                        action='store_true', required=False,
                        help='Debug logging')

    parser.add_argument('-d', '--device',
                        type=str, required=False,
                        default='/dev/input/event0',
                        help='Device path')

    parser.add_argument('-s', '--scale',
                        type=int, required=False,
                        default=6000,
                        help='Scale to use for mouse sensitivity. Lower means slower movement')

    parser.add_argument('-i', '--interval',
                        type=int, required=False,
                        default=120,
                        help='Update interval for mouse movement per second')

    parser.add_argument('-L', '--left',
                        type=int, required=False,
                        default=ecodes.BTN_TL,
                        help='Input code for the left mouse click')

    parser.add_argument('-R', '--right',
                        type=int, required=False,
                        default=ecodes.BTN_TR,
                        help='Input code for the right mouse click')

    return parser.parse_args(args)


def movement_loop(ui, interval):
    global last_x, last_y
    dt = 1.0 / interval

    while True:
        if last_x != 0:
            ui.write(ecodes.EV_REL, ecodes.REL_X, last_x)
        if last_y != 0:
            ui.write(ecodes.EV_REL, ecodes.REL_Y, last_y)

        if last_x != 0 or last_y != 0:
            ui.syn()

        time.sleep(dt)


def main(args):
    global last_x, last_y

    btn_left_src = args.left
    btn_right_src = args.right

    dev = InputDevice(args.device)

    ui = UInput({
        ecodes.EV_REL: (ecodes.REL_X, ecodes.REL_Y),
        ecodes.EV_KEY: (BTN_LEFT, BTN_RIGHT),
    }, name="joy2mouse")

    print("joy2mouse Emulation running. CTRL+C to exit.")

    # Background thread
    thr = threading.Thread(target=movement_loop, args=(ui, args.interval,), daemon=True)
    thr.start()

    try:
        for event in dev.read_loop():
            # Axis
            if event.type == ecodes.EV_ABS:
                debug_print(args.debug, f"Axis event: {event.code}")
                if event.code == AXIS_X:
                    last_x = int(event.value / args.scale)
                    debug_print(args.debug, f"X: {last_x}, Y: {last_y}")

                if event.code == AXIS_Y:
                    # Note, to invert the Y axis negate this
                    last_y = int(event.value / args.scale)
                    debug_print(args.debug, f"X: {last_x}, Y: {last_y}")

            # Buttons
            if event.type == ecodes.EV_KEY:
                debug_print(args.debug, f"Button event: {event.code}")

                if event.code == btn_left_src:
                    ui.write(ecodes.EV_KEY, BTN_LEFT, event.value)
                    debug_print(args.debug, "Left click")
                    ui.syn()

                if event.code == btn_right_src:
                    ui.write(ecodes.EV_KEY, BTN_RIGHT, event.value)
                    debug_print(args.debug, "Right click")
                    ui.syn()

    except KeyboardInterrupt:
        print("\nExited.")
    finally:
        ui.close()


if __name__ == "__main__":
    ARGS = commandline(sys.argv[1:])
    main(ARGS)
