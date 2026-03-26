// Package domain contains core entities and interfaces.
package domain

import "time"

// User represents a MAX user stored in Postgres.
type User struct {
	ID         int64     `json:"id"`
	MaxUserID  int64     `json:"maxUserID"`
	MaxUsername string   `json:"maxUsername"`
	MaxFirstName string  `json:"maxFirstName"`
	MaxLastName  string  `json:"maxLastName"`
	LanguageCode string  `json:"languageCode"`
	PhotoURL     string  `json:"photoUrl"`
	FirstName  string    `json:"firstName"`
	LastName   string    `json:"lastName"`
	IsAdmin    bool      `json:"isAdmin"`
	IsBlocked  bool      `json:"isBlocked"`
	IsApproved bool      `json:"isApproved"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// ProfileComplete returns true when user filled first/last name.
func (u User) ProfileComplete() bool {
	return u.FirstName != "" && u.LastName != ""
}
