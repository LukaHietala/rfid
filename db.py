from calendar import week
import sqlite3
from datetime import datetime, timedelta, date, time
from dateutil.relativedelta import relativedelta
import json

from utils import format_datetime, to_datetime, to_time

DB_NAME = "rfid.db"

# Source - https://stackoverflow.com/a/9538363
# Posted by bbengfort, modified by community. See post 'Timeline' for change history
# Retrieved 2026-08-13, License - CC BY-SA 3.0

def dict_from_row(row):
    if not row:
        return
    return dict(zip(row.keys(), row))       

# TODO: DB mutex
# Accumulate done runs on a seperate thread so it might corrupt the db

def get_connection():
    return sqlite3.connect(DB_NAME, timeout=5)

def init_db():
    con = get_connection()
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
            rfid_id TEXT NOT NULL UNIQUE,
            name TEXT NOT NULL,
            status TEXT CHECK( status IN ('IN','OUT') )   NOT NULL DEFAULT 'OUT',
            created_at TEXT NOT NULL,
            start_date TEXT NOT NULL,
            end_date TEXT NOT NULL,
            schedule TEXT NOT NULL,
            done_seconds INTEGER NOT NULL DEFAULT 0,
            excluded_days TEXT DEFAULT '[]'
        ) 
    """)

    cur.execute("""
        CREATE TABLE IF NOT EXISTS breaks (
            time INTEGER NOT NULL DEFAULT 0
        )
    """)

    con.execute('pragma journal_mode=wal')

    con.commit()
    con.close()

def log_scan(rfid_id, name):
    timestamp = format_datetime(datetime.now())

    con = get_connection()
    cur = con.cursor()

    cur.execute("""
        INSERT INTO scans (rfid_id, name, timestamp)
        VALUES (?, ?, ?)
    """, (str(rfid_id), name, timestamp))

    con.commit()
    con.close()

    return timestamp

def get_scans(limit=1000):
    """ Returns all scans """
    con = get_connection()
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
    """ Returns all students. """
    con = get_connection()
    con.row_factory = sqlite3.Row
    cur = con.cursor()

    cur.execute("""
        SELECT *
        FROM students
        LIMIT ?
    """, (limit,))

    rows = cur.fetchall()

    students = []
    for row in rows:
        student = dict_from_row(row)
        remaining = get_student_remaining(student)
        student["remaining"] = remaining
        students.append(student)

    con.close()

    return students

def get_student(rfid_id):
    """ Gets student based on rfid_id and retunrns it """
    con = get_connection()
    con.row_factory = sqlite3.Row
    cur = con.cursor()
    
    cur.execute("""
        SELECT *
        FROM students
        WHERE rfid_id = ?
    """, (str(rfid_id),))

    student = dict_from_row(cur.fetchone())
    if student:
        remaining = get_student_remaining(student)
        student["remaining"] = remaining

    con.close()

    return student

def create_student(rfid_id, name, start_date, end_date, schedule, excluded_days="[]"):
    """ Creates a student and returns it """
    timestamp = format_datetime(datetime.now())

    con = get_connection()
    con.row_factory = sqlite3.Row
    cur = con.cursor()

    # Date format is: '%d.%m.%Y - 20.08.2026
    cur.execute("""
        INSERT INTO students (rfid_id, name, created_at, start_date, end_date, schedule, excluded_days)
        VALUES (?, ?, ?, ?, ?, ?, ?)
        RETURNING *
    """, (str(rfid_id), name, timestamp, start_date, end_date, schedule, excluded_days))

    student = dict_from_row(cur.fetchone())
    if student:
        remaining = get_student_remaining(student)
        student["remaining"] = remaining

    con.commit()
    con.close()

    return student

def update_student(student_id, name, status, start_date, end_date, schedule, done_seconds, excluded_days="[]"):
    """ Updates a student and returns it. """
    con = get_connection()
    con.row_factory = sqlite3.Row
    cur = con.cursor()
    
    if not isinstance(done_seconds, int):
        done_seconds = max(0, done_seconds)

    # Date format is: '%d.%m.%Y' - 20.08.2026
    cur.execute("""
        UPDATE students 
        SET (name, status, start_date, end_date, schedule, done_seconds, excluded_days) 
        = (?, ?, ?, ?, ?, ?, ?)
        WHERE id = ?
        RETURNING *
    """, (name, status, start_date, end_date, schedule, done_seconds, excluded_days, student_id,))

    student = dict_from_row(cur.fetchone())
    if student:
        remaining = get_student_remaining(student)
        student["remaining"] = remaining

    con.commit()
    con.close()

    return student

def remove_student(id):
    """
    Deletes a student from the database by id.
    Returns deleted student
    """
    con = get_connection()
    con.row_factory = sqlite3.Row
    cur = con.cursor()

    cur.execute("""
        DELETE FROM students
        WHERE id = ?
        RETURNING *
    """, (id,))


    student = dict_from_row(cur.fetchone())
    if student:
        remaining = get_student_remaining(student)
        student["remaining"] = remaining

    con.commit()
    con.close()

    return student

def toggle_student_status(id):
    """
    Toggles student status between "IN" and "OUT"
    Return updated status
    """
    con = get_connection()
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

    return status


def accumulate_done(interval):
    """
    Accumulates done_seconds every interval.
    Returns updated students
    """
    con = get_connection()
    con.row_factory = sqlite3.Row
    cur = con.cursor()

    # TODO: Limit to workday and add overtime somehow

    cur.execute("""
        UPDATE students 
        SET done_seconds = done_seconds + ?
        WHERE status = "IN"
        RETURNING *
    """, (interval,))

    rows = cur.fetchall()

    con.commit()
    con.close()

    return [dict(row) for row in rows]

def replace_workdays(start_date, end_date):
    """    
    Replace the starting and ending dates of all students.    
    Returns updated students table
    """
    con = get_connection()
    con.row_factory = sqlite3.Row
    cur = con.cursor()

    cur.execute("""
        UPDATE students 
        SET start_date = ?, end_date = ?
        RETURNING *
    """, (start_date, end_date,))

    rows = cur.fetchall()

    con.commit()
    con.close()

def add_excluded_days(excluded_days : list[str]):
    """
    Adds excluded_days to all students, without overwriting
    existing days. Returns updated students table.
    """

    con = get_connection()
    cur = con.cursor()

    students = get_students()
    for student in students:
        new_excluded = json.loads(student["excluded_days"])
        new_excluded += excluded_days
        new_excluded = list(set(new_excluded))
        
        cur.execute("""
        UPDATE students 
        SET excluded_days = ?
        WHERE id=?
        """, (json.dumps(new_excluded), student["id"],))

    con.commit()
    con.close()
        
    return get_students()

def get_break_time():
    """ Gets break time in seconds """
    con = get_connection()
    cur = con.cursor()
    
    cur.execute("""
        SELECT time
        FROM breaks
    """)

    time = cur.fetchone()[0]

    con.close()

    return time

def set_break_time(seconds):
    """ Sets break time in seconds """
    con = get_connection()
    cur = con.cursor()

    cur.execute("""
        UPDATE breaks
        SET time = ?
        RETURNING *
    """, (seconds,))

    row = cur.fetchone()[0]

    con.commit()
    con.close()

    return row

def get_student_remaining(student):    
    """ Returns student's remaining time for working in seconds """

    # Dates for when the work starts and ends
    today = datetime.today()
    start = to_datetime(student["start_date"])
    end = to_datetime(student["end_date"])

    # Break time in a day in seconds
    break_per_day = timedelta(seconds=get_break_time())
    # List of days to exclude in calculations, such as holidays
    # json: ["2026-08-20"]
    excluded_days = json.loads(student["excluded_days"])
    # Students schedule, null is free day
    # json: [{"start": "08:00", "end": "16:00"}, null...] (length: 7)
    schedule = json.loads(student["schedule"])
    # Done work in seconds
    done_time = timedelta(seconds=int(student["done_seconds"]))
    
    # Accumulate required work time
    # Between start and today, days after end are not included
    work = timedelta(0)
    while start.date() <= today.date():
        # Stop when end date is reached
        if start > end:
            break
        # Skip weekends and excluded_days
        if schedule[start.weekday()] and not str(start.date()) in excluded_days:
            # Accumulate required work
            hours: dict[str, str] = schedule[start.weekday()]
            work += (to_time(hours["end"]) - to_time(hours["start"])) - break_per_day 
        start += timedelta(days=1)
    
    return (work - done_time).total_seconds()

if __name__ == "__main__":
    init_db()
