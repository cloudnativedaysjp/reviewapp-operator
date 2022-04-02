package utils

import "time"

/* DatetimeFactory (for mocking time.Now()) */

type DatetimeFactory struct {
	nowFunc func() time.Time
}

func NewDatetimeFactory() *DatetimeFactory {
	return &DatetimeFactory{func() time.Time { return time.Now().UTC() }}
}

func NewDatetimeMockFactory(t time.Time) *DatetimeFactory {
	return &DatetimeFactory{func() time.Time { return t.UTC() }}
}

func (f DatetimeFactory) Now() datetime {
	return datetime(f.nowFunc().Format(time.RFC3339))
}

/* Datetime */

type datetime string

func NewDatetime(str string) (datetime, error) {
	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return "", err
	}
	return datetime(t.UTC().Format(time.RFC3339)), nil
}

func (m datetime) ToString() string {
	return string(m)
}

func (m datetime) Before(d datetime, duration time.Duration) bool {
	a, _ := time.Parse(time.RFC3339, string(m))
	b, _ := time.Parse(time.RFC3339, string(d))
	return a.Before(b.Add(duration))
}
