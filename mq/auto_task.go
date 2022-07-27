package mq

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"

	"zuccacm-server/db"
)

func init() {
	runner := cron.New()
	AddTask(runner, "40 * * * *", refreshSubmission)
	AddTask(runner, "10 * * * *", refreshRatingCodeforces)
	AddTask(runner, "20 4 * * *", refreshGroupSubmission)
	runner.Start()
}

func AddTask(taskRunner *cron.Cron, spec string, cmd func()) {
	if log.GetLevel() == log.DebugLevel {
		return
	}
	if _, err := taskRunner.AddFunc(spec, cmd); err != nil {
		log.Fatal(err)
	}
}

func refreshSubmission() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ojAccounts := db.GetAllAccountsGroupByOJ(ctx)
	for _, x := range ojAccounts {
		username := make([]string, 0)
		for _, u := range x.Accounts {
			if u.Account != "" {
				username = append(username, u.Account)
			}
		}
		if len(username) > 0 {
			ExecTask(Topic(x.OjId), SubmissionTask(username, 1000, nil, 0))
		}
	}
}

func refreshGroupSubmission() {
	codeforces := db.GetOJByName("codeforces").OjId
	ExecTask(Topic(codeforces), SubmissionTask(nil, 0, []string{"5H0hEjEiuF"}, 3000))
}

func refreshRatingCodeforces() {
	codeforces := db.GetOJByName("codeforces").OjId
	refreshRating(codeforces)
}

func refreshRating(ojId int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	accounts := db.GetAccountsByOJ(ctx, ojId)
	username := make([]string, 0)
	for _, x := range accounts {
		if x.Account != "" {
			username = append(username, x.Account)
		}
	}
	if len(username) > 0 {
		ExecTask(Topic(1), RatingTask(username))
	}
}
