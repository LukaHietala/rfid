import threading

from flask import Flask
from flask import render_template
from flask_socketio import SocketIO, emit

from db import get_scans, get_students
from rfid import start_logger

app = Flask(__name__)
app.config['SECRET_KEY'] = 'kiljuva_pomeranian'
socketio = SocketIO(app)

def new_scan(scan):
    """
    Sends new scan

    :scan: dict with name and rfid_id
    """
    socketio.emit('new_scan', scan)

def update_status(student_id, status):
    """
    Sends the status of a student

    :student_id: id
    :status: boolean
    """
    pass

@app.route("/")
def index():
    scans = get_scans()
    students = get_students()

    return render_template('index.html', scans=scans, students=students)

@socketio.on('test')
def test(json):
    print('asfdsadf ' + str(json))

if __name__ == "__main__":
    thread = threading.Thread(
        target=start_logger,
        args=(new_scan,), # :DD
        daemon=False
    )

    thread.start()
    socketio.run(app, host="0.0.0.0")
