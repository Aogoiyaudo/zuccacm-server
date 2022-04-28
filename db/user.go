package db

import (
	"context"
	"database/sql"
	"sort"

	log "github.com/sirupsen/logrus"
)

type UserSimple struct {
	Username string `json:"username" db:"username"`
	Nickname string `json:"nickname" db:"nickname"`
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

func GetUsers(ctx context.Context, isEnable, isOfficial bool, page Page) []User {
	query := "SELECT * FROM user"
	if isEnable && isOfficial {
		query += ` WHERE is_enable = true
AND username IN
(
    SELECT username
    FROM team_user_rel, team_group_rel, team_group
    WHERE team_user_rel.team_id = team_group_rel.team_id
    AND team_group.group_id = team_group_rel.group_id
    AND is_grade
)`
	} else if isEnable {
		query += " WHERE is_enable = true"
	} else if isOfficial {
		query += ` WHERE username IN
(
    SELECT username
    FROM team_user_rel, team_group_rel, team_group
    WHERE team_user_rel.team_id = team_group_rel.team_id
    AND team_group.group_id = team_group_rel.group_id
    AND is_grade
)`
	}
	users := make([]User, 0)
	mustSelect(ctx, &users, page.query(query))
	return users
}

func GetGroupsByUser(ctx context.Context, username string, isGrade bool) []TeamGroup {
	query := `SELECT * FROM team_group
WHERE group_id IN
(
    SELECT group_id FROM team_user_rel, team_group_rel
    WHERE team_user_rel.team_id = team_group_rel.team_id
    AND username=?
)
AND is_grade=?`
	groups := make([]TeamGroup, 0)
	err := instance.SelectContext(ctx, &groups, query, username, isGrade)
	if err == sql.ErrNoRows {
		err = nil
	}
	if err != nil {
		panic(err)
	}
	for i := range groups {
		groups[i].Teams = make([]Team, 0)
	}
	return groups
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

// UpdUser update User basic info (nickname, id_card, phone, qq, t_shirt)
// self-team will update Team.Name at the same time
func UpdUser(ctx context.Context, user User) {
	tx := instance.MustBeginTx(ctx, nil)
	defer tx.Rollback()
	query := `UPDATE user
SET nickname=:nickname, id_card=:id_card, phone=:phone, qq=:qq, t_shirt=:t_shirt
WHERE username=:username`
	mustNamedExecTx(tx, ctx, query, user)
	query = "UPDATE team SET name=:name WHERE id=:id"

	team := GetTeamBySelf(ctx, user.Username)
	team.Name = user.Nickname
	mustNamedExecTx(tx, ctx, query, team)
	mustCommit(tx)
}

func UpdUserAdmin(ctx context.Context, user User) {
	query := "UPDATE user SET is_admin=:is_admin WHERE username=:username"
	mustNamedExec(ctx, query, user)
}

func UpdUserRating(ctx context.Context, users []User) {
	query := `INSERT INTO user(username, cf_rating) VALUES(:username, :cf_rating)
ON DUPLICATE KEY UPDATE cf_rating=VALUES(cf_rating)`
	mustNamedExec(ctx, query, users)
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

func UpdUserGroups(ctx context.Context, username string, groups []int) {
	team := GetTeamBySelf(ctx, username)
	teamGroups := make([]TeamGroupRel, 0)
	for _, g := range groups {
		teamGroups = append(teamGroups, TeamGroupRel{
			GroupId: g,
			TeamId:  team.Id,
		})
	}

	tx := instance.MustBeginTx(ctx, nil)
	query := "DELETE FROM team_group_rel WHERE team_id=?"
	mustExecTx(tx, ctx, query, team.Id)
	if len(teamGroups) > 0 {
		query = "INSERT INTO team_group_rel(group_id, team_id) VALUES(:group_id, :team_id)"
		mustNamedExecTx(tx, ctx, query, teamGroups)
	}
	mustCommit(tx)
}

type Award struct {
	Username string `json:"username" db:"username"`
	Medal    int    `json:"medal" db:"medal"`
	Award    string `json:"award" db:"award"`
	XcpcId   int    `json:"xcpc_id" db:"xcpc_id"`
}

// GetAwardsByUsername return awards of 1 user
func GetAwardsByUsername(ctx context.Context, username string) []Award {
	query := getAwardsSQL + " AND user.username=? ORDER BY xcpc_date"
	ret := make([]Award, 0)
	mustSelect(ctx, &ret, query, username)
	return ret
}

// GetAwardsAll return awards of all users
// only return enable users if isEnable=true
func GetAwardsAll(ctx context.Context, isEnable bool) []Award {
	query := getAwardsSQL
	if isEnable {
		query += " AND is_enable"
	}
	query += " ORDER BY xcpc_date"
	ret := make([]Award, 0)
	mustSelect(ctx, &ret, query)
	return ret
}

type userGroup struct {
	GroupId   int    `json:"group_id" db:"group_id"`
	GroupName string `json:"group_name" db:"group_name"`
	Users     []User `json:"users"`
}

// GetOfficialGroups return official groups without users
// Official groups are groups which is_grade=true, such as 2018, 2019
func GetOfficialGroups(ctx context.Context) []userGroup {
	query := `SELECT group_id, group_name FROM team_group WHERE is_grade`
	groups := make([]userGroup, 0)
	mustSelect(ctx, &groups, query)
	for i := range groups {
		groups[i].Users = make([]User, 0)
	}
	return groups
}

// GetOfficialUsers return official groups with users
// return all users if is_enable=false
// groups with no user will be ignored
// each user can be in at most 1 group at a time
// official group should only contain teams with is_self=true
func GetOfficialUsers(ctx context.Context, isEnable bool) []userGroup {
	grp := make(map[int]*userGroup)
	groups := GetOfficialGroups(ctx)
	for _, row := range groups {
		grp[row.GroupId] = &userGroup{
			GroupId:   row.GroupId,
			GroupName: row.GroupName,
			Users:     make([]User, 0),
		}
	}
	query := `SELECT user.username, nickname, cf_rating, is_enable, is_admin, team_group.group_id
FROM user, team_group_rel, team_group, team_user_rel
WHERE user.username = team_user_rel.username AND team_user_rel.team_id = team_group_rel.team_id
AND team_group.group_id = team_group_rel.group_id AND is_grade`
	if isEnable {
		query += " AND is_enable"
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
		grp[x.GroupId].Users = append(grp[x.GroupId].Users, User{
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
