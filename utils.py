import re
from datetime import datetime

date_format = '%d.%m.%Y %H:%M:%S' # 10.8.2026 12:12:12
time_format = '%H.%M' # 10:20
weekmask_regex = '^[0,1]{7}$' # 1111100

def validate_weekmask(weekmask):
    return re.search(weekmask_regex, weekmask)

def validate_datetime(date_str):
    pass

def validate_time(time_str):
    pass

def format_datetime(date):
    return date.strftime(date_format)

def to_datetime(date_str):
    return datetime.strptime(date_str, date_format)

def to_time(time_str):
    return datetime.strptime(time_str, time_format).time()
