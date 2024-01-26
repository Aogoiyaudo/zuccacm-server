package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
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

func GetXcpc(ctx context.Context, xcpc_id string) (ret *Xcpc, err error) {
	ret = &Xcpc{}
	err = instance.GetContext(ctx, ret, "SELECT * FROM xcpc WHERE id = ?", xcpc_id)
	if errors.Is(err, sql.ErrNoRows) {
		log.WithField("xcpc_id", xcpc_id).Warn("xcpc not found")
		ret = nil
		err = errors.New("xcpc not found")
		return ret, err
	}
	return ret, err
}
func GetXcpcs(ctx context.Context) []Xcpc {
	xcpcs := make([]Xcpc, 0)
	query := "SELECT * FROM xcpc"
	mustSelect(ctx, &xcpcs, query)
	return xcpcs
}
func GetXcpcTeamRels(ctx context.Context) []XcpcTeamRel {
	xcpc_team_rels := make([]XcpcTeamRel, 0)
	query := "SELECT * FROM xcpc_team_rel"
	mustSelect(ctx, &xcpc_team_rels, query)
	return xcpc_team_rels
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
func AddXcpcTeamRel(ctx context.Context, xcpc_team_rel XcpcTeamRel) {
	tx := instance.MustBeginTx(ctx, nil)
	defer tx.Rollback()
	fmt.Println(xcpc_team_rel.TeamId)
	fmt.Println(xcpc_team_rel.XcpcId)
	res := mustNamedExecTx(tx, ctx, addXcpcTeamRelSQL, xcpc_team_rel)
	_, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}
	mustCommit(tx)
}
