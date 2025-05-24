package db

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"
)

type HistoryRepo interface {
	CreateHistories(ctx context.Context, data []History) error
}

type HistoryRepoImp struct {
	db *sql.DB
}

func NewHistoryRepoImpl(db *sql.DB) *HistoryRepoImp {
	ret := &HistoryRepoImp{
		db: db,
	}
	return ret
}

func (h *HistoryRepoImp) CreateHistories(ctx context.Context, data []History) error {
	if len(data) > 0 {
		insertHist := "INSERT INTO history (status, order_id, created_at) VALUES "
		var vals []string
		var dataVals []interface{}
		for _, x := range data {
			d := "(?,?,?)"
			vals = append(vals, d)
			dataVals = append(dataVals, x.Status)
			dataVals = append(dataVals, x.OrderID)
			dataVals = append(dataVals, x.CreatedAt)
		}
		valsStr := strings.Join(vals, ",")
		valsStr = strings.TrimRight(valsStr, ",")
		insertHist += valsStr
		tx, err := h.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer h.rollBack(tx)
		lastStat := data[len(data)-1]
		if lastStat.Status != Pending {
			_, err = tx.ExecContext(ctx, "UPDATE orders SET current_status = ?, updated_at = ? WHERE order_id = ?", lastStat.Status, lastStat.CreatedAt, lastStat.OrderID)
			if err != nil {
				return err
			}
		}

		_, err = tx.ExecContext(ctx, insertHist, dataVals...)
		if err != nil {
			return err
		}
		err = tx.Commit()
		if err != nil {
			return err
		}

	}
	return nil

}

func (h *HistoryRepoImp) rollBack(tx *sql.Tx) {
	err := tx.Rollback()
	if err != nil {
		if !errors.Is(err, sql.ErrTxDone) {
			log.Printf("an error ocuured during transaxtion rollback: %s\n", err.Error())
		}
	}

}
