package db

import (
	"context"
	"sort"
	"time"

	"zuccacm-server/utils"
)

type Submission struct {
	Id          int      `json:"id" db:"id"`
	Username    string   `json:"username" db:"username"`
	OjId        int      `json:"oj_id" db:"oj_id"`
	AccountOjId int      `json:"account_oj_id" db:"account_oj_id"`
	Sid         string   `json:"sid" db:"sid"`
	Pid         string   `json:"pid" db:"pid"`
	IsAccepted  bool     `json:"is_accepted" db:"is_accepted"`
	CreateTime  Datetime `json:"create_time" db:"create_time"`
}

type dbSubmission struct {
	Id          int       `json:"id" db:"id"`
	Username    string    `json:"username" db:"username"`
	OjId        int       `json:"oj_id" db:"oj_id"`
	AccountOjId int       `json:"account_oj_id" db:"account_oj_id"`
	Sid         string    `json:"sid" db:"sid"`
	Pid         string    `json:"pid" db:"pid"`
	IsAccepted  bool      `json:"is_accepted" db:"is_accepted"`
	CreateTime  time.Time `json:"create_time" db:"create_time"`
}

func (s *Submission) dbType() *dbSubmission {
	return &dbSubmission{
		Id:          s.Id,
		Username:    s.Username,
		OjId:        s.OjId,
		AccountOjId: s.AccountOjId,
		Sid:         s.Sid,
		Pid:         s.Pid,
		IsAccepted:  s.IsAccepted,
		CreateTime:  time.Time(s.CreateTime),
	}
}

func (s *dbSubmission) jsonType() *Submission {
	return &Submission{
		Id:          s.Id,
		Username:    s.Username,
		OjId:        s.OjId,
		AccountOjId: s.AccountOjId,
		Sid:         s.Sid,
		Pid:         s.Pid,
		IsAccepted:  s.IsAccepted,
		CreateTime:  Datetime(s.CreateTime),
	}
}

func AddSubmission(ctx context.Context, s []Submission) {
	accounts := GetAllAccounts(ctx)
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
		k := key{si.AccountOjId, si.Username}
		if _, ok := mp[k]; !ok {
			continue
		}
		si.Username = mp[k]
		data = append(data, *si.dbType())
	}
	query := `INSERT IGNORE INTO submission(username, oj_id, account_oj_id, sid, pid, is_accepted, create_time)
VALUES(:username, :oj_id, :account_oj_id, :sid, :pid, :is_accepted, :create_time)`
	tx := instance.MustBeginTx(ctx, nil)
	n := len(data)
	groupSize := 5000
	for i := 0; i < n; i += groupSize {
		mustNamedExecTx(tx, ctx, query, data[i:utils.Min(i+groupSize, n)])
	}
	mustCommit(tx)
}

// GetSubmissionsInContest return submissions from team_user in this contest
func GetSubmissionsInContest(ctx context.Context, contestId int) []Submission {
	query := `
SELECT submission.username AS username, is_accepted, create_time, submission.oj_id AS oj_id, submission.pid AS pid
FROM submission, contest_problem
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

func GetAcceptedSubmissionByUsername(ctx context.Context, username string, begin, end time.Time) []Submission {
	query := `SELECT min(create_time) AS create_time
FROM submission WHERE is_accepted AND username=?
GROUP BY oj_id, pid HAVING min(create_time) BETWEEN ? AND ?`
	ret := make([]Submission, 0)
	mustSelect(ctx, &ret, query, username, begin, end)
	return ret
}

func GetSubmissionByUsername(ctx context.Context, username string, begin, end time.Time) []Submission {
	query := "SELECT * FROM submission WHERE username=? AND create_time BETWEEN ? AND ?"
	ret := make([]Submission, 0)
	mustSelect(ctx, &ret, query, username, begin, end)
	return ret
}

type overviewCell struct {
	Username   string `json:"username"`
	Nickname   string `json:"nickname"`
	Solved     int    `json:"solved"`
	Submission int    `json:"submission"`
}

type overview struct {
	GroupId   int            `json:"group_id"`
	GroupName string         `json:"group_name"`
	Users     []overviewCell `json:"users"`
}

func GetOverview(ctx context.Context, begin, end time.Time) []overview {
	ret := make([]overview, 0)
	var groups []TeamGroup
	query := `
SELECT group_id, group_name
FROM official_user
WHERE is_enable
GROUP BY group_id HAVING COUNT(*) > 0`
	mustSelect(ctx, &groups, query)
	mp := make(map[int]int)
	for i, g := range groups {
		ret = append(ret, overview{
			GroupId:   g.GroupId,
			GroupName: g.GroupName,
			Users:     make([]overviewCell, 0),
		})
		mp[g.GroupId] = i
	}
	var data []struct {
		Username   string `db:"username"`
		Nickname   string `db:"nickname"`
		GroupId    int    `db:"group_id"`
		Solved     int    `db:"solved"`
		Submission int    `db:"submission"`
	}
	query = `
SELECT username, nickname, group_id,
(
    SELECT COUNT(*) FROM
    (
         SELECT MIN(create_time) FROM submission
         WHERE is_accepted AND username = official_user.username
         GROUP BY oj_id, pid HAVING MIN(create_time) BETWEEN ? AND ?
    ) tmp
) solved,
(
    SELECT COUNT(*) FROM submission
    WHERE username = official_user.username AND create_time BETWEEN ? AND ?
) submission
FROM official_user
WHERE is_enable`
	var args []interface{}
	for i := 0; i < 2; i++ {
		args = append(args, begin)
		args = append(args, end)
	}
	mustSelect(ctx, &data, query, args...)
	for _, u := range data {
		i := mp[u.GroupId]
		ret[i].Users = append(ret[i].Users, overviewCell{
			Username:   u.Username,
			Nickname:   u.Nickname,
			Solved:     u.Solved,
			Submission: u.Submission,
		})
	}
	for k := range ret {
		sort.SliceStable(ret[k].Users, func(i, j int) bool {
			x, y := ret[k].Users[i], ret[k].Users[j]
			if x.Solved != y.Solved {
				return x.Solved > y.Solved
			} else if x.Submission != y.Submission {
				return x.Submission > y.Submission
			} else {
				return x.Username < y.Username
			}
		})
	}
	sort.SliceStable(ret, func(i, j int) bool {
		return ret[i].GroupName > ret[j].GroupName
	})
	return ret
}
