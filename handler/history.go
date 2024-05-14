package handler

import (
	"net/http"
	"strconv"
	"time"
	"zuccacm-server/db"
)

var historyRouter = Router.PathPrefix("/history").Subrouter()

func init() {
	Router.HandleFunc("/historys", getHistorys).Methods("GET")
	Router.HandleFunc("/history/{historyid}", getHistory).Methods("GET")
	historyRouter.HandleFunc("/add", adminOnly(addHistory)).Methods("POST")
	Router.HandleFunc("/history_edit", adminOnly(updHistory)).Methods("POST")
}
func getHistory(w http.ResponseWriter, r *http.Request) {
	id := getParamURL(r, "historyid")
	di, _ := strconv.Atoi(id)
	history := db.MustGetHistory(r.Context(), di)
	dataResponse(w, history)
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
		Md:         "",
	}
	db.AddHistory(r.Context(), history)
	msgResponse(w, http.StatusOK, "添加事件成功")
}
func updHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var args struct {
		Id int    `json:"id"`
		Md string `json:"md"`
	}
	decodeParamVar(r, &args)
	di := args.Id
	db.UpdHistory(ctx, di, args.Md)
	msgResponse(w, http.StatusOK, "更新信息成功")
}
