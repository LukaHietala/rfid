from datetime import datetime, timedelta, date, time
from dateutil.relativedelta import relativedelta
import numpy as np

def get_student_remaining(student):
    done_seconds = student["done_seconds"]
    day_starts_str = student["start_time"]
    day_ends_str = student["end_time"]
    start_date_str = student["start_date"] 
    end_date_str = student["end_date"]
    weekmask = student["weekmask"]

    done_time = timedelta(seconds=done_seconds)
    day_starts_time = datetime.strptime(day_starts_str, "%H.%M").time()
    day_ends_time = datetime.strptime(day_ends_str, "%H.%M").time()

    day_length = datetime.combine(date.today(), day_ends_time) - datetime.combine(date.today(), day_starts_time)

    date_format = '%d.%m.%Y %H:%M:%S'

    start_date = datetime.strptime(start_date_str, date_format)
    end_date = datetime.strptime(end_date_str, date_format)

    current_date = datetime.now()

    business_days = np.busday_count(
        np.datetime64(start_date.date(), "D"),
        np.datetime64(min(current_date.date(), end_date.date()), "D") + np.timedelta64(1, "D"),
        weekmask=weekmask
    )

    return (((datetime.combine(date.today(), current_date.time()) - datetime.combine(date.today(), day_starts_time))) + (day_length * (business_days - 1)) - done_time).total_seconds()

