package db

import "context"

type OJ struct {
	OjId   int    `json:"oj_id" db:"oj_id"`
	OjName string `json:"oj_name" db:"oj_name"`
}

func GetAllOJ(ctx context.Context) []OJ {
	query := "SELECT * FROM oj WHERE oj_id > 0"
	ret := make([]OJ, 0)
	mustSelect(ctx, &ret, query)
	return ret
}

// GetOjMap return map[oj.id]{oj.name}
func GetOjMap(ctx context.Context) map[int]string {
	oj := GetAllOJ(ctx)
	ret := make(map[int]string)
	for _, x := range oj {
		ret[x.OjId] = x.OjName
	}
	return ret
}

type Account struct {
	Username string `json:"username" db:"username"`
	OjId     int    `json:"oj_id" db:"oj_id"`
	Account  string `json:"account" db:"account"`
}

func GetAccountByUsername(ctx context.Context, username string) []Account {
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
	query := "DELETE FROM submission WHERE username=:username AND oj_id=:oj_id"
	mustNamedExecTx(tx, ctx, query, account)
	query = `INSERT INTO oj_user_rel(oj_id, username, account)
VALUES(:oj_id, :username, :account) ON DUPLICATE KEY UPDATE account=VALUES(account)`
	mustNamedExecTx(tx, ctx, query, account)
	mustCommit(tx)
}

// GetEnableAccount return all enable accounts
func GetEnableAccount(ctx context.Context) (ret []Account) {
	mustSelect(ctx, &ret, "SELECT * FROM oj_user_rel WHERE username IN (SELECT username FROM user WHERE is_enable=1)")
	return
}

type ojAccount struct {
	OjId     int
	Accounts []Account
}

func GetAllAccounts(ctx context.Context) []ojAccount {
	query := "SELECT * FROM oj_user_rel WHERE oj_id > 0"
	data := make([]Account, 0)
	mustSelect(ctx, &data, query)

	oj := GetAllOJ(ctx)
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
