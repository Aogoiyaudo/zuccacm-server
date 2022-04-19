package handler

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"zuccacm-server/db"
	"zuccacm-server/mq"
	"zuccacm-server/utils"
)

var contestRouter = Router.PathPrefix("/contest").Subrouter()
var contestGroupRouter = Router.PathPrefix("/contest_group").Subrouter()

func init() {
	contestRouter.HandleFunc("/add", adminOnly(addContest)).Methods("POST")
	contestRouter.HandleFunc("/upd", adminOnly(updContest)).Methods("POST")
	contestRouter.HandleFunc("/refresh", adminOnly(refreshContest)).Methods("POST")
	contestRouter.HandleFunc("/{id}", getContest).Methods("GET")

	Router.HandleFunc("/contest_groups", getContestGroups).Methods("GET")
	contestGroupRouter.HandleFunc("/{id}", getContests).Methods("GET")
	contestGroupRouter.HandleFunc("/{id}/overview", getContestGroupOverview).Methods("GET")
}

// getContest return contest info and standing
func getContest(w http.ResponseWriter, r *http.Request) {
	type row struct {
		Id             string          `json:"id"`
		Name           string          `json:"name"`
		Solved         int             `json:"solved"`
		ProblemResults []problemResult `json:"problem_results"`
	}
	type standing struct {
		Team  row   `json:"team"`
		Users []row `json:"users"`
	}
	var data struct {
		Contest   db.Contest `json:"contest"`
		Standings []standing `json:"standings"`
	}
	id := getParamIntURL(r, "id")
	ctx := r.Context()
	data.Contest = db.GetContestById(ctx, id)
	mpIndex := make(map[string]int)
	for i, p := range data.Contest.Problems {
		mpIndex[p.Index] = i
	}
	teams := db.GetTeamsInContest(ctx, id)
	mpUser := make(map[string]*row)
	mpTeam := make(map[int]*row)
	for _, t := range teams {
		mpTeam[t.Id] = &row{
			Id:             strconv.Itoa(t.Id),
			Name:           t.Name,
			ProblemResults: make([]problemResult, len(data.Contest.Problems)),
		}
		for _, u := range t.Users {
			mpUser[u.Username] = &row{
				Id:             u.Username,
				Name:           u.Nickname,
				ProblemResults: make([]problemResult, len(data.Contest.Problems)),
			}
		}
	}
	for _, v := range mpUser {
		for i := range v.ProblemResults {
			v.ProblemResults[i].AcceptedTime = -1
		}
	}
	for _, v := range mpTeam {
		for i := range v.ProblemResults {
			v.ProblemResults[i].AcceptedTime = -1
		}
	}
	sub := db.GetSubmissionsInContest(ctx, id)
	for _, s := range sub {
		i := mpIndex[s.Index]
		mpUser[s.Username].ProblemResults[i].Submissions = append(mpUser[s.Username].ProblemResults[i].Submissions, submissionInfo{
			IsAccepted: s.IsAccepted,
			CreateTime: db.Datetime(s.CreateTime),
		})
		if mpUser[s.Username].ProblemResults[i].AcceptedTime != -1 {
			continue
		}
		if s.IsAccepted {
			t := (s.CreateTime.Unix() - time.Time(data.Contest.StartTime).Unix()) / 60
			mpUser[s.Username].ProblemResults[i].AcceptedTime = int(t)
			mpUser[s.Username].Solved++
		} else {
			mpUser[s.Username].ProblemResults[i].Dirt++
		}
	}
	data.Standings = make([]standing, 0)
	for _, t := range teams {
		var x standing
		x.Team = *mpTeam[t.Id]
		for _, u := range t.Users {
			y := *mpUser[u.Username]
			x.Users = append(x.Users, y)
			for i, pr := range y.ProblemResults {
				x.Team.ProblemResults[i] = maxResult(x.Team.ProblemResults[i], pr)
			}
		}
		for _, pr := range x.Team.ProblemResults {
			if pr.AcceptedTime != -1 {
				x.Team.Solved++
			}
		}
		// self_team should have no user, normal_team should have no submission
		if t.IsSelf {
			x.Team.Id = t.Users[0].Username
			x.Team.Name = t.Users[0].Nickname
			x.Users = nil
		} else {
			for i := range x.Team.ProblemResults {
				x.Team.ProblemResults[i].Submissions = nil
			}
		}
		data.Standings = append(data.Standings, x)
	}
	sort.SliceStable(data.Standings, func(i, j int) bool {
		return data.Standings[i].Team.Solved > data.Standings[j].Team.Solved
	})
	dataResponse(w, data)
}

func addContest(w http.ResponseWriter, r *http.Request) {
	var contest db.Contest
	decodeParamVar(r, &contest)
	db.AddContest(r.Context(), contest)
	msgResponse(w, http.StatusOK, "添加比赛成功")
}

func updContest(w http.ResponseWriter, r *http.Request) {
	var contest db.Contest
	decodeParamVar(r, &contest)
	db.UpdContest(r.Context(), contest)
	msgResponse(w, http.StatusOK, "更新比赛成功")
}

func refreshContest(w http.ResponseWriter, r *http.Request) {
	args := struct {
		Id    int    `json:"id"`
		Group string `json:"group"`
	}{}
	decodeParamVar(r, &args)
	ctx := r.Context()
	c := db.GetContestById(ctx, args.Id)
	if c.OjId == 0 {
		panic(utils.ErrorMessage{Msg: "can't refresh contest without oj_id"})
	}
	mq.ExecTask(mq.Topic(c.OjId), mq.ContestTask(args.Id, c.Cid, args.Group))
	msgResponse(w, http.StatusOK, "任务已创建：获取比赛")
}

func getContestGroups(w http.ResponseWriter, r *http.Request) {
	isEnable := getParamBool(r, "is_enable", true)
	groups := db.GetContestGroups(r.Context(), isEnable)
	dataResponse(w, groups)
}

func getContests(w http.ResponseWriter, r *http.Request) {
	id := getParamIntURL(r, "id")
	begin := getParamTime(r, "begin_time", defaultBeginTime)
	end := getParamTime(r, "end_time", defaultEndTime).Add(time.Hour * 24)
	dataResponse(w, db.GetContests(r.Context(), id, begin, end))
}

func getContestGroupOverview(w http.ResponseWriter, r *http.Request) {
	id := getParamIntURL(r, "id")
	begin := getParamTime(r, "begin_time", defaultBeginTime)
	end := getParamTime(r, "end_time", defaultEndTime).Add(time.Hour * 24)
	ctx := r.Context()
	userGroup := db.GetOfficialUsers(ctx, false)
	type Data struct {
		GroupId   int           `json:"group_id"`
		GroupName string        `json:"group_name"`
		Users     []db.Overview `json:"users"`
	}
	grp := make(map[int]*Data)
	grpId := make(map[string]int)
	for _, x := range userGroup {
		grp[x.GroupId] = &Data{
			GroupId:   x.GroupId,
			GroupName: x.GroupName,
			Users:     make([]db.Overview, 0),
		}
		for _, u := range x.Users {
			grpId[u.Username] = x.GroupId
		}
	}
	overview := db.GetContestGroupOverview(ctx, id, begin, end)
	for _, x := range overview {
		i := grpId[x.Username]
		grp[i].Users = append(grp[i].Users, x)
	}
	data := make([]Data, 0)
	for _, v := range grp {
		if len(v.Users) > 0 {
			data = append(data, *v)
		}
	}
	sort.SliceStable(data, func(i, j int) bool {
		return data[i].GroupName > data[j].GroupName
	})
	dataResponse(w, data)
}
