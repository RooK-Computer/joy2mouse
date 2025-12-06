import threading
import time
from evdev import InputDevice, ecodes, UInput

DEV_PATH = "/dev/input/event0"

AXIS_X = ecodes.ABS_RX
AXIS_Y = ecodes.ABS_RY
BTN_LEFT_SRC = ecodes.BTN_TL2
BTN_RIGHT_SRC = ecodes.BTN_TR2

BTN_LEFT = ecodes.BTN_LEFT
BTN_RIGHT = ecodes.BTN_RIGHT

SCALE = 6000      # Empfindlichkeit
UPDATE_HZ = 120   # Wie oft pro Sekunde die Maus bewegt wird

# Globale Achs-Werte
last_x = 0
last_y = 0


def movement_loop(ui):
    global last_x, last_y
    dt = 1.0 / UPDATE_HZ

    while True:
        # kontinuierliche Mausbewegung
        if last_x != 0:
            ui.write(ecodes.EV_REL, ecodes.REL_X, last_x)
        if last_y != 0:
            ui.write(ecodes.EV_REL, ecodes.REL_Y, last_y)

        if last_x != 0 or last_y != 0:
            ui.syn()

        time.sleep(dt)


def main():
    global last_x, last_y

    dev = InputDevice(DEV_PATH)

    ui = UInput({
        ecodes.EV_REL: (ecodes.REL_X, ecodes.REL_Y),
        ecodes.EV_KEY: (BTN_LEFT, BTN_RIGHT),
    }, name="joy2mouse_continuous")

    print("Maus-Emulation (kontinuierlich) läuft. STRG+C zum Beenden.")

    # Start Hintergrund-Thread
    thr = threading.Thread(target=movement_loop, args=(ui,), daemon=True)
    thr.start()

    try:
        for event in dev.read_loop():
            # Achsen
            if event.type == ecodes.EV_ABS:
                if event.code == AXIS_X:
                    last_x = int(event.value / SCALE)

                elif event.code == AXIS_Y:
                    # Y wurde von dir invertiert gewünscht – also kein Minus hier
                    last_y = int(event.value / SCALE)

            # Buttons
            elif event.type == ecodes.EV_KEY:
                if event.code == BTN_LEFT_SRC:
                    ui.write(ecodes.EV_KEY, BTN_LEFT, event.value)
                    ui.syn()
                elif event.code == BTN_RIGHT_SRC:
                    ui.write(ecodes.EV_KEY, BTN_RIGHT, event.value)
                    ui.syn()

    except KeyboardInterrupt:
        print("\nBeendet.")
    finally:
        ui.close()


if __name__ == "__main__":
    main()

