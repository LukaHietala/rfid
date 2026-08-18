import time
import random

import RPi.GPIO as GPIO
from mfrc522 import SimpleMFRC522

from db import log_scan, get_student, toggle_student_status, get_student_remaining

def read_tag(reader):
    id, data = reader.read()
    return id, data.strip()

def write_data(reader, data):
    id, written = reader.write(data)
    return id, written

def start_logger(new_scan):
    reader = SimpleMFRC522()

    last_tag = {
        "rfid_id": None,
        "last_scanned": 0
    }

    try:
        while True:
            rfid_id, _ = read_tag(reader)

            now = time.time()
            if (rfid_id == last_tag["rfid_id"] and (now - last_tag["last_scanned"]) < 2):
                continue

            last_tag["rfid_id"] = rfid_id
            last_tag["last_scanned"] = now

            if not rfid_id:
                continue

            student = get_student(rfid_id)

            if student:
                new_status = toggle_student_status(student["id"])
                log_scan(student["rfid_id"], student["name"])
                new_scan({"found" : True , "name" : student["name"], "rfid_id" : student["rfid_id"], "status" : new_status})
                
                print("Found student:", student)
            else:
                new_scan({"found" : False, "rfid_id" : rfid_id, "msg" : "Oppilasta ei löytynyt" })
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
