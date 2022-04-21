package db

import (
	"context"
	"time"
)

type Submission struct {
	Id         int      `json:"id" db:"id"`
	Username   string   `json:"username" db:"username"`
	OjId       int      `json:"oj_id" db:"oj_id"`
	Sid        string   `json:"sid" db:"sid"`
	Pid        string   `json:"pid" db:"pid"`
	IsAccepted bool     `json:"is_accepted" db:"is_accepted"`
	CreateTime Datetime `json:"create_time" db:"create_time"`
}

type dbSubmission struct {
	Id         int       `json:"id" db:"id"`
	Username   string    `json:"username" db:"username"`
	OjId       int       `json:"oj_id" db:"oj_id"`
	Sid        string    `json:"sid" db:"sid"`
	Pid        string    `json:"pid" db:"pid"`
	IsAccepted bool      `json:"is_accepted" db:"is_accepted"`
	CreateTime time.Time `json:"create_time" db:"create_time"`
}

func (s *Submission) dbType() *dbSubmission {
	return &dbSubmission{
		Id:         s.Id,
		Username:   s.Username,
		OjId:       s.OjId,
		Sid:        s.Sid,
		Pid:        s.Pid,
		IsAccepted: s.IsAccepted,
		CreateTime: time.Time(s.CreateTime),
	}
}

func (s *dbSubmission) jsonType() *Submission {
	return &Submission{
		Id:         s.Id,
		Username:   s.Username,
		OjId:       s.OjId,
		Sid:        s.Sid,
		Pid:        s.Pid,
		IsAccepted: s.IsAccepted,
		CreateTime: Datetime(s.CreateTime),
	}
}

func AddSubmission(ctx context.Context, s []Submission) {
	accounts := GetEnableAccount(ctx)
	type key struct {
		oj      int
		account string
	}
	mp := make(map[key]string)
	for _, account := range accounts {
		mp[key{account.OjId, account.Account}] = account.Username
	}
	data := make([]dbSubmission, 0)
	for _, si := range s {
		k := key{si.OjId, si.Username}
		if _, ok := mp[k]; !ok {
			continue
		}
		si.Username = mp[k]
		data = append(data, *si.dbType())
	}
	query := `INSERT IGNORE INTO submission(username, oj_id, sid, pid, is_accepted, create_time)
VALUES(:username, :oj_id, :sid, :pid, :is_accepted, :create_time)`
	mustNamedExec(ctx, query, data)
}

// GetSubmissionsInContest return submissions from team_user in this contest
func GetSubmissionsInContest(ctx context.Context, contestId int) []Submission {
	query := `
SELECT submission.username AS username, is_accepted, create_time, submission.oj_id AS oj_id, submission.pid AS pid
FROM submission,
     contest_problem
WHERE submission.oj_id = contest_problem.oj_id
  AND submission.pid = contest_problem.pid
  AND contest_problem.contest_id = ?
  AND username IN (SELECT DISTINCT username
                   FROM team_user_rel,
                        contest_team_rel
                   WHERE team_user_rel.team_id = contest_team_rel.team_id
                     AND contest_id = ?)
ORDER BY create_time`
	ret := make([]Submission, 0)
	mustSelect(ctx, &ret, query, contestId, contestId)
	return ret
}

func GetSubmissionByUsername(ctx context.Context, username string, begin, end time.Time) []Submission {
	query := `SELECT * FROM submission
WHERE is_accepted AND username=?
AND create_time BETWEEN ? AND ?
ORDER BY create_time`
	ret := make([]Submission, 0)
	mustSelect(ctx, &ret, query, username, begin, end)
	return ret
}
