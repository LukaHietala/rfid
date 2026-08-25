Initialize database:

```python
python db.py
```

Create `.venv`, install deps and run the app with:

```python
python app.py
```

If you want to run it presistently create a Systemd like this for example:

```ini
# /etc/systemd/system/rfid.service 
[Unit]
Description=rfid
After=network.target

[Service]
type=exec
WorkingDirectory=
ExecStart=
User=
Restart=on-failure
RestartSec=10
TimeoutStopSec=30
KillSignal=SIGTERM

[Install]
WantedBy=multi-user.target
```

Fill in the missing fields. Remember to run using python in `.venv`, usually `.venv/bin/python`. This app doesn't have a web server set up yet, so `allow_unsafe_werkzeug=True)` is used.

Auto-start chromium in kiosk mode with autoplay allowed:

```bash
mkdir -p ~/.config/labwc
touch ~/.config/labwc/autostart
chmod +x ~/.config/labwc/autostart
echo "/usr/bin/chromium --kiosk --force-device-scale-factor=1.50 --autoplay-policy=no-user-gesture-required http://localhost:5000 &" > ~/.config/labwc/autostart
```
