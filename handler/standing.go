package handler

import "zuccacm-server/db"

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
