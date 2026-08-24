import re
from datetime import datetime
import time

date_format = '%d.%m.%Y' # 10.08.2026
time_format = '%H:%M' # 10:20

def validate_datetime(date_str):
    pass

def validate_time(time_str):
    pass

def format_datetime(date):
    return date.strftime(date_format)

def to_datetime(date_str):
    return datetime.strptime(date_str, date_format)

def to_time(time_str):
    return datetime.strptime(time_str, time_format)
