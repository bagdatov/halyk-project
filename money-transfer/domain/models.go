package domain

import "time"

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
}

type Account struct {
	ID              int64       `json:"ID,omitempty"`
	OwnerID         int64       `json:"OwnerID,omitempty"`
	IIN             string      `json:"IIN,omitempty"`
	Amount          int64       `json:"Amount,omitempty"`
	Registered      time.Time   `json:"Registered,omitempty"`
	LastTransaction Transaction `json:"LastTransaction,omitempty"`
}

type Transaction struct {
	ID         int64     `json:"ID,omitempty"`
	SenderID   int64     `json:"SenderID,omitempty"`
	ReceiverID int64     `json:"ReceiverID,omitempty"`
	Amount     int64     `json:"Amount,omitempty"`
	Date       time.Time `json:"Date,omitempty"`
}
