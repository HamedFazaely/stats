package db

import (
	"context"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type OrderRepo interface {
	CreateOrder(ctx context.Context, ord *Order) (*Order, error)
	GetPickedupNoneDeliveredOrders(ctx context.Context) ([]Order, error)
}

type OrderRepoImpl struct {
	db *sql.DB
}

func NewOrderRepoImpl(db *sql.DB) *OrderRepoImpl {
	ret := &OrderRepoImpl{
		db: db,
	}
	return ret
}

func (o *OrderRepoImpl) CreateOrder(ctx context.Context, ord *Order) (*Order, error) {
	_, err := o.db.ExecContext(ctx, "INSERT INTO orders (order_id, current_status, status_endpoint) VALUES (?,?,?)", ord.OrderID, ord.CurrentStatus, ord.StatusEP)
	if err != nil {
		return nil, err
	}

	return ord, nil
}

func (o *OrderRepoImpl) GetPickedupNoneDeliveredOrders(ctx context.Context) ([]Order, error) {
	rows, err := o.db.QueryContext(ctx, "SELECT order_id, status_endpoint FROM orders WHERE current_status IN (?,?,?)", ProviderSeen, PickedUp, InProgress)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ret []Order
	for rows.Next() {
		var ord Order
		err = rows.Scan(&ord.OrderID, &ord.StatusEP)
		if err != nil {
			return nil, err
		}
		ret = append(ret, ord)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}
