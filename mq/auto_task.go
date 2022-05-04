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
	AddTask(runner, "*/30 * * * *", refreshSubmission)
	AddTask(runner, "10/30 * * * *", refreshRatingCodeforces)
	runner.Start()
}

func AddTask(taskRunner *cron.Cron, spec string, cmd func()) {
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
	ExecTask(Topic(1), SubmissionTask(nil, 0, []string{"5H0hEjEiuF"}, 100000))
}

func refreshRatingCodeforces() {
	refreshRating(db.GetOJByName("codeforces").OjId)
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
