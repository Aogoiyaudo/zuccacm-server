package handler

import (
	"net/http"

	"zuccacm-server/db"
	"zuccacm-server/mq"
)

var submissionRouter = Router.PathPrefix("/submission").Subrouter()

func init() {
	submissionRouter.HandleFunc("/add", addSubmissions).Methods("POST")
	submissionRouter.HandleFunc("/refresh", refreshSubmission).Methods("POST")
}

// addSubmissions is an api only for spiderhost
func addSubmissions(w http.ResponseWriter, r *http.Request) {
	var submissions []db.Submission
	decodeParamVar(r, &submissions)
	db.AddSubmission(r.Context(), submissions)
	msgResponse(w, http.StatusOK, "add submissions success")
}

// refreshSubmission fetch new submissions from users or groups (such like codeforces-group)
// default submission-count is 100 and 1000 respectively
func refreshSubmission(w http.ResponseWriter, r *http.Request) {
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
	msgResponse(w, http.StatusOK, "任务已创建：获取提交")
}
