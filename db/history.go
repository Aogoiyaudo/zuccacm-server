package db

import (
	"context"
	"time"
)

type History struct {
	Id         int       `json:"id" db:"id"`
	Name       string    `json:"historyname" db:"name"`
	Start_time time.Time `json:"start_time" db:"start_time"`
	End_time   time.Time `json:"end_time" db:"end_time"`
}

// GetHistorys return all Historys
func GetHistorys(ctx context.Context, isEnable bool) []History {
	historys := make([]History, 0)
	query := "SELECT * FROM History"
	if isEnable {
		query += " WHERE is_enable=true"
	}
	mustSelect(ctx, &historys, query)
	return historys
}
func AddHistory(ctx context.Context, history History) {
	tx := instance.MustBeginTx(ctx, nil)
	defer tx.Rollback()
	res := mustNamedExecTx(tx, ctx, addHistorySQL, history)
	_, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}
	mustCommit(tx)
}
