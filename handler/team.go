package handler

import (
	"net/http"
	"zuccacm-server/db"
)

var teamRouter = Router.PathPrefix("/team").Subrouter()

func init() {
	Router.HandleFunc("/teams", getTeams).Methods("GET")
	teamRouter.HandleFunc("/{team_id}", getTeam).Methods("GET")
	Router.HandleFunc("/team_groups", getTeamGroups).Methods("GET")
	Router.HandleFunc("/team_group/add", addTeamGroup).Methods("POST")

	teamRouter.HandleFunc("/add", adminOnly(addTeam)).Methods("POST")
	teamRouter.HandleFunc("/upd_enable", adminOnly(updTeamEnable)).Methods("POST")
}
func getTeam(w http.ResponseWriter, r *http.Request) {
	teamId := getParamURL(r, "team_id")
	team, err := db.GetTeam(r.Context(), teamId)
	if err != nil {
		msgResponse(w, http.StatusBadRequest, "队伍不存在")
		return
	}
	dataResponse(w, team)
}
func getTeams(w http.ResponseWriter, r *http.Request) {
	isEnable := getParamBool(r, "is_enable", false)
	teams := db.GetTeams(r.Context(), isEnable)
	dataResponse(w, teams)
}

func getTeamGroups(w http.ResponseWriter, r *http.Request) {
	isGrade := getParamBool(r, "is_grade", false)
	isEnable := getParamBool(r, "is_enable", false)
	showEmpty := getParamBool(r, "show_empty", false)
	groups := db.GetTeamGroupsWithTeams(r.Context(), isGrade, isEnable, showEmpty)
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
func addTeamGroup(w http.ResponseWriter, r *http.Request) {
	var args struct {
		Name  string    `json:"name"`
		Teams []db.Team `json:"teams"`
	}
	decodeParamVar(r, &args)
	teamGroup := db.TeamGroup{
		GroupName: args.Name,
		IsGrade:   false,
	}
	for _, s := range args.Teams {
		teamGroup.Teams = append(teamGroup.Teams, s)
	}
	db.AddTeamGroup(r.Context(), teamGroup)
	msgResponse(w, http.StatusOK, "添加队伍分组成功")
}
func updTeamEnable(w http.ResponseWriter, r *http.Request) {
	var team db.Team
	decodeParamVar(r, &team)
	db.UpdTeamEnable(r.Context(), team)
	msgResponse(w, http.StatusOK, "修改队伍状态成功")
}
