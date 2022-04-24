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
	if _, err := runner.AddFunc("*/30 * * * *", refreshSubmission); err != nil {
		log.Fatal(err)
	}
	if _, err := runner.AddFunc("23/30 * * * *", refreshRating); err != nil {
		log.Fatal(err)
	}
	runner.Start()
}

func refreshSubmission() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ojAccounts := db.GetAllAccountsGroupByOJ(ctx)
	for _, x := range ojAccounts {
		username := make([]string, 0)
		for _, u := range x.Accounts {
			username = append(username, u.Account)
		}
		if len(username) > 0 {
			ExecTask(Topic(x.OjId), SubmissionTask(username, 100, nil, 0))
		}
	}
}

func refreshRating() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	accounts := db.GetAccountsByOJ(ctx, 1)
	username := make([]string, 0)
	for _, x := range accounts {
		username = append(username, x.Account)
	}
	if len(username) > 0 {
		ExecTask(Topic(1), RatingTask(username))
	}
}
