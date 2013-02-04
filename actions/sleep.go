package actions

import "time"

func Sleep(milliseconds int64) {
	time.Sleep(time.Duration(milliseconds) * time.Millisecond)
}
