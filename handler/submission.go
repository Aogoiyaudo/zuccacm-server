package handler

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"zuccacm-server/db"
	"zuccacm-server/mq"
)

var submissionRouter = Router.PathPrefix("/submission").Subrouter()

func init() {
	submissionRouter.HandleFunc("/add", addSubmissions).Methods("POST")
	submissionRouter.HandleFunc("/refresh_all", adminOnly(refreshAllSubmission)).Methods("POST")
	submissionRouter.HandleFunc("/refresh", userSelfOrAdminOnly(refreshSubmission)).Methods("POST")

	Router.HandleFunc("/overview", submissionOverview).Methods("GET")
}

// addSubmissions is an api only for spiderhost
func addSubmissions(w http.ResponseWriter, r *http.Request) {
	type submission struct {
		Username   string      `json:"username"`
		OJ         string      `json:"oj"`
		Sid        string      `json:"sid"`
		Pid        string      `json:"pid"`
		IsAccepted bool        `json:"is_accepted"`
		CreateTime db.Datetime `json:"create_time"`
	}
	args := struct {
		AccountOjId int          `json:"account_oj_id"`
		Submissions []submission `json:"submissions"`
	}{}
	decodeParamVar(r, &args)
	ctx := r.Context()
	log.WithFields(log.Fields{
		"account_oj_id": args.AccountOjId,
		"submission":    args.Submissions[0],
	}).Debug()

	oj := db.OJMapStoI(db.GetAllOJ(ctx))
	data := make([]db.Submission, 0)
	for _, s := range args.Submissions {
		data = append(data, db.Submission{
			Username:    s.Username,
			OjId:        oj[s.OJ],
			AccountOjId: args.AccountOjId,
			Sid:         s.Sid,
			Pid:         s.Pid,
			IsAccepted:  s.IsAccepted,
			CreateTime:  s.CreateTime,
		})
	}
	log.Debug(data[0])
	db.AddSubmission(ctx, data)
	msgResponse(w, http.StatusOK, "add submissions success")
}

// refreshAllSubmission fetch new submissions from users or groups (such like codeforces-group)
// default submission-count is 100 and 1000 respectively
func refreshAllSubmission(w http.ResponseWriter, r *http.Request) {
	args := struct {
		OjId       int      `json:"oj_id"`
		Username   []string `json:"username"`
		Count      int      `json:"count"`
		Group      []string `json:"group"`
		GroupCount int      `json:"group_count"`
	}{}
	args.Count = 100
	args.GroupCount = 1000
	decodeParamVar(r, &args)
	mq.ExecTask(mq.Topic(args.OjId), mq.SubmissionTask(args.Username, args.Count, args.Group, args.GroupCount))
	msgResponse(w, http.StatusOK, "任务已创建：刷新提交")
}

// refreshSubmission fetch the latest count submissions of a specific user
func refreshSubmission(w http.ResponseWriter, r *http.Request) {
	args := struct {
		OjId     int    `json:"oj_id"`
		Username string `json:"username"`
		Count    int    `json:"count"`
	}{}
	args.Count = 1e9
	decodeParamVar(r, &args)
	account := db.GetAccount(r.Context(), args.Username, args.OjId)
	mq.ExecTask(mq.Topic(args.OjId), mq.SubmissionTask([]string{account}, args.Count, nil, 0))
	msgResponse(w, http.StatusOK, "任务已创建：刷新提交")
}

func submissionOverview(w http.ResponseWriter, r *http.Request) {
	begin, end := getParamDateInterval(r)
	data := db.GetOverview(r.Context(), begin, end)
	dataResponse(w, data)
}
