package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"zuccacm-server/config"
)

var instance *sqlx.DB

func init() {
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		config.Instance.DBConfig.User,
		config.Instance.DBConfig.Pwd,
		config.Instance.DBConfig.Host,
		config.Instance.DBConfig.Port,
		config.Instance.DBConfig.Database,
	)
	db, err := sqlx.Connect("mysql", dataSource)
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10) // size of connect pool
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetConnMaxIdleTime(time.Minute * 3)
	instance = db
}

func mustGet(ctx context.Context, dest interface{}, query string, args ...interface{}) {
	if err := instance.GetContext(ctx, dest, query, args...); err != nil {
		panic(err)
	}
}

func mustSelect(ctx context.Context, dest interface{}, query string, args ...interface{}) {
	if err := instance.SelectContext(ctx, dest, query, args...); err != nil {
		panic(err)
	}
}

func mustExec(ctx context.Context, query string, args ...interface{}) {
	instance.MustExecContext(ctx, query, args...)
}

func mustExecTx(tx *sqlx.Tx, ctx context.Context, query string, args ...interface{}) {
	tx.MustExecContext(ctx, query, args...)
}

func mustNamedExecTx(tx *sqlx.Tx, ctx context.Context, query string, arg interface{}) sql.Result {
	ret, err := tx.NamedExecContext(ctx, query, arg)
	if err != nil {
		panic(err)
	}
	return ret
}

func mustNamedExec(ctx context.Context, query string, arg interface{}) sql.Result {
	ret, err := instance.NamedExecContext(ctx, query, arg)
	if err != nil {
		panic(err)
	}
	return ret
}

func mustCommit(tx *sqlx.Tx) {
	if err := tx.Commit(); err != nil {
		panic(err)
	}
}

// Datetime is used to deal with json time instead of time.Time
type Datetime time.Time

func (t Datetime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", t.String())), nil
}

func (t *Datetime) UnmarshalJSON(b []byte) error {
	dt, err := time.ParseInLocation("\"2006-01-02 15:04:05\"", string(b), time.Local)
	*t = Datetime(dt)
	return err
}

func (t Datetime) String() string {
	return time.Time(t).Format("2006-01-02 15:04:05")
}

func (t Datetime) Date() string {
	return time.Time(t).Format("2006-01-02")
}

func (t Datetime) Unix() int64 {
	return time.Time(t).Unix()
}
