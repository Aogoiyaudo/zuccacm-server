package db

import (
	"context"
	"database/sql"
	log "github.com/sirupsen/logrus"
	"time"
	"zuccacm-server/enum/errorx"
)

type History struct {
	Id         int       `json:"id" db:"id"`
	Name       string    `json:"historyname" db:"name"`
	Start_time time.Time `json:"start_time" db:"start_time"`
	End_time   time.Time `json:"end_time" db:"end_time"`
	Md         string    `json:"md" db:"md"`
}

// GetHistorys return all Historys
func GetHistoryById(ctx context.Context, id int) (ret *History) {
	ret = &History{}
	err := instance.GetContext(ctx, ret, "SELECT * FROM history WHERE id = ?", id)
	if err == sql.ErrNoRows {
		log.WithField("history id = ", id).Error(" history not found")
	}
	return ret
}
func MustGetHistory(ctx context.Context, id int) (ret *History) {
	ret = GetHistoryById(ctx, id)
	if ret == nil {
		panic(errorx.ErrNotFound.New())
	}
	return
}
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
func UpdHistory(ctx context.Context, id int, md string) {
	tx := instance.MustBeginTx(ctx, nil)
	history := MustGetHistory(ctx, id)
	history.Md = md
	defer tx.Rollback()
	query := `UPDATE history
SET id=:id, name=:name, start_time=:start_time, end_time=:end_time, md=:md
WHERE id=:id`
	mustNamedExecTx(tx, ctx, query, history)
	mustCommit(tx)
}
