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
	runner.Start()
}

func refreshSubmission() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	ojAccounts := db.GetAllAccounts(ctx)
	for _, x := range ojAccounts {
		username := make([]string, 0)
		for _, u := range x.Accounts {
			username = append(username, u.Account)
		}
		ExecTask(Topic(x.OjId), SubmissionTask(username, 100, nil, 0))
	}
}
