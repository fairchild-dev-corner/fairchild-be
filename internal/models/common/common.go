package models

import "time"

type SuccessResponse struct {
	SuccessID  string      `json:"successId"`
	Status     string      `json:"status"`
	HttpCode   int         `json:"httpCode"`
	ResponseAt time.Time   `json:"requestedAt"`
	Body       interface{} `json:"body"`
}

type ErrorResponse struct {
	ErrorCode int       `json:"errorCode"`
	TimeStamp time.Time `json:"timeStamp"`
	ErrorMsg  string    `json:"errorMsg"`
}
