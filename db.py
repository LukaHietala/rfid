import sqlite3
from datetime import datetime

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

def create_student(rfid_id, name, start_date, end_date, start_time, end_time, weekmask="1111100"):
    timestamp = datetime.now().isoformat(timespec="seconds")

    con = sqlite3.connect(DB_NAME)
    con.row_factory = sqlite3.Row
    cur = con.cursor()

    # Date format is: '%Y-%m-%d %H:%M:%S' - 2026-08-20 12:52:36
    # Time format is: '%H.%M' - 8.00
    # Weekmask: '1111100' - 1 is work day, 0 is free day
    cur.execute("""
        INSERT INTO students (rfid_id, name, created_at, start_date, end_date, start_time, end_time, weekmask)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        RETURNING *
    """, (str(rfid_id), name, timestamp, start_date, end_date, start_time, end_time, weekmask))

    student = cur.fetchone()

    con.commit()
    con.close()

    return dict_from_row(student)

def update_student():
    pass

def toggle_student_status(id):
    con = sqlite3.connect(DB_NAME)
    cur = con.cursor()

    cur.execute("""
        SELECT status
        FROM students
        WHERE id = ?
    """, (str(id),))

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
        WHERE id = ?
    """, (status, str(id),))

    con.commit()
    con.close()

if __name__ == "__main__":
    init_db()
