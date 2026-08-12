import sqlite3
from datetime import datetime

DB_NAME = "rfid.db"

def init_db():
    con = sqlite3.connect(DB_NAME)
    cur = con.cursor()

    cur.execute("""
        CREATE TABLE IF NOT EXISTS scans (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            rfid_id TEXT NOT NULL,
            name TEXT NOT NULL,
            timestamp TEXT NOT NULL
        )
    """)

    con.commit()
    con.close()

def log_scan(rfid_id, name):
    timestamp = datetime.now().isoformat(timespec="seconds")

    con = sqlite3.connect(DB_NAME)
    cur = con.cursor()

    cur.execute("""
        INSERT INTO scans (rfid_id, name, timestamp)
        VALUES (?, ?, ?)
    """, (str(rfid_id), name, timestamp))

    con.commit()
    con.close()

    return timestamp

if __name__ == "__main__":
    init_db()
