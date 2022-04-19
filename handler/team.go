package handler

import (
	"net/http"

	"zuccacm-server/db"
)

var teamRouter = Router.PathPrefix("/team").Subrouter()

func init() {
	teamRouter.HandleFunc("/add", adminOnly(addTeam)).Methods("POST")
}

func addTeam(w http.ResponseWriter, r *http.Request) {
	var args struct {
		Name  string   `json:"name"`
		Users []string `json:"users"`
	}
	decodeParamVar(r, &args)
	team := db.Team{
		Name:     args.Name,
		IsEnable: true,
		IsSelf:   false,
	}
	for _, s := range args.Users {
		team.Users = append(team.Users, db.User{Username: s})
	}
	db.AddTeam(r.Context(), team)
	msgResponse(w, http.StatusOK, "添加队伍成功")
}
