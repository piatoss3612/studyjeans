package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func FormatSnowflakeToTime(s string) (string, error) {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return "", nil
	}

	timestamp := (n >> 22) + 1420070400000
	t := time.Unix(timestamp/1000, 0)

	creationTime := fmt.Sprintf("%d년 %d월 %d일", t.Year(), t.Month(), t.Day())
	return creationTime, nil
}

func FormatUptime(t time.Time) string {
	regex := regexp.MustCompilePOSIX("[hms]")
	uptimeString := time.Since(t).String()
	var trimmedString string
	if strings.Contains(uptimeString, ".") {
		trimmedString = uptimeString[:strings.Index(uptimeString, ".")] + "s"
	} else {
		trimmedString = uptimeString
	}

	var splitUptime []string
	var (
		r []string
		z int
	)
	is := regex.FindAllStringIndex(trimmedString, -1)
	if is == nil {
		splitUptime = append(r, trimmedString)
	} else {
		for _, i := range is {
			r = append(r, trimmedString[z:i[1]])
			z = i[1]
		}
		splitUptime = append(r, trimmedString[z:])
	}

	uptime := strings.Join(splitUptime, " ")
	uptime = strings.Replace(uptime, "h", "시간", 1)
	uptime = strings.Replace(uptime, "m", "분", 1)
	uptime = strings.Replace(uptime, "s", "초", 1)

	return uptime
}

func FormatRebootDate(t time.Time) string {
	return fmt.Sprintf("%02d월 %02d일", int(t.Month()), t.Day())
}
