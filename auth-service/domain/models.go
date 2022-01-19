package domain

import (
	"regexp"
	"time"
)

type User struct {
	ID         int64     `json:"ID,omitempty"`
	Email      string    `json:"Email,omitempty"`
	Password   string    `json:"Password,omitempty"`
	FirstName  string    `json:"FirstName,omitempty"`
	LastName   string    `json:"LastName,omitempty"`
	IIN        string    `json:"IIN,omitempty"`
	Registered time.Time `json:"Registered,omitempty"`
	Phone      string    `json:"Phone,omitempty"`
	Role       string    `json:"Role,omitempty"`
	Wallets    string    `json:"Wallets,omitempty"`
}

func (u *User) Valid() bool {

	if len(u.FirstName) < 2 || len(u.LastName) < 2 {
		return false
	}

	if len(u.Password) < 5 || len(u.Email) < 5 {
		return false
	}

	if len(u.IIN) < 12 || len(u.Phone) < 11 {
		return false
	}

	if re := regexp.MustCompile(`^([a-z0-9_-]+\.)*[a-z0-9_-]+@[a-z0-9_-]+(\.[a-z0-9_-]+)*\.[a-z]{2,6}$`); !re.MatchString(u.Email) {
		return false
	}
	return true
}
