from datetime import datetime, timedelta
import numpy as np

done_hrs = timedelta(hours=8)
day_length_hrs = timedelta(hours=8)

start_date = datetime.now()
current_date = start_date + timedelta(days=0, minutes=1)

business_days = np.busday_count(
    np.datetime64(start_date.date(), "D"),
    np.datetime64(current_date.date(), "D") + np.timedelta64(1, "D")
)

print("business days", business_days)

remaining = (day_length_hrs * business_days) - done_hrs

print("remaining hours", remaining.total_seconds() // 3600)
