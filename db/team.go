package db

import "context"

type Team struct {
	Id       int    `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	IsEnable bool   `json:"is_enable" db:"is_enable"`
	IsSelf   bool   `json:"is_self" db:"is_self"`
	Users    []User `json:"users"`
}

type TeamUser struct {
	TeamId   int    `json:"team_id" db:"team_id"`
	Username string `json:"username" db:"username"`
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
			Users:  make([]User, 0),
		}
	}
	for _, x := range data {
		mpTeam[x.TeamId].Users = append(mpTeam[x.TeamId].Users, User{
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
