package handler

import (
	"net/http"
	"time"
	"zuccacm-server/db"
)

var xcpcRouter = Router.PathPrefix("/xcpc").Subrouter()

func init() {
	Router.HandleFunc("/xcpcs", getXcpcs).Methods("GET")
	xcpcRouter.HandleFunc("/add", adminOnly(addXcpc)).Methods("POST")
}
func getXcpcs(w http.ResponseWriter, r *http.Request) {
	xcpcs := db.GetXcpcs(r.Context())
	dataResponse(w, xcpcs)
}
func addXcpc(w http.ResponseWriter, r *http.Request) {
	var args struct {
		Name string    `json:"name" db:"name"`
		Date time.Time `json:"date" db:"date"`
	}
	decodeParamVar(r, &args)
	xcpc := db.Xcpc{
		Name: args.Name,
		Date: args.Date,
	}
	db.AddXcpc(r.Context(), xcpc)
	msgResponse(w, http.StatusOK, "添加奖项成功")
}
