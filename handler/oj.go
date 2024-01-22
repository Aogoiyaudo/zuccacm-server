package handler

import (
	"net/http"

	"zuccacm-server/db"
)

func init() {
	Router.HandleFunc("/oj", getOJ).Methods("GET")
	Router.HandleFunc("/oj/all", getAllOJ).Methods("GET")
}

func getOJ(w http.ResponseWriter, r *http.Request) {
	oj := db.GetAllEnableOJ(r.Context())
	dataResponse(w, oj)
}
func getAllOJ(w http.ResponseWriter, r *http.Request) {
	oj := db.GetAllOJ(r.Context())
	dataResponse(w, oj)
}
