package core

import "time"

type Review struct {
	ID         string    `json:"id"`
	CardID     string    `json:"cardId"`
	Rating     Rating    `json:"rating"`
	ReviewedAt time.Time `json:"reviewedAt"`
	Previous   State     `json:"previous"`
	Next       State     `json:"next"`
}
