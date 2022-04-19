package db

import (
	"context"
	"database/sql"
	"sort"

	log "github.com/sirupsen/logrus"
)

// SimpleUser hide some private info, such like id_card, phone
type SimpleUser struct {
	Username string `json:"username" json:"username"`
	Nickname string `json:"nickname" json:"nickname"`
	CfRating int    `json:"cf_rating" json:"cf_rating"`
	IsEnable bool   `json:"is_enable" json:"is_enable"`
	IsAdmin  bool   `json:"is_admin" json:"is_admin"`
}

type User struct {
	Username string `db:"username" json:"username"`
	Nickname string `db:"nickname" json:"nickname"`
	CfRating int    `db:"cf_rating" json:"cf_rating"`
	IsEnable bool   `db:"is_enable" json:"is_enable"`
	IsAdmin  bool   `db:"is_admin" json:"is_admin"`
	IdCard   string `db:"id_card" json:"id_card"`
	Phone    string `db:"phone" json:"phone"`
	QQ       string `db:"qq" json:"qq"`
	TShirt   string `db:"t_shirt" json:"t_shirt"`
}

// GetUserByUsername return nil when user not found
func GetUserByUsername(ctx context.Context, username string) (ret *User) {
	ret = &User{}
	err := instance.GetContext(ctx, ret, "SELECT * FROM user WHERE username = ?", username)
	if err == sql.ErrNoRows {
		log.WithField("username", username).Warn("user not found")
		ret = nil
		err = nil
	}
	if err != nil {
		panic(err)
	}
	return
}

func AddUser(ctx context.Context, user User) {
	tx := instance.MustBeginTx(ctx, nil)
	defer tx.Rollback()

	mustNamedExecTx(tx, ctx, addUserSQL, user)
	team := Team{
		Name:     user.Nickname,
		IsEnable: user.IsEnable,
		IsSelf:   true,
	}
	ret := mustNamedExecTx(tx, ctx, addTeamSQL, team)
	tmp, err := ret.LastInsertId()
	if err != nil {
		panic(err)
	}
	team.Id = int(tmp)
	mustNamedExecTx(tx, ctx, addTeamUserRelSQL, TeamUser{team.Id, user.Username})
	mustCommit(tx)
}

func UpdUser(ctx context.Context, user User) {
	team := GetTeamBySelf(ctx, user.Username)
	team.Name = user.Nickname
	tx := instance.MustBeginTx(ctx, nil)
	defer tx.Rollback()
	mustNamedExecTx(tx, ctx, updUserSQL, user)
	mustNamedExecTx(tx, ctx, updTeamSQL, team)
	mustCommit(tx)
}

func UpdUserAdmin(ctx context.Context, user User) {
	mustNamedExec(ctx, updUserAdminSQL, user)
}

func UpdUserEnable(ctx context.Context, user User) {
	team := GetTeamBySelf(ctx, user.Username)
	team.IsEnable = user.IsEnable
	tx := instance.MustBeginTx(ctx, nil)
	defer tx.Rollback()
	mustNamedExec(ctx, updUserEnableSQL, user)
	mustNamedExec(ctx, updTeamEnableSQL, team)
	mustCommit(tx)
}

type Account struct {
	OjId     int    `json:"oj_id" db:"oj_id"`
	Username string `json:"username" db:"username"`
	Account  string `json:"account" db:"account"`
}

// UpdAccount update if account already exists, otherwise insert
func UpdAccount(ctx context.Context, account Account) {
	mustNamedExec(ctx, "INSERT INTO oj_user_rel(oj_id, username, account) VALUES(:oj_id, :username, :account) ON DUPLICATE KEY UPDATE account=VALUES(account)", account)
}

func GetEnableAccount(ctx context.Context) (ret []Account) {
	mustSelect(ctx, &ret, "SELECT * FROM oj_user_rel WHERE username IN (SELECT username FROM user WHERE is_enable=1)")
	return
}

// GetUserAward return all user-award if isEnable=false
func GetUserAward(ctx context.Context, isEnable bool) (ret []struct {
	Username string `db:"username"`
	Medal    int    `db:"medal"`
	Award    string `db:"award"`
}) {
	query := `SELECT user.username AS username, medal, award
FROM user, team_user_rel, xcpc_team_rel, xcpc
WHERE user.username=team_user_rel.username
AND team_user_rel.team_id=xcpc_team_rel.team_id
AND xcpc.id = xcpc_team_rel.xcpc_id`
	if isEnable {
		query += " AND is_enable=true"
	}
	query += " ORDER BY xcpc_date"
	mustSelect(ctx, &ret, query)
	return
}

type userGroup struct {
	GroupId   int          `json:"group_id" db:"group_id"`
	GroupName string       `json:"group_name" db:"group_name"`
	Users     []SimpleUser `json:"users"`
}

// GetOfficialUserGroups return official groups without users
// Official groups are groups which is_grade=true, such as 2018, 2019
func GetOfficialUserGroups(ctx context.Context) []userGroup {
	query := `SELECT group_id, group_name FROM user_group WHERE is_grade=true`
	ret := make([]userGroup, 0)
	mustSelect(ctx, &ret, query)
	return ret
}

// GetOfficialUsers return official groups with users
// return all users if is_enable=false
// groups with no user will be ignored
func GetOfficialUsers(ctx context.Context, isEnable bool) []userGroup {
	grp := make(map[int]*userGroup)
	groups := GetOfficialUserGroups(ctx)
	for _, row := range groups {
		grp[row.GroupId] = &userGroup{
			GroupId:   row.GroupId,
			GroupName: row.GroupName,
			Users:     make([]SimpleUser, 0),
		}
	}
	query := `SELECT user.username AS username, nickname, cf_rating, is_enable, is_admin, user_group.group_id AS group_id
FROM user, user_group_rel, user_group
WHERE user.username = user_group_rel.username
AND user_group.group_id = user_group_rel.group_id AND is_grade=true`
	if isEnable {
		query += " AND is_enable=true"
	}
	query += " ORDER BY group_name DESC"
	var data []struct {
		Username string `db:"username"`
		Nickname string `db:"nickname"`
		CfRating int    `db:"cf_rating"`
		IsEnable bool   `db:"is_enable"`
		IsAdmin  bool   `db:"is_admin"`
		GroupId  int    `db:"group_id"`
	}
	mustSelect(ctx, &data, query)
	for _, x := range data {
		grp[x.GroupId].Users = append(grp[x.GroupId].Users, SimpleUser{
			Username: x.Username,
			Nickname: x.Nickname,
			CfRating: x.CfRating,
			IsEnable: x.IsEnable,
			IsAdmin:  x.IsAdmin,
		})
	}
	ret := make([]userGroup, 0)
	for _, v := range grp {
		if len(v.Users) > 0 {
			ret = append(ret, *v)
		}
	}
	sort.SliceStable(ret, func(i, j int) bool {
		return ret[i].GroupName > ret[j].GroupName
	})
	return ret
}
