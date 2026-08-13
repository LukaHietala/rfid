import sqlite3
from datetime import datetime, timedelta, date, time
from dateutil.relativedelta import relativedelta
import numpy as np

DB_NAME = "rfid.db"

# Source - https://stackoverflow.com/a/9538363
# Posted by bbengfort, modified by community. See post 'Timeline' for change history
# Retrieved 2026-08-13, License - CC BY-SA 3.0

def dict_from_row(row):
    if not row:
        return
    return dict(zip(row.keys(), row))       

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

    cur.execute("""
        CREATE TABLE IF NOT EXISTS students (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            rfid_id TEXT NOT NULL,
            name TEXT NOT NULL,
            status TEXT CHECK( status IN ('IN','OUT') )   NOT NULL DEFAULT 'OUT',
            created_at TEXT NOT NULL,
            start_date TEXT NOT NULL,
            end_date TEXT NOT NULL,
            start_time TEXT NOT NULL,
            end_time TEXT NOT NULL,
            done_seconds INTEGER NOT NULL DEFAULT 0,
            weekmask TEXT NOT NULL
        )
    """)



    # millä aikavälillä? 1.8.2026 - 30.8.2026
    # kauan per päivä.
    # kauan vielä pitää tehdä töitä, joka työpäivä lisätään esim 7 tuntia
    # jonka pitää kiriä

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

def get_scans(limit=1000):
    con = sqlite3.connect(DB_NAME)
    con.row_factory = sqlite3.Row
    cur = con.cursor()

    cur.execute("""
        SELECT *
        FROM scans
        LIMIT ?
    """, (limit,))

    rows = cur.fetchall()
    con.close()

    return [dict(row) for row in rows]

def get_students(limit=1000):
    con = sqlite3.connect(DB_NAME)
    con.row_factory = sqlite3.Row
    cur = con.cursor()

    cur.execute("""
        SELECT *
        FROM students
        LIMIT ?
    """, (limit,))

    rows = cur.fetchall()
    con.close()

    return [dict(row) for row in rows]

def get_student(rfid_id):
    con = sqlite3.connect(DB_NAME)
    con.row_factory = sqlite3.Row
    cur = con.cursor()
    
    cur.execute("""
        SELECT *
        FROM students
        WHERE rfid_id = ?
    """, (str(rfid_id),))

    student = cur.fetchone()

    con.close()
    return dict_from_row(student)

def get_student_remaining(student):
    done_seconds = student["done_seconds"]
    day_starts_str = student["start_time"]
    day_ends_str = student["end_time"]
    start_date_str = student["start_date"] 
    end_date_str = student["end_date"]
    weekmask = student["weekmask"]

    done_time = timedelta(seconds=done_seconds)
    day_starts_time = datetime.strptime(day_starts_str, "%H.%M").time()
    day_ends_time = datetime.strptime(day_ends_str, "%H.%M").time()

    day_length = datetime.combine(date.today(), day_ends_time) - datetime.combine(date.today(), day_starts_time)

    date_format = '%Y-%m-%d %H:%M:%S'

    start_date = datetime.strptime(start_date_str, date_format)
    end_date = datetime.strptime(end_date_str, date_format)

    current_date = datetime.now()

    business_days = np.busday_count(
        np.datetime64(start_date.date(), "D"),
        np.datetime64(min(current_date.date(), end_date.date()), "D") + np.timedelta64(1, "D"),
        weekmask=weekmask
    )

    return (((datetime.combine(date.today(), current_date.time()) - datetime.combine(date.today(), day_starts_time))) + (day_length * (business_days - 1)) - done_time).total_seconds()


def toggle_student_status(rfid_id):
    con = sqlite3.connect(DB_NAME)
    cur = con.cursor()

    cur.execute("""
        SELECT status
        FROM students
        WHERE rfid_id = ?
    """, (str(rfid_id),))

    res = cur.fetchone()

    if not res:
        con.commit()
        con.close()
        return
    
    status = res[0]
    
    if status == "IN":
        status = "OUT"
    else:
        status = "IN"
    
    cur.execute("""
        UPDATE students
    	SET status = ?
        WHERE rfid_id = ?
    """, (status, str(rfid_id),))

    con.commit()
    con.close()

if __name__ == "__main__":
    init_db()
