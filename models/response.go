package models

type AuthFailBadRequest struct {
	Errors []string `json:"errors" example:"bad request"`
}

type AuthFailInternalServerError struct {
	Errors []string `json:"errors" example:"internal server error"`
}
