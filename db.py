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

    cur.execute("""
        CREATE TABLE IF NOT EXISTS students (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            rfid_id TEXT NOT NULL,
            name TEXT NOT NULL,
            status TEXT CHECK( status IN ('IN','OUT') )   NOT NULL DEFAULT 'OUT',
            created_at TEXT NOT NULL
        )
    """)



    # monta päivää pitää olla töissä
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
    cur = con.cursor()
    
    cur.execute("""
        SELECT *
        FROM students
        WHERE rfid_id = ?
    """, (str(rfid_id),))

    student = cur.fetchone()

    con.close()
    return student

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
