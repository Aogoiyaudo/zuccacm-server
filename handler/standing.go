package handler

import (
	"sort"

	"zuccacm-server/db"
)

const defaultAcceptedTime = -1000000000

type submissionInfo struct {
	IsAccepted bool        `json:"is_accepted"`
	CreateTime db.Datetime `json:"create_time"`
}

type problemResult struct {
	AcceptedTime int              `json:"accepted_time"`
	Dirt         int              `json:"dirt"`
	Submissions  []submissionInfo `json:"submissions"`
}

func (x problemResult) less(y problemResult) bool {
	if x.AcceptedTime != -1 && y.AcceptedTime == -1 {
		return false
	} else if x.AcceptedTime == -1 && y.AcceptedTime != -1 {
		return true
	}
	if x.AcceptedTime == -1 {
		return x.Dirt < y.Dirt
	} else {
		return x.Dirt > y.Dirt
	}
}

func maxResult(x, y problemResult) problemResult {
	if x.less(y) {
		return y
	} else {
		return x
	}
}

// calcProblemResult return results of a problem
// problemResult.AcceptedTime as follows:
// unsolved --- -1
// solved   --- [0, duration]
// upsolved --- duration + 1
func calcProblemResult(submissions []submissionInfo, startTime db.Datetime, duration int) problemResult {
	sort.SliceStable(submissions, func(i, j int) bool {
		return submissions[i].CreateTime.Unix() < submissions[j].CreateTime.Unix()
	})
	ret := problemResult{
		AcceptedTime: defaultAcceptedTime,
		Submissions:  make([]submissionInfo, 0),
	}
	for _, s := range submissions {
		ret.Submissions = append(ret.Submissions, s)
		if ret.AcceptedTime >= 0 {
			continue
		}
		if s.IsAccepted {
			ret.AcceptedTime = int((s.CreateTime.Unix() - startTime.Unix()) / 60)
		} else {
			ret.Dirt++
		}
	}
	if ret.AcceptedTime == defaultAcceptedTime {
		ret.AcceptedTime = -1
	} else if ret.AcceptedTime < 0 || ret.AcceptedTime > duration {
		ret.AcceptedTime = duration + 1
	}
	return ret
}
