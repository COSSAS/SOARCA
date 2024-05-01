package time

import "time"

type ITime interface {
	Now() time.Time
}

type Time struct {
}

func (t *Time) Now() time.Time {
	return time.Now()
}
