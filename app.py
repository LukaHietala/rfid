import threading

from flask import Flask
from flask import render_template
from flask_socketio import SocketIO, emit
from apscheduler.schedulers.background import BackgroundScheduler

from db import get_scans, get_students, create_student, accumulate_done
from rfid import start_logger

app = Flask(__name__)
app.config['SECRET_KEY'] = 'kiljuva_pomeranian'
socketio = SocketIO(app)

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
    return render_template('admin.html', students=get_students())

@socketio.on('add_student')
def add_student(json):
    socketio.emit('new_student', create_student(json["rfid_id"], json["name"], "13.08.2026 12:52:36", "20.08.2026 12:52:36", "8.00", "16.00"))

@socketio.on('edit_student')
def edit_student(json):
    
    print(json)
    
if __name__ == "__main__":
    # Listen to RFID reader
    thread = threading.Thread(
        target=start_logger,
        args=(new_scan,), # :DD
        daemon=False
    )
    thread.start()

    # Accumulate done hours for students
    interval = 10
    scheduler.add_job(lambda: accumulate_done(interval), 'interval', seconds=interval)
    scheduler.start()

    # Serve the webapp
    socketio.run(app, host="0.0.0.0")

    scheduler.shutdown()
