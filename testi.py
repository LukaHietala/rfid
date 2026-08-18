from datetime import datetime, timedelta, date, time
import numpy as np

# Tietokannassa sekunteina tehdyt tunnit
done_seconds = 100000

done_time = timedelta(seconds=done_seconds)

# Tulee dbsta
day_starts_str = "8.00"
day_ends_str = "16.00"

# Muuttaa arvot merkkijonoista time objekteiksi
day_starts_time= datetime.strptime(day_starts_str, "%H.%M").time()
day_ends_time = datetime.strptime(day_ends_str, "%H.%M").time()

# Laskee paivan pituuden
day_length = datetime.combine(date.today(), day_ends_time) - datetime.combine(date.today(), day_starts_time)

# Date tulee dbsta seuraavalla formaatilla, TODO!: muista laittaa juuri tuolla formaatilla
start_date_str = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
date_format = '%Y-%m-%d %H:%M:%S'

start_date = datetime.strptime(start_date_str, date_format)

# 0 on tämä päivä, jne. timedelta lisäys vain debug varten
current_date = start_date + timedelta(days=1)

business_days = np.busday_count(
    np.datetime64(start_date.date(), "D"),
    np.datetime64(current_date.date(), "D") + np.timedelta64(1, "D"),
    weekmask="1111100",
    holidays=['2026-08-18']
) -1 # Jokin kirjasto offset defaulttina

print("business days", business_days)
left = ((datetime.combine(date.today(), current_date.time()) - datetime.combine(date.today(), day_starts_time))) + (day_length * business_days)

remaining = left - done_time


print("remaining hours", remaining.total_seconds() / 3600)

