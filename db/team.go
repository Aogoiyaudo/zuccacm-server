package db

import (
	"context"
	"sort"
)

type Team struct {
	Id       int          `json:"id" db:"id"`
	Name     string       `json:"team_name" db:"name"`
	IsEnable bool         `json:"is_enable" db:"is_enable"`
	IsSelf   bool         `json:"is_self" db:"is_self"`
	Users    []UserSimple `json:"team_members" `
}

type TeamUser struct {
	TeamId   int    `json:"team_id" db:"team_id"`
	Username string `json:"username" db:"username"`
}

type TeamGroup struct {
	GroupId   int    `json:"group_id" db:"group_id"`
	GroupName string `json:"group_name" db:"group_name"`
	IsGrade   bool   `json:"is_grade" db:"is_grade"`
	Teams     []Team `json:"teams"`
}

type TeamGroupRel struct {
	GroupId int `json:"group_id" db:"group_id"`
	TeamId  int `json:"team_id" db:"team_id"`
}

// GetTeams return all teams with user
func GetTeams(ctx context.Context, isEnable bool) []Team {
	teams := make([]Team, 0)
	query := "SELECT * FROM team"
	if isEnable {
		query += " WHERE is_enable=true"
	}
	mustSelect(ctx, &teams, query)
	mp := make(map[int]int)
	for i, t := range teams {
		mp[t.Id] = i
		teams[i].Users = make([]UserSimple, 0)
	}

	data := make([]struct {
		TeamId   int    `db:"team_id"`
		Username string `db:"username"`
		Nickname string `db:"nickname"`
	}, 0)
	query = `SELECT team_id, user.username, nickname FROM team_user_rel, user
WHERE team_user_rel.username = user.username`
	if isEnable {
		query += " AND team_id IN (SELECT id FROM team WHERE is_enable)"
	}
	mustSelect(ctx, &data, query)
	for _, x := range data {
		i := mp[x.TeamId]
		teams[i].Users = append(teams[i].Users, UserSimple{
			Username: x.Username,
			Nickname: x.Nickname,
		})
	}
	return teams
}

func GetTeamGroups(ctx context.Context, isGrade bool) []TeamGroup {
	query := "SELECT * FROM team_group"
	if isGrade {
		query += " WHERE is_grade"
	}
	groups := make([]TeamGroup, 0)
	mustSelect(ctx, &groups, query)
	for i := range groups {
		groups[i].Teams = make([]Team, 0)
	}
	return groups
}

// GetTeamGroupsWithTeams return groups with teams and team_users
func GetTeamGroupsWithTeams(ctx context.Context, isGrade, isEnable, showEmpty bool) []TeamGroup {
	groups := GetTeamGroups(ctx, isGrade)
	teams := GetTeams(ctx, isEnable)
	groupId := make(map[int]int)
	teamId := make(map[int]int)
	for i, g := range groups {
		groupId[g.GroupId] = i
	}
	for i, t := range teams {
		teamId[t.Id] = i
	}

	query := "SELECT * FROM team_group_rel"
	if isGrade && isEnable {
		query += ` WHERE group_id IN (SELECT group_id FROM team_group WHERE is_grade)
AND team_id IN (SELECT id FROM team WHERE is_enable)`
	} else if isGrade {
		query += " WHERE group_id IN (SELECT group_id FROM team_group WHERE is_grade)"
	} else if isEnable {
		query += " WHERE team_id IN (SELECT id FROM team WHERE is_enable)"
	}
	data := make([]struct {
		GroupId int `db:"group_id"`
		TeamId  int `db:"team_id"`
	}, 0)
	mustSelect(ctx, &data, query)
	for _, x := range data {
		gid := groupId[x.GroupId]
		tid := teamId[x.TeamId]
		groups[gid].Teams = append(groups[gid].Teams, teams[tid])
	}
	ret := make([]TeamGroup, 0)
	for _, x := range groups {
		if len(x.Teams) > 0 || showEmpty {
			ret = append(ret, x)
		}
	}
	sort.SliceStable(ret, func(i, j int) bool {
		return ret[i].GroupName > ret[j].GroupName
	})
	return ret
}

func AddTeam(ctx context.Context, team Team) {
	tx := instance.MustBeginTx(ctx, nil)
	defer tx.Rollback()
	res := mustNamedExecTx(tx, ctx, addTeamSQL, team)
	id, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}
	team.Id = int(id)
	var users []TeamUser
	for _, user := range team.Users {
		users = append(users, TeamUser{team.Id, user.Username})
	}
	if len(users) > 0 {
		mustNamedExecTx(tx, ctx, addTeamUserRelSQL, users)
	}
	mustCommit(tx)
}

// GetTeamBySelf return the self team of this user
func GetTeamBySelf(ctx context.Context, username string) (ret Team) {
	mustGet(ctx, &ret, "SELECT * FROM team WHERE is_self=true AND id IN (SELECT team_id FROM team_user_rel WHERE username=?)", username)
	return
}

// GetTeamsInContest return teams (with users) in this contest
func GetTeamsInContest(ctx context.Context, contestId int) []Team {
	query := `SELECT id AS team_id, name AS team_name, is_self, user.username AS username, nickname
FROM team, team_user_rel, contest_team_rel, user
WHERE team_user_rel.team_id = contest_team_rel.team_id
AND team.id = team_user_rel.team_id
AND user.username = team_user_rel.username
AND contest_id = ?`
	var data []struct {
		TeamId   int    `db:"team_id"`
		TeamName string `db:"team_name"`
		IsSelf   bool   `db:"is_self"`
		Username string `db:"username"`
		Nickname string `db:"nickname"`
	}
	mustSelect(ctx, &data, query, contestId)
	mpTeam := make(map[int]*Team)
	for _, x := range data {
		mpTeam[x.TeamId] = &Team{
			Id:     x.TeamId,
			Name:   x.TeamName,
			IsSelf: x.IsSelf,
			Users:  make([]UserSimple, 0),
		}
	}
	for _, x := range data {
		mpTeam[x.TeamId].Users = append(mpTeam[x.TeamId].Users, UserSimple{
			Username: x.Username,
			Nickname: x.Nickname,
		})
	}
	ret := make([]Team, 0)
	for _, t := range mpTeam {
		ret = append(ret, *t)
	}
	return ret
}
func UpdTeamEnable(ctx context.Context, team Team) {
	tx := instance.MustBeginTx(ctx, nil)
	defer tx.Rollback()
	mustNamedExec(ctx, updTeamEnableSQL, team)
	mustCommit(tx)
}
