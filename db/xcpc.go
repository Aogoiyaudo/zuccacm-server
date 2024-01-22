package db

import (
	"context"
	"time"
)

type Xcpc struct {
	Id   int       `json:"id"   db:"id"`
	Name string    `json:"name" db:"name"`
	Date time.Time `json:"date" db:"date"`
}
type XcpcTeamRel struct {
	XcpcId int    `json:"xcpc_id" db:"xcpc_id"`
	TeamId int    `json:"team_id" db:"team_id"`
	Medal  int    `json:"medal" db:"medal"`
	Award  string `json:"award" db:"award"`
}

func GetXcpcs(ctx context.Context) []Xcpc {
	xcpcs := make([]Xcpc, 0)
	query := "SELECT * FROM xcpc"
	mustSelect(ctx, &xcpcs, query)
	return xcpcs
}
func AddXcpc(ctx context.Context, xcpc Xcpc) {
	tx := instance.MustBeginTx(ctx, nil)
	defer tx.Rollback()
	res := mustNamedExecTx(tx, ctx, addXcpcSQL, xcpc)
	_, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}
	mustCommit(tx)
}
