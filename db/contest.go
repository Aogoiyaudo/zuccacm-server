package db

import (
	"context"
	"fmt"
	"sort"
	"time"
)

type ContestGroup struct {
	Id       int    `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	IsEnable bool   `json:"is_enable" db:"is_enable"`
}

type Problem struct {
	ContestId  int    `json:"contest_id" db:"contest_id"`
	OjId       int    `json:"oj_id" db:"oj_id"`
	Pid        string `json:"pid" db:"pid"`
	Index      string `json:"index" db:"index"`
	ProblemURL string `json:"problem_url,omitempty"`
}

type ContestGroupRel struct {
	GroupId   int `json:"group_id" db:"group_id"`
	ContestId int `json:"contest_id" db:"contest_id"`
}

type ContestTeamRel struct {
	ContestId int `json:"contest_id" db:"contest_id"`
	TeamId    int `json:"team_id" db:"team_id"`
}

type Contest struct {
	Id           int       `json:"id" db:"id"`
	OjId         int       `json:"oj_id" db:"oj_id"`
	Cid          string    `json:"cid" db:"cid"`
	Name         string    `json:"name" db:"name"`
	StartTime    Datetime  `json:"start_time" db:"start_time"`
	Duration     int       `json:"duration" db:"duration"`
	MaxSolved    int       `json:"max_solved" db:"max_solved"`
	Participants int       `json:"participants" db:"participants"`
	Problems     []Problem `json:"problems"`
	Groups       []int     `json:"groups"`
	Teams        []int     `json:"teams"`
}

type dbContest struct {
	Id           int       `json:"id" db:"id"`
	OjId         int       `json:"oj_id" db:"oj_id"`
	Cid          string    `json:"cid" db:"cid"`
	Name         string    `json:"name" db:"name"`
	StartTime    time.Time `json:"start_time" db:"start_time"`
	Duration     int       `json:"duration" db:"duration"`
	MaxSolved    int       `json:"max_solved" db:"max_solved"`
	Participants int       `json:"participants" db:"participants"`
	Problems     []Problem `json:"problems"`
	Groups       []int     `json:"groups"`
	Teams        []int     `json:"teams"`
}

func (c *Contest) dbType() *dbContest {
	return &dbContest{
		Id:           c.Id,
		OjId:         c.OjId,
		Cid:          c.Cid,
		Name:         c.Name,
		StartTime:    time.Time(c.StartTime),
		Duration:     c.Duration,
		MaxSolved:    c.MaxSolved,
		Participants: c.Participants,
		Problems:     c.Problems,
		Groups:       c.Groups,
		Teams:        c.Teams,
	}
}

// GetContestGroups return all groups if isEnable=false
func GetContestGroups(ctx context.Context, isEnable bool) []ContestGroup {
	query := "SELECT * FROM contest_group"
	if isEnable {
		query += " WHERE is_enable"
	}
	ret := make([]ContestGroup, 0)
	mustSelect(ctx, &ret, query)
	return ret
}

// GetContestsByGroup only get contests basic info (without problems)
// If group_id <= 0, return contests of any groups
func GetContestsByGroup(ctx context.Context, groupId int, begin, end time.Time, page Page) []Contest {
	query := `SELECT * FROM contest
WHERE start_time BETWEEN ? AND ?
AND id IN (SELECT contest_id FROM contest_group_rel`
	if groupId > 0 {
		query += fmt.Sprintf(" WHERE group_id = %d", groupId)
	}
	query += ")ORDER BY start_time DESC"
	ret := make([]Contest, 0)
	mustSelect(ctx, &ret, page.query(query), begin, end)
	for i := range ret {
		ret[i].Problems = make([]Problem, 0)
		ret[i].Groups = make([]int, 0)
		ret[i].Teams = make([]int, 0)
	}
	return ret
}

func GetGroupsByContest(ctx context.Context, contestId int) []ContestGroup {
	query := `SELECT * FROM contest_group WHERE id IN
      (SELECT group_id FROM contest_group_rel WHERE contest_id = ?)`
	groups := make([]ContestGroup, 0)
	mustSelect(ctx, &groups, query, contestId)
	return groups
}

// GetContestsByUser get contests (with problems) the user should participant in during [begin, end]
// If groupId=0 then return contests in any groups meets the above conditions
func GetContestsByUser(ctx context.Context, username string, begin, end time.Time, groupId int) []Contest {
	contests := make([]Contest, 0)
	query := `
SELECT id, name, start_time, duration FROM contest
WHERE start_time BETWEEN ? AND ?
AND id IN
(
    SELECT DISTINCT contest_team_rel.contest_id
    FROM contest_group_rel, contest_team_rel, team_user_rel
    WHERE contest_team_rel.contest_id = contest_group_rel.contest_id
    AND team_user_rel.team_id = contest_team_rel.team_id
    AND username = ? AND group_id BETWEEN ? AND ?
)
ORDER BY start_time DESC`
	args := []interface{}{begin, end, username}
	if groupId == 0 {
		args = append(args, 0, int(1e9))
	} else {
		args = append(args, groupId, groupId)
	}
	mustSelect(ctx, &contests, query, args...)
	mp := make(map[int]int)
	for i, c := range contests {
		mp[c.Id] = i
		if len(c.Problems) == 0 {
			c.Problems = make([]Problem, 0)
		}
		if len(c.Groups) == 0 {
			c.Groups = make([]int, 0)
		}
	}

	problems := make([]Problem, 0)
	query = `
SELECT * FROM contest_problem WHERE contest_id IN
(
    SELECT id FROM contest
    WHERE start_time BETWEEN ? AND ?
    AND id IN
    (
        SELECT DISTINCT contest_team_rel.contest_id
        FROM contest_group_rel, contest_team_rel, team_user_rel
        WHERE contest_team_rel.contest_id = contest_group_rel.contest_id
        AND team_user_rel.team_id = contest_team_rel.team_id
        AND username = ? AND group_id BETWEEN ? AND ?
    )
)
ORDER BY contest_id,` + "`index`"
	mustSelect(ctx, &problems, query, args...)
	for _, p := range problems {
		i := mp[p.ContestId]
		contests[i].Problems = append(contests[i].Problems, p)
	}
	return contests
}

// GetContestById get contest full info (with problems)
func GetContestById(ctx context.Context, id int) Contest {
	var c Contest
	mustGet(ctx, &c, "SELECT * FROM contest WHERE id=?", id)
	c.Problems = make([]Problem, 0)
	mustSelect(ctx, &c.Problems, "SELECT * FROM contest_problem WHERE contest_id = ?", id)
	sort.SliceStable(c.Problems, func(i, j int) bool {
		if len(c.Problems[i].Index) == len(c.Problems[j].Index) {
			return c.Problems[i].Index < c.Problems[j].Index
		}
		return len(c.Problems[i].Index) < len(c.Problems[j].Index)
	})
	return c
}

// AddContest return the new Contest with Contest.Id
func AddContest(ctx context.Context, c Contest) Contest {
	query := `INSERT INTO contest(oj_id, cid, name, start_time, duration, max_solved, participants)
VALUES(:oj_id, :cid, :name, :start_time, :duration, :max_solved, :participants)`
	tx := instance.MustBeginTx(ctx, nil)
	defer tx.Rollback()
	res := mustNamedExecTx(tx, ctx, query, c.dbType())
	id, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}
	c.Id = int(id)
	if len(c.Problems) > 0 {
		for i := range c.Problems {
			c.Problems[i].ContestId = c.Id
		}
		mustNamedExecTx(tx, ctx, addContestProblemSQL, c.Problems)
	}
	if len(c.Groups) > 0 {
		groups := make([]ContestGroupRel, 0)
		for _, x := range c.Groups {
			groups = append(groups, ContestGroupRel{
				GroupId:   x,
				ContestId: c.Id,
			})
		}
		mustNamedExecTx(tx, ctx, addContestGroupRelSQL, groups)
	}
	if len(c.Teams) > 0 {
		teams := make([]ContestTeamRel, 0)
		for _, x := range c.Teams {
			teams = append(teams, ContestTeamRel{
				ContestId: c.Id,
				TeamId:    x,
			})
		}
		mustNamedExecTx(tx, ctx, addContestTeamRelSQL, teams)
	}
	mustCommit(tx)
	return c
}

func UpdContest(ctx context.Context, c Contest) {
	tx := instance.MustBeginTx(ctx, nil)
	defer tx.Rollback()
	query := `UPDATE contest
SET oj_id=:oj_id, cid=:cid, name=:name, start_time=:start_time,
duration=:duration, max_solved=:max_solved, participants=:participants
WHERE id=:id`
	mustNamedExecTx(tx, ctx, query, c.dbType())
	mustExecTx(tx, ctx, "DELETE FROM contest_problem WHERE contest_id=?", c.Id)
	mustExecTx(tx, ctx, "DELETE FROM contest_group_rel WHERE contest_id=?", c.Id)
	mustExecTx(tx, ctx, "DELETE FROM contest_team_rel WHERE contest_id=?", c.Id)
	if len(c.Problems) > 0 {
		for i := range c.Problems {
			c.Problems[i].ContestId = c.Id
		}
		mustNamedExecTx(tx, ctx, addContestProblemSQL, c.Problems)
	}
	if len(c.Groups) > 0 {
		groups := make([]ContestGroupRel, 0)
		for i := range c.Groups {
			groups = append(groups, ContestGroupRel{
				GroupId:   c.Groups[i],
				ContestId: c.Id,
			})
		}
		mustNamedExecTx(tx, ctx, addContestGroupRelSQL, groups)
	}
	if len(c.Teams) > 0 {
		teams := make([]ContestTeamRel, 0)
		for _, x := range c.Teams {
			teams = append(teams, ContestTeamRel{
				ContestId: c.Id,
				TeamId:    x,
			})
		}
		mustNamedExecTx(tx, ctx, addContestTeamRelSQL, teams)
	}
	mustCommit(tx)
}

func DelContest(ctx context.Context, contestId int) {
	query := "DELETE FROM contest WHERE id=?"
	mustExec(ctx, query, contestId)
}

// PullContest only refresh the basic-info and problems of a specific contest
func PullContest(ctx context.Context, c Contest) {
	tx := instance.MustBeginTx(ctx, nil)
	defer tx.Rollback()
	query := `UPDATE contest
SET oj_id=:oj_id, cid=:cid, name=:name, start_time=:start_time,
duration=:duration, max_solved=:max_solved, participants=:participants
WHERE id=:id`
	mustNamedExecTx(tx, ctx, query, c.dbType())
	mustExecTx(tx, ctx, "DELETE FROM contest_problem WHERE contest_id=?", c.Id)
	if len(c.Problems) > 0 {
		for i := range c.Problems {
			c.Problems[i].ContestId = c.Id
		}
		mustNamedExecTx(tx, ctx, addContestProblemSQL, c.Problems)
	}
	mustCommit(tx)
}

type ContestsOverviewCell struct {
	Username string `json:"username" db:"username"`
	Nickname string `json:"nickname" db:"nickname"`
	Solved   int    `json:"solved" db:"solved"`
	Upsolved int    `json:"upsolved" db:"upsolved"`
}

type ContestsOverview struct {
	GroupId   int                    `json:"group_id"`
	GroupName string                 `json:"group_name"`
	Users     []ContestsOverviewCell `json:"users"`
}

func getContestsOverviewByCells(ctx context.Context, cells []ContestsOverviewCell) []ContestsOverview {
	groups := GetOfficialUsers(ctx, true)
	grp := make(map[int]*ContestsOverview)
	grpId := make(map[string]int)
	for _, x := range groups {
		grp[x.GroupId] = &ContestsOverview{
			GroupId:   x.GroupId,
			GroupName: x.GroupName,
			Users:     make([]ContestsOverviewCell, 0),
		}
		for _, u := range x.Users {
			grpId[u.Username] = x.GroupId
		}
	}
	for _, x := range cells {
		i := grpId[x.Username]
		grp[i].Users = append(grp[i].Users, x)
	}
	ret := make([]ContestsOverview, 0)
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

func GetContestsOverview(ctx context.Context, begin, end time.Time) []ContestsOverview {
	query := `
SELECT username, nickname,
(
    SELECT COUNT(DISTINCT submission.pid, submission.oj_id)
    FROM submission, contest_problem, contest
    WHERE submission.oj_id = contest_problem.oj_id AND submission.pid = contest_problem.pid
      AND contest.id = contest_problem.contest_id AND is_accepted=true
      AND create_time BETWEEN start_time AND DATE_ADD(start_time, interval duration MINUTE )
      AND user.username = submission.username
      AND contest_id IN (SELECT id FROM contest WHERE start_time BETWEEN ? AND ?)
) solved,
(
    SELECT COUNT(DISTINCT submission.pid, submission.oj_id)
    FROM submission, contest_problem
    WHERE submission.oj_id = contest_problem.oj_id AND submission.pid = contest_problem.pid
      AND is_accepted=true AND user.username = submission.username
      AND contest_id IN (SELECT id FROM contest WHERE start_time BETWEEN ? AND ?)
) upsolved
FROM user
WHERE is_enable AND username IN
(
    SELECT username FROM team_user_rel, contest_team_rel
    WHERE team_user_rel.team_id = contest_team_rel.team_id
      AND contest_id IN (SELECT id FROM contest WHERE start_time BETWEEN ? AND ?)
)
ORDER BY upsolved DESC, solved DESC, username`

	var args []interface{}
	for i := 0; i < 3; i++ {
		args = append(args, begin)
		args = append(args, end)
	}
	cells := make([]ContestsOverviewCell, 0)
	mustSelect(ctx, &cells, query, args...)
	return getContestsOverviewByCells(ctx, cells)
}

func GetContestsOverviewByGroup(ctx context.Context, id int, begin, end time.Time) []ContestsOverview {
	query := `
SELECT username, nickname,
(
    SELECT COUNT(DISTINCT submission.pid, submission.oj_id)
    FROM submission, contest_problem, contest
    WHERE submission.oj_id = contest_problem.oj_id AND submission.pid = contest_problem.pid
      AND contest.id = contest_problem.contest_id AND is_accepted=true
      AND create_time BETWEEN start_time AND DATE_ADD(start_time, interval duration MINUTE )
      AND user.username = submission.username
      AND contest_id IN (SELECT contest_id FROM contest_group_rel, contest WHERE contest_id = contest.id AND group_id = ? AND start_time BETWEEN ? AND ?)
) solved,
(
    SELECT COUNT(DISTINCT submission.pid, submission.oj_id)
    FROM submission, contest_problem
    WHERE submission.oj_id = contest_problem.oj_id AND submission.pid = contest_problem.pid
      AND is_accepted=true AND user.username = submission.username
      AND contest_id IN (SELECT contest_id FROM contest_group_rel, contest WHERE contest_id = contest.id AND group_id = ? AND start_time BETWEEN ? AND ?)
) upsolved
FROM user
WHERE is_enable AND username IN
(
    SELECT username FROM team_user_rel, contest_team_rel
    WHERE team_user_rel.team_id = contest_team_rel.team_id
      AND contest_id IN (SELECT contest_id FROM contest_group_rel, contest WHERE contest_id = contest.id AND group_id = ? AND start_time BETWEEN ? AND ?)
)
ORDER BY upsolved DESC, solved DESC, username`

	var args []interface{}
	for i := 0; i < 3; i++ {
		args = append(args, id)
		args = append(args, begin)
		args = append(args, end)
	}
	cells := make([]ContestsOverviewCell, 0)
	mustSelect(ctx, &cells, query, args...)
	return getContestsOverviewByCells(ctx, cells)
}
