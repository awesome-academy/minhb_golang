package utils

import "time"

var Location = time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60)

func FormatVN(t time.Time, layout string) string {
	return t.In(Location).Format(layout)
}
