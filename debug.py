import RPi.GPIO as GPIO
from mfrc522 import SimpleMFRC522
from rfid import write_data

reader = SimpleMFRC522()

data = input("mita kirjoittaa")
write_data(reader, data)


