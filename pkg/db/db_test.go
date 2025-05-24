package db

import (
	"context"
	"testing"
	"time"

	"gitlab.com/Hamed1984/stats/pkg/conf"
)

func TestCreateOrder(t *testing.T) {
	t.Setenv("DB_PASSWORD", "hamed1984")
	t.Setenv("MQ_PASSWORD", "hamed1984")
	conf := conf.GetConffiguration()
	db, err := NewMysqlConnection(conf)
	if err != nil {
		t.Fatal(err)
	}
	repo := NewOrderRepoImpl(db)
	res, err := repo.CreateOrder(context.Background(), &Order{
		OrderID:       60,
		CurrentStatus: "4",
		StatusEP:      "z.xckcjzxc';l",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}

func TestGetOrders(t *testing.T) {
	t.Setenv("DB_PASSWORD", "hamed1984")
	t.Setenv("MQ_PASSWORD", "hamed1984")
	conf := conf.GetConffiguration()
	db, err := NewMysqlConnection(conf)
	if err != nil {
		t.Fatal(err)
	}
	repo := NewOrderRepoImpl(db)
	res, err := repo.GetPickedupNoneDeliveredOrders(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range res {
		t.Log(x)
	}
}

func TestCreateHistories(t *testing.T) {
	t.Setenv("DB_PASSWORD", "hamed1984")
	t.Setenv("MQ_PASSWORD", "hamed1984")
	conf := conf.GetConffiguration()
	db, err := NewMysqlConnection(conf)
	if err != nil {
		t.Fatal(err)
	}
	repo := NewHistoryRepoImpl(db)
	var data []History
	for i := 0; i < 10; i++ {
		h := History{
			Status:    "3",
			OrderID:   54,
			CreatedAt: time.Now().Add(time.Duration((i+3)*5) * time.Second),
		}
		data = append(data, h)
	}
	err = repo.CreateHistories(context.Background(), data)
	if err != nil {
		t.Fatal(err)
	}
}
