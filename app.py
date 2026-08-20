import threading
import json
import bcrypt

from flask import Flask
from flask import render_template, jsonify, session, redirect, request
from flask_socketio import SocketIO, emit
from apscheduler.schedulers.background import BackgroundScheduler

from db import get_scans, get_students, create_student, accumulate_done, update_student, replace_workdays, add_excluded_days, remove_student
from rfid import start_logger

app = Flask(__name__)
app.config['SECRET_KEY'] = 'kiljuva_pomeranian'
socketio = SocketIO(app, cors_allowed_origins="*")

scheduler = BackgroundScheduler()

def new_scan(scan):
    """
    Sends new scan via websocket

    :scan: dict with name and rfid_id
    """
    socketio.emit('new_scan', scan)

@app.route("/")
def index():
    return render_template('index.html', students=get_students())

@app.route("/admin")
def admin():
    if not session.get("admin"):
        return redirect("/login")
    return render_template('admin.html', students=get_students())

@app.route("/login", methods=["GET", "POST"])
def login():
    if request.method == "POST":
        passwordBytes = (request.form.get("password")).encode("utf-8")
        if bcrypt.checkpw(passwordBytes, b'$2b$12$L7L7.Omjx5.xNck4cTzb6uKxIQqmhdgcegAzCX3Y9d2RqwwOwti9K'):
            session["admin"] = True
            return redirect("/admin")
        return render_template("login.html", wrong_pass=True)
    return render_template("login.html", wrong_pass=False)

@app.route("/logout")
def logout():
    session.pop("admin", None)
    return redirect("/login")

@socketio.on('add_student')
def add_student(json):
    socketio.emit('new_student', create_student(json["rfid_id"], json["name"], json["start_date"], json["end_date"], json["start_time"], json["end_time"],
                  excluded_days = json["excluded_days"]))
    
@socketio.on('delete_student')
def delete_student(json):
    socketio.emit('remove_student', remove_student(json["id"]))

@socketio.on('edit_student')
def edit_student(json):
    updated_student = update_student(json["id"],  json["name"], json["start_date"],
                         json["end_date"], json["start_time"], json["end_time"],
                         json["done_seconds"], json["weekmask"], json["excluded_days"])
    socketio.emit('update_student', updated_student)
    
@socketio.on('set_workdays')
def set_workdays(json):
    replace_workdays(json["start"], json["end"])
    socketio.emit('update_all', get_students())

@socketio.on('add_holidays')
def add_holidays(json):
    add_excluded_days(json["holidays"])
    socketio.emit('update_all', get_students())

@app.route("/api/students", methods=["GET"])
def get_students_json():
    return jsonify(get_students()), 200

if __name__ == "__main__":
    # Listen to RFID reader
    thread = threading.Thread(
        target=start_logger,
        args=(new_scan,), # :DD
        daemon=False
    )
    thread.start()

    # Accumulate done seconds for students
    interval = 10
    scheduler.add_job(lambda: accumulate_done(interval), 'interval', seconds=interval, max_instances=1)
    scheduler.start()

    # Serve the webapp
    socketio.run(app, host="0.0.0.0")

    scheduler.shutdown()
