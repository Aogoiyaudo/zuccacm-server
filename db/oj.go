package db

import "context"

type OJ struct {
	OjId   int    `json:"oj_id" db:"oj_id"`
	OjName string `json:"oj_name" db:"oj_name"`
}

func GetOjAll(ctx context.Context) []OJ {
	query := "SELECT * FROM oj WHERE oj_id > 0"
	ret := make([]OJ, 0)
	mustSelect(ctx, &ret, query)
	return ret
}

// GetOjMap return map[oj.id]{oj.name}
func GetOjMap(ctx context.Context) map[int]string {
	oj := GetOjAll(ctx)
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
func UpdAccount(ctx context.Context, account Account) {
	query := `INSERT INTO oj_user_rel(oj_id, username, account)
VALUES(:oj_id, :username, :account) ON DUPLICATE KEY UPDATE account=VALUES(account)`
	mustNamedExec(ctx, query, account)
}

// GetEnableAccount return all enable accounts
func GetEnableAccount(ctx context.Context) (ret []Account) {
	mustSelect(ctx, &ret, "SELECT * FROM oj_user_rel WHERE username IN (SELECT username FROM user WHERE is_enable=1)")
	return
}
