package Basic

import (
	"time"
)

// 時間格式常數
const (
	StdLongMonth      = "January"
	StdMonth          = "Jan"
	StdNumMonth       = "1"
	StdZeroMonth      = "01"
	StdLongWeekDay    = "Monday"
	StdWeekDay        = "Mon"
	StdDay            = "2"
	StdUnderDay       = "_2"
	StdZeroDay        = "02"
	StdHour           = "15"
	StdHour12         = "3"
	StdZeroHour12     = "03"
	StdMinute         = "4"
	StdZeroMinute     = "04"
	StdSecond         = "5"
	StdZeroSecond     = "05"
	StdLongYear       = "2006"
	StdYear           = "06"
	StdPM             = "PM"
	Stdpm             = "pm"
	StdTZ             = "MST"
	StdISO8601TZ      = "Z0700"  // prints Z for UTC
	StdISO8601ColonTZ = "Z07:00" // prints Z for UTC
	StdNumTZ          = "-0700"  // always numeric
	StdNumShortTZ     = "-07"    // always numeric
	StdNumColonTZ     = "-07:00" // always numeric
)

// 日期格式 2017-07-31; glue "-"
func DateFormat(t time.Time, glue string) string {
	return t.Format(StdLongYear + glue + StdZeroMonth + glue + StdZeroDay)
}

func ToDay(timeZone string) string {
	var zone int

	switch timeZone {
	case "US_EST":
		zone = -4
	case "Taipei":
		zone = 8
	default:
		zone = 0
	}

	now := time.Now().UTC().Add(time.Duration(zone) * time.Hour)
	dateTime := DateFormat(now, "-")

	return dateTime
}

func GetDayBeforeMonth(timeZone string, keep int) []string {
	list := 30

	dateTime := ToDay(timeZone)
	onvertTime, _ := time.Parse("2006-01-02", dateTime)

	var dayList = make([]string, list)
	for i := 0; i < list; i++ {
		// 調整時間
		ajustedTime := onvertTime.Add(time.Duration(-24*keep-24*i) * time.Hour)
		day := ajustedTime.Format("2006-01-02")
		dayList[i] = day
	}

	return dayList
}
