package handler

import (
	"net/http"

	"zuccacm-server/db"
	"zuccacm-server/mq"
)

var submissionRouter = Router.PathPrefix("/submission").Subrouter()

func init() {
	submissionRouter.HandleFunc("/add", addSubmissions).Methods("POST")
	submissionRouter.HandleFunc("/refresh_all", adminOnly(refreshAllSubmission)).Methods("POST")
	submissionRouter.HandleFunc("/refresh", userSelfOrAdminOnly(refreshSubmission)).Methods("POST")
}

// addSubmissions is an api only for spiderhost
func addSubmissions(w http.ResponseWriter, r *http.Request) {
	submissions := make([]struct {
		Username   string      `json:"username"`
		OJ         string      `json:"oj"`
		Sid        string      `json:"sid"`
		Pid        string      `json:"pid"`
		IsAccepted bool        `json:"is_accepted"`
		CreateTime db.Datetime `json:"create_time"`
	}, 0)
	decodeParamVar(r, &submissions)
	ctx := r.Context()

	oj := db.ParseOJMap(db.GetAllOJ(ctx))
	data := make([]db.Submission, 0)
	for _, s := range submissions {
		data = append(data, db.Submission{
			Username:   s.Username,
			OjId:       oj[s.OJ],
			Sid:        s.Sid,
			Pid:        s.Pid,
			IsAccepted: s.IsAccepted,
			CreateTime: s.CreateTime,
		})
	}
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
