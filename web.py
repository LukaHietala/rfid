from flask import Flask
from flask import render_template
from flask_socketio import SocketIO

from db import get_scans

app = Flask(__name__)
app.config['SECRET_KEY'] = 'kiljuva_pomeranian'
socketio = SocketIO(app)

@app.route("/")
def index():
    scans = get_scans()

    return render_template('index.html', scans=scans)

@socketio.on('test')
def test(json):
    print('asfdsadf ' + str(json))

if __name__ == "__main__":
    socketio.run(app)
