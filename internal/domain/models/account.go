package models

type Account struct {
	Id         int64 `json:"id"`
	UserId     int64 `json:"userId"`
	DailyLimit int   `json:"dailyLimit"`
}
