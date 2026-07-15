package domain

import (
	"encoding/json"
	"time"
)

type ClientBackup struct {
	UserID    string          `json:"-"`
	Config    json.RawMessage `json:"config"`
	State     json.RawMessage `json:"state"`
	UpdatedAt time.Time       `json:"updatedAt"`
}
