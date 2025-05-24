package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var ErrGettingOrderStats = errors.New("error getting order statuses")

type OrderStatus struct {
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type ResponseFromProvider struct {
	Msg  string        `json:"message"`
	Data []OrderStatus `json:"data"`
}

func GetOrderStatusFromProvider(url string, id uint64) (*ResponseFromProvider, error) {
	url = fmt.Sprintf("%s/%d", url, id)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, ErrGettingOrderStats
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var ret ResponseFromProvider
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}
