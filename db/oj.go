package db

import (
	"context"
	"database/sql"
	log "github.com/sirupsen/logrus"
)

type OJ struct {
	OjId   int    `json:"oj_id" db:"oj_id"`
	OjName string `json:"oj_name" db:"oj_name"`
}

func OJMapItoS(oj []OJ) map[int]string {
	mp := make(map[int]string)
	for _, x := range oj {
		mp[x.OjId] = x.OjName
	}
	return mp
}

func OJMapStoI(oj []OJ) map[string]int {
	mp := make(map[string]int)
	for _, x := range oj {
		mp[x.OjName] = x.OjId
	}
	return mp
}

func GetOJByName(ojName string) (oj OJ) {
	query := "SELECT oj_id, oj_name FROM oj WHERE oj_name = ?"
	if err := instance.Get(&oj, query, ojName); err != nil {
		panic(err)
	}
	return
}
func GetOJById(ojId int) (oj OJ) {
	query := "SELECT oj_id, oj_name FROM oj WHERE oj_id = ?"
	if err := instance.Get(&oj, query, ojId); err != nil {
		panic(err)
	}
	return
}
func GetAllOJ(ctx context.Context) []OJ {
	query := "SELECT oj_id, oj_name FROM oj WHERE oj_id > 0"
	ret := make([]OJ, 0)
	mustSelect(ctx, &ret, query)
	return ret
}

func GetAllEnableOJ(ctx context.Context) []OJ {
	query := "SELECT oj_id, oj_name FROM oj WHERE oj_id > 0 AND is_enable"
	ret := make([]OJ, 0)
	mustSelect(ctx, &ret, query)
	return ret
}

type Account struct {
	Username string `json:"username" db:"username"`
	OjId     int    `json:"oj_id" db:"oj_id"`
	Account  string `json:"account" db:"account"`
}

func GetAccount(ctx context.Context, username string, ojId int) (account string) {
	query := "SELECT account FROM oj_user_rel WHERE username=? AND oj_id=?"
	err := instance.GetContext(ctx, &account, query, username, ojId)
	if err == sql.ErrNoRows {
		log.WithFields(log.Fields{
			"username": username,
			"oj_id":    ojId,
		}).Warn("account not found")
		account, err = "", nil
	}
	if err != nil {
		panic(err)
	}
	return account
}

func GetAccountsByUsername(ctx context.Context, username string) []Account {
	query := "SELECT * FROM oj_user_rel WHERE username=?"
	ret := make([]Account, 0)
	mustSelect(ctx, &ret, query, username)
	return ret
}

// UpdAccount update if account already exists, otherwise insert
// this will cause the user's submissions on the OJ to be cleared
func UpdAccount(ctx context.Context, account Account) {
	tx := instance.MustBeginTx(ctx, nil)
	defer tx.Rollback()
	// clear submissions
	query := "DELETE FROM submission WHERE username=:username AND account_oj_id=:oj_id"
	mustNamedExecTx(tx, ctx, query, account)
	query = `INSERT INTO oj_user_rel(oj_id, username, account)
VALUES(:oj_id, :username, :account) ON DUPLICATE KEY UPDATE account=VALUES(account)`
	mustNamedExecTx(tx, ctx, query, account)
	mustCommit(tx)
}

func GetAllAccounts(ctx context.Context) []Account {
	query := `SELECT * FROM oj_user_rel WHERE oj_id > 0
AND oj_id IN (SELECT oj_id FROM oj WHERE is_enable)`
	ret := make([]Account, 0)
	mustSelect(ctx, &ret, query)
	return ret
}

func GetAccountsByOJ(ctx context.Context, ojId int) []Account {
	ret := make([]Account, 0)
	mustSelect(ctx, &ret, "SELECT * FROM oj_user_rel WHERE oj_id=?", ojId)
	return ret
}

type ojAccount struct {
	OjId     int
	Accounts []Account
}

func GetAllAccountsGroupByOJ(ctx context.Context) []ojAccount {
	data := GetAllAccounts(ctx)

	oj := GetAllEnableOJ(ctx)
	mp := make(map[int]int)
	ret := make([]ojAccount, 0)
	for i, x := range oj {
		mp[x.OjId] = i
		ret = append(ret, ojAccount{
			OjId:     x.OjId,
			Accounts: make([]Account, 0),
		})
	}
	for _, x := range data {
		i := mp[x.OjId]
		ret[i].Accounts = append(ret[i].Accounts, x)
	}
	return ret
}

func GetAllAccountsMap(ctx context.Context) map[Account]string {
	ret := make(map[Account]string)
	accounts := GetAllAccounts(ctx)
	for _, ac := range accounts {
		ret[Account{OjId: ac.OjId, Account: ac.Account}] = ac.Username
	}
	return ret
}
