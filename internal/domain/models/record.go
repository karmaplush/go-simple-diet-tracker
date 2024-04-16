package models

import "time"

type Record struct {
	Id          int64     `json:"id"`
	AccountId   int64     `json:"accountId"`
	Value       int       `json:"value"`
	DateRecord  time.Time `json:"dateRecord"`
	DateCreated time.Time `json:"dateCreated"`
}
