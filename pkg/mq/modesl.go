package mq

import "time"

type StatusMsg struct {
	OrderID   uint64    `json:"order_id"`
	Status    string    `json:"status"`
	StatusEP  string    `json:"status_ep"`
	CreatedAt time.Time `json:"created_at"`
}

type StatusPubError struct {
	Err error
	Msg *StatusMsg
}
