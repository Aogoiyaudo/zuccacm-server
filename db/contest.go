package db

import (
	"context"
	"time"
)

type ContestGroup struct {
	Id       int    `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	IsEnable bool   `json:"is_enable" db:"is_enable"`
}

type Problem struct {
	ContestId int    `json:"contest_id" db:"contest_id"`
	OjId      int    `json:"oj_id" db:"oj_id"`
	Pid       string `json:"pid" db:"pid"`
	Index     string `json:"index" db:"index"`
}

type Contest struct {
	Id           int       `json:"id" db:"id"`
	OjId         int       `json:"oj_id,omitempty" db:"oj_id"`
	Cid          string    `json:"cid,omitempty" db:"cid"`
	Name         string    `json:"name,omitempty" db:"name"`
	StartTime    Datetime  `json:"start_time,omitempty" db:"start_time"`
	Duration     int       `json:"duration,omitempty" db:"duration"`
	MaxSolved    int       `json:"max_solved,omitempty" db:"max_solved"`
	Participants int       `json:"participants,omitempty" db:"participants"`
	Problems     []Problem `json:"problems"`
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
	}
}

func (c *dbContest) jsonType() *Contest {
	return &Contest{
		Id:           c.Id,
		OjId:         c.OjId,
		Cid:          c.Cid,
		Name:         c.Name,
		StartTime:    Datetime(c.StartTime),
		Duration:     c.Duration,
		MaxSolved:    c.MaxSolved,
		Participants: c.Participants,
		Problems:     c.Problems,
	}
}

// GetContestGroups return all groups if isEnable=false
func GetContestGroups(ctx context.Context, isEnable bool) []ContestGroup {
	query := "SELECT * FROM contest_group"
	if isEnable {
		query += " WHERE is_enable=true"
	}
	ret := make([]ContestGroup, 0)
	mustSelect(ctx, &ret, query)
	return ret
}

// GetContests only get contests basic info (without problems)
func GetContests(ctx context.Context, groupId int, begin, end time.Time) []Contest {
	query := `SELECT * FROM contest
WHERE start_time BETWEEN ? AND ?
AND id IN (SELECT contest_id FROM contest_group_rel WHERE group_id = ?)
ORDER BY start_time DESC`
	ret := make([]Contest, 0)
	mustSelect(ctx, &ret, query, begin, end, groupId)
	for i := range ret {
		ret[i].Problems = make([]Problem, 0)
	}
	return ret
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
	mustSelect(ctx, &c.Problems, "SELECT * FROM contest_problem WHERE contest_id = ? ORDER BY `index`", id)
	return c
}

func AddContest(ctx context.Context, c Contest) {
	tx := instance.MustBeginTx(ctx, nil)
	defer tx.Rollback()
	mustNamedExecTx(tx, ctx, "INSERT INTO contest(oj_id, cid, name, start_time, duration, max_solved, participants) VALUES(:oj_id, :cid, :name, :start_time, :duration, :max_solved, :participants)", c.dbType())
	if len(c.Problems) > 0 {
		mustNamedExecTx(tx, ctx, addContestProblemSQL, c.Problems)
	}
	mustCommit(tx)
}

func UpdContest(ctx context.Context, c Contest) {
	tx := instance.MustBeginTx(ctx, nil)
	defer tx.Rollback()
	mustNamedExecTx(tx, ctx, "UPDATE contest SET oj_id=:oj_id, cid=:cid, name=:name, start_time=:start_time, duration=:duration, max_solved=:max_solved, participants=:participants WHERE id=:id", c.dbType())
	mustNamedExecTx(tx, ctx, "DELETE FROM contest_problem WHERE contest_id=:id", c)
	if len(c.Problems) > 0 {
		mustNamedExecTx(tx, ctx, addContestProblemSQL, c.Problems)
	}
	mustCommit(tx)
}

type Overview struct {
	Username string `json:"username" db:"username"`
	Nickname string `json:"nickname" db:"nickname"`
	Solved   int    `json:"solved" db:"solved"`
	Upsolved int    `json:"upsolved" db:"upsolved"`
}

func GetContestGroupOverview(ctx context.Context, id int, begin, end time.Time) []Overview {
	query := `SELECT
username, nickname,
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
WHERE username IN
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
	ret := make([]Overview, 0)
	mustSelect(ctx, &ret, query, args...)
	return ret
}
