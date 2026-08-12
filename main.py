import time
import RPi.GPIO as GPIO
from mfrc522 import SimpleMFRC522
import random

from rfid import read_data, write_data
from db import log_scan

def main():
    reader = SimpleMFRC522()

    last_tag = {
        "id": None,
        "last_scanned": 0
    }

    print("Hold a tag near the reader")

    try:
        while True:
            id, data = read_data(reader)

            now = time.time()
            if (id == last_tag["id"] and (now - last_tag["last_scanned"]) < 5):
                continue

            last_tag["id"] = id
            last_tag["last_scanned"] = now

            if not data or not id:
                continue

            log_scan(id, data)

            print("ID: %s\nText: %s" % (id,data))
            time.sleep(0.5)
    except KeyboardInterrupt:
        pass
    except Exception as e:
        print(str(e))
    finally:
        GPIO.cleanup()


if __name__ == "__main__":
    main()
