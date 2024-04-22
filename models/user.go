package models

type User struct {
	Id    int    `json:"id" json:"id"`
	Name  string `json:"name" json:"name"`
	Email string `json:"email" json:"email"`
}
