package time

import "time"

type ITime interface {
	Now() time.Time
	Sleep(duration time.Duration)
}

type Time struct {
}

func (t *Time) Now() time.Time {
	return time.Now()
}

func (t *Time) Sleep(duration time.Duration) {
	time.Sleep(duration)
}
