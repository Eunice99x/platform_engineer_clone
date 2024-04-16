package models

import "time"

type Token struct {
	Id        int       `json:"id" db:"id"`
	Key       string    `json:"key" db:"key"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	Revoked   bool      `json:"revoked" db:"revoked"`
	Expired   bool      `json:"expired" db:"expired"`
	CreatedBy string    `json:"created_by" db:"created_by"`
}
