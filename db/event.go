package db

import (
	"context"
	"time"
)

type Event struct {
	Id         int       `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	Start_time time.Time `json:"start_time" db:"start_time"`
	End_time   time.Time `json:"end_time" db:"end_time"`
}

// GetEvents return all events
func GetEvents(ctx context.Context, isEnable bool) []Event {
	events := make([]Event, 0)
	query := "SELECT * FROM event"
	if isEnable {
		query += " WHERE is_enable=true"
	}
	mustSelect(ctx, &events, query)
	return events
}
func AddEvent(ctx context.Context, event Event) {
	tx := instance.MustBeginTx(ctx, nil)
	defer tx.Rollback()
	res := mustNamedExecTx(tx, ctx, addEventSQL, event)
	_, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}
	mustCommit(tx)
}
