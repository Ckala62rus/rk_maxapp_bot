package domain

import "time"

// SearchHistory stores audit for user search requests.
type SearchHistory struct {
	ID          int64
	UserID      int64
	Code        string
	DurationMs  int64
	Rows        int
	Success     bool
	ErrorMessage string
	CreatedAt   time.Time
}
