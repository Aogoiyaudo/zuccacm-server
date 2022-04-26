package handler

import (
	"net/http"

	"zuccacm-server/db"
)

var teamRouter = Router.PathPrefix("/team").Subrouter()

func init() {
	Router.HandleFunc("/teams", getTeams).Methods("GET")
	Router.HandleFunc("/team_groups", getTeamGroups).Methods("GET")

	teamRouter.HandleFunc("/add", adminOnly(addTeam)).Methods("POST")
}

func getTeams(w http.ResponseWriter, r *http.Request) {
	isEnable := getParamBool(r, "is_enable", false)
	teams := db.GetTeams(r.Context(), isEnable)
	dataResponse(w, teams)
}

func getTeamGroups(w http.ResponseWriter, r *http.Request) {
	isGrade := getParamBool(r, "is_grade", false)
	isEnable := getParamBool(r, "is_enable", false)
	groups := db.GetTeamGroupsWithTeams(r.Context(), isGrade, isEnable)
	dataResponse(w, groups)
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
		team.Users = append(team.Users, db.UserSimple{Username: s})
	}
	db.AddTeam(r.Context(), team)
	msgResponse(w, http.StatusOK, "添加队伍成功")
}
