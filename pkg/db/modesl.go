package db

import "time"

const (
	Pending      = "0"
	ProviderSeen = "1"
	PickedUp     = "2"
	InProgress   = "3"
	Delivered    = "4"
)

type Order struct {
	OrderID       uint64
	CurrentStatus string
	StatusEP      string
	CreatedAt     time.Time
	UpdatedAt     *time.Time
}

type History struct {
	ID        uint64
	Status    string
	OrderID   uint64
	CreatedAt time.Time
}
