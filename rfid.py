import time
import random

import RPi.GPIO as GPIO
from mfrc522 import SimpleMFRC522

from db import log_scan, get_student, toggle_student_status

def read_data(reader):
    id, data = reader.read()
    return id, data.strip()

def write_data(reader, data):
    id, written = reader.write(data)
    return id, written

def start_logger(new_scan):
    reader = SimpleMFRC522()

    last_tag = {
        "id": None,
        "last_scanned": 0
    }

    try:
        while True:
            id, data = read_data(reader)

            now = time.time()
            if (id == last_tag["id"] and (now - last_tag["last_scanned"]) < 2):
                continue

            last_tag["id"] = id
            last_tag["last_scanned"] = now

            if not data or not id:
                continue

            log_scan(id, data)
            new_scan({"name" : data, "rfid_id" : id})
            toggle_student_status(id)
            student = get_student(id)
            if student:
                print("Found student:", student)
            else:
                print("Student not found based on tag")

            time.sleep(0.5)
    except KeyboardInterrupt:
        pass
    except Exception as e:
        print(str(e))
    finally:
        GPIO.cleanup()

if __name__ == "__main__":
    start_logger()
