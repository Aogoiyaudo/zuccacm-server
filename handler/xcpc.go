package handler

import (
	"net/http"
	"strconv"
	"time"
	"zuccacm-server/db"
)

var xcpcRouter = Router.PathPrefix("/xcpc").Subrouter()
var xcpc_team_relRouter = Router.PathPrefix("/xcpc_team_rel").Subrouter()

func init() {
	Router.HandleFunc("/xcpcs", adminOnly(getXcpcs)).Methods("GET")
	xcpcRouter.HandleFunc("/{xcpc_id}", adminOnly(getXcpc)).Methods("GET")
	Router.HandleFunc("/xcpc_team_rels", adminOnly(getXcpcTeamRels)).Methods("GET")
	xcpcRouter.HandleFunc("/add", adminOnly(addXcpc)).Methods("POST")
	xcpc_team_relRouter.HandleFunc("/add", adminOnly(addXcpcTeamRel)).Methods("POST")
}
func getXcpc(w http.ResponseWriter, r *http.Request) {
	xcpcId := getParamURL(r, "xcpc_id")
	xcpc, err := db.GetXcpc(r.Context(), xcpcId)
	if err != nil {
		msgResponse(w, http.StatusBadRequest, "比赛不存在")
		return
	}
	dataResponse(w, xcpc)
}
func getXcpcs(w http.ResponseWriter, r *http.Request) {
	type back struct {
		Id   int    `json:"id"   db:"id"`
		Name string `json:"name" db:"name"`
		Date string `json:"date" db:"date"`
	}
	xcpcs := db.GetXcpcs(r.Context())
	data := make([]back, 0)
	for _, i := range xcpcs {
		data = append(data, back{
			Id:   i.Id,
			Name: i.Name,
			Date: time.Time(i.Date).Format("2006-01-02"),
		})
	}
	dataResponse(w, data)
}
func getXcpcTeamRels(w http.ResponseWriter, r *http.Request) {
	type back struct {
		TeamId   int    `json:"team_id"`
		XcpcId   int    `json:"xcpc_id"`
		TeamName string `json:"team_name"`
		XcpcName string `json:"xcpc_name"`
		Medal    int    `json:"medal"`
		Award    string `json:"award"`
	}
	xcpc_team_rels := db.GetXcpcTeamRels(r.Context())
	data := make([]back, 0)
	for _, i := range xcpc_team_rels {
		team, _ := db.GetTeam(r.Context(), strconv.Itoa(i.TeamId))
		xcpc, _ := db.GetXcpc(r.Context(), strconv.Itoa(i.XcpcId))
		data = append(data, back{
			TeamId:   i.TeamId,
			XcpcId:   i.XcpcId,
			TeamName: team.Name,
			XcpcName: xcpc.Name,
			Medal:    i.Medal,
			Award:    i.Award,
		})
	}
	dataResponse(w, data)
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
func addXcpcTeamRel(w http.ResponseWriter, r *http.Request) {
	var args struct {
		TeamId string `json:"team_id"`
		XcpcId string `json:"xcpc_id"`
	}
	decodeParamVar(r, &args)
	xid, err1 := strconv.Atoi(args.XcpcId)
	tid, err2 := strconv.Atoi(args.TeamId)
	if err1 != nil {
		msgResponse(w, http.StatusBadRequest, "输入格式有误")
		return
	}
	if err2 != nil {
		msgResponse(w, http.StatusBadRequest, "输入格式有误")
		return
	}
	xcpc_team_rel := db.XcpcTeamRel{
		XcpcId: xid,
		TeamId: tid,
		Medal:  0,
		Award:  "",
	}
	db.AddXcpcTeamRel(r.Context(), xcpc_team_rel)
	msgResponse(w, http.StatusOK, "增加参赛队伍成功")
}
