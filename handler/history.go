package handler

import (
	"net/http"
	"time"
	"zuccacm-server/db"
)

var historyRouter = Router.PathPrefix("/history").Subrouter()

func init() {
	Router.HandleFunc("/historys", getHistorys).Methods("GET")
	historyRouter.HandleFunc("/add", adminOnly(addHistory)).Methods("POST")
}
func getHistorys(w http.ResponseWriter, r *http.Request) {
	isEnable := getParamBool(r, "is_enable", false)
	historys := db.GetHistorys(r.Context(), isEnable)
	dataResponse(w, historys)
}
func addHistory(w http.ResponseWriter, r *http.Request) {
	var args struct {
		Name       string    `json:"historyname" db:"name"`
		Start_time time.Time `json:"start_time" db:"start_time"`
		End_time   time.Time `json:"end_time" db:"end_time"`
	}
	decodeParamVar(r, &args)
	history := db.History{
		Name:       args.Name,
		Start_time: args.Start_time,
		End_time:   args.End_time,
	}
	db.AddHistory(r.Context(), history)
	msgResponse(w, http.StatusOK, "添加事件成功")
}
