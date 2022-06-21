package handler

import (
	"errors"
	"fmt"
	"unicode"

	log "github.com/sirupsen/logrus"
)

func getProblemURL(oj, pid string) string {
	switch oj {
	case "codeforces":
		return cfProblemURL(pid)
	case "poj":
		return pojProblemURL(pid)
	case "nowcoder":
		return fmt.Sprintf("https://ac.nowcoder.com/acm/problem/%s", pid)
	default:
		log.WithFields(log.Fields{
			"oj":  oj,
			"pid": pid,
		}).Error("can't get problem url: not supported oj")
		return ""
	}
}

func cfProblemURL(pid string) string {
	sgu := "acmsguru"
	if len(pid) > len(sgu) && pid[:len(sgu)] == sgu {
		return fmt.Sprintf("https://codeforces.com/problemsets/acmsguru/problem/99999/%s", pid[len(sgu):])
	}
	cid, index, err := decodePid(pid)
	if err != nil {
		log.WithField("pid", pid).Error(err)
	}
	if len(cid) >= 6 {
		return fmt.Sprintf("https://codeforces.com/gym/%s/problem/%s", cid, index)
	} else {
		return fmt.Sprintf("https://codeforces.com/contest/%s/problem/%s", cid, index)
	}
}

// decodePid will split '1670A1' to '{cid: 1670, index:A1}'
func decodePid(pid string) (cid, index string, err error) {
	for i, c := range pid {
		if !unicode.IsDigit(c) {
			cid = pid[:i]
			index = pid[i:]
			return
		}
	}
	err = errors.New(fmt.Sprintf("decode pid %s failed", pid))
	return
}

func pojProblemURL(pid string) string {
	return fmt.Sprintf("http://poj.org/problem?id=%s", pid)
}
