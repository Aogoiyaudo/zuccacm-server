package handler

import (
	"net/http"
	"time"
	"zuccacm-server/db"
)

var eventRouter = Router.PathPrefix("/event").Subrouter()

func init() {
	Router.HandleFunc("/events", getEvents).Methods("GET")
	eventRouter.HandleFunc("/add", adminOnly(addEvent)).Methods("POST")
}
func getEvents(w http.ResponseWriter, r *http.Request) {
	isEnable := getParamBool(r, "is_enable", false)
	events := db.GetEvents(r.Context(), isEnable)
	dataResponse(w, events)
}
func addEvent(w http.ResponseWriter, r *http.Request) {
	var args struct {
		Name       string    `json:"name" db:"name"`
		Start_time time.Time `json:"start_time" db:"start_time"`
		End_time   time.Time `json:"end_time" db:"end_time"`
	}
	decodeParamVar(r, &args)
	event := db.Event{
		Name:       args.Name,
		Start_time: args.Start_time,
		End_time:   args.End_time,
	}
	db.AddEvent(r.Context(), event)
	msgResponse(w, http.StatusOK, "添加活动成功")
}
