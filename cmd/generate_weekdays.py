import csv
from datetime import date, timedelta

start_date = date(2026, 1, 1)
end_date = date(2026, 12, 31)

with open("weekdays.csv", "a", newline="") as csvfile:
    writer = csv.writer(csvfile)
    writer.writerow(["date"])

    current_date = start_date
    while current_date <= end_date:
        if current_date.weekday() < 5:
            writer.writerow([current_date.strftime("%Y%m%d")])
        current_date += timedelta(days=1)

print("Weekdays CSV file generated successfully.")