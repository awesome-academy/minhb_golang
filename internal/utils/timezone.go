package utils

import "time"

var Location = time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60)

func FormatVN(t time.Time, layout string) string {
	return t.In(Location).Format(layout)
}

func DayRange(t time.Time) (time.Time, time.Time) {
	t = t.In(Location)
	from := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, Location)
	return from, from.AddDate(0, 0, 1)
}
