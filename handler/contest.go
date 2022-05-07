package handler

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"zuccacm-server/db"
	"zuccacm-server/mq"
)

var contestRouter = Router.PathPrefix("/contest").Subrouter()
var contestGroupRouter = Router.PathPrefix("/contest_group").Subrouter()

func init() {
	contestRouter.HandleFunc("/add", adminOnly(addContest)).Methods("POST")
	contestRouter.HandleFunc("/upd", adminOnly(updContest)).Methods("POST")
	contestRouter.HandleFunc("/del", adminOnly(delContest)).Methods("POST")
	contestRouter.HandleFunc("/refresh", adminOnly(refreshContest)).Methods("POST")
	contestRouter.HandleFunc("/pull", pullContest).Methods("POST")

	Router.HandleFunc("/contests", getAllContests).Methods("GET")
	contestRouter.HandleFunc("/{id}", getContest).Methods("GET")
	contestRouter.HandleFunc("/{id}/standings", getContestStandings).Methods("GET")

	Router.HandleFunc("/contest_groups", getContestGroups).Methods("GET")
	contestGroupRouter.HandleFunc("/{id}", getContests).Methods("GET")
	contestGroupRouter.HandleFunc("/{id}/overview", getContestGroupOverview).Methods("GET")
}

func getContests(w http.ResponseWriter, r *http.Request) {
	id := getParamIntURL(r, "id")
	begin := getParamDate(r, "begin_time", defaultBeginTime)
	end := getParamDate(r, "end_time", defaultEndTime).Add(time.Hour * 24).Add(time.Second * -1)
	page := decodePage(r)
	dataResponse(w, db.GetContestsByGroup(r.Context(), id, begin, end, page))
}

// getContest return contest info
func getContest(w http.ResponseWriter, r *http.Request) {
	id := getParamIntURL(r, "id")
	ctx := r.Context()
	contest := db.GetContestById(ctx, id)
	groups := db.GetGroupsByContest(ctx, id)
	teams := db.GetTeamsInContest(ctx, id)
	data := struct {
		Id           int               `json:"id"`
		OjId         int               `json:"oj_id"`
		Cid          string            `json:"cid"`
		Name         string            `json:"name"`
		StartTime    db.Datetime       `json:"start_time"`
		Duration     int               `json:"duration"`
		MaxSolved    int               `json:"max_solved"`
		Participants int               `json:"participants"`
		Problems     []db.Problem      `json:"problems"`
		Groups       []db.ContestGroup `json:"groups"`
		Teams        []db.Team         `json:"teams"`
	}{
		Id:           contest.Id,
		OjId:         contest.OjId,
		Cid:          contest.Cid,
		Name:         contest.Name,
		StartTime:    contest.StartTime,
		Duration:     contest.Duration,
		MaxSolved:    contest.MaxSolved,
		Participants: contest.Participants,
		Problems:     contest.Problems,
		Groups:       groups,
		Teams:        teams,
	}
	dataResponse(w, data)
}

// getContestStandings return contest info and standing
func getContestStandings(w http.ResponseWriter, r *http.Request) {
	type Row struct {
		Id             string          `json:"id"`
		Name           string          `json:"name"`
		Solved         int             `json:"solved"`
		ProblemResults []problemResult `json:"problem_results"`
	}
	type standing struct {
		Team  Row   `json:"team"`
		Users []Row `json:"users"`
	}
	var data struct {
		Contest   db.Contest `json:"contest"`
		Standings []standing `json:"standings"`
	}
	id := getParamIntURL(r, "id")
	ctx := r.Context()

	sub := db.GetSubmissionsInContest(ctx, id)
	type Key struct {
		Username string
		OjId     int
		Pid      string
	}
	mpSub := make(map[Key][]submissionInfo)
	for _, s := range sub {
		key := Key{s.Username, s.OjId, s.Pid}
		mpSub[key] = append(mpSub[key], submissionInfo{s.IsAccepted, s.CreateTime})
	}
	data.Standings = make([]standing, 0)

	contest := db.GetContestById(ctx, id)
	teams := db.GetTeamsInContest(ctx, id)
	for _, t := range teams {
		x := standing{
			Team: Row{
				Id:             strconv.Itoa(t.Id),
				Name:           t.Name,
				ProblemResults: make([]problemResult, len(contest.Problems)),
			},
			Users: make([]Row, 0),
		}
		// init team Row
		for i := range contest.Problems {
			x.Team.ProblemResults[i] = calcProblemResult(nil, contest.StartTime, contest.Duration)
		}
		for _, u := range t.Users {
			uRow := Row{
				Id:             u.Username,
				Name:           u.Nickname,
				ProblemResults: make([]problemResult, len(contest.Problems)),
			}
			for i, p := range contest.Problems {
				key := Key{u.Username, p.OjId, p.Pid}
				uRow.ProblemResults[i] = calcProblemResult(mpSub[key], contest.StartTime, contest.Duration)
			}
			for _, pr := range uRow.ProblemResults {
				if pr.AcceptedTime != -1 {
					uRow.Solved++
				}
			}
			// upd team Row
			for i, pr := range uRow.ProblemResults {
				x.Team.ProblemResults[i] = maxResult(x.Team.ProblemResults[i], pr)
			}
			x.Users = append(x.Users, uRow)
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
		if data.Standings[i].Team.Solved == data.Standings[j].Team.Solved {
			return data.Standings[i].Team.Name < data.Standings[j].Team.Name
		}
		return data.Standings[i].Team.Solved > data.Standings[j].Team.Solved
	})
	oj := db.OJMapItoS(db.GetAllOJ(ctx))
	for i, p := range contest.Problems {
		contest.Problems[i].ProblemURL = getProblemURL(oj[p.OjId], p.Pid)
	}
	data.Contest = contest
	dataResponse(w, data)
}

func addContest(w http.ResponseWriter, r *http.Request) {
	var contest db.Contest
	contest.StartTime = db.Datetime(defaultBeginTime)
	decodeParamVar(r, &contest)
	contest = db.AddContest(r.Context(), contest)
	if contest.OjId > 0 {
		mq.ContestTask(contest.OjId, contest.Id, contest.Cid)
	}
	msgResponse(w, http.StatusOK, "添加比赛成功")
}

func updContest(w http.ResponseWriter, r *http.Request) {
	var contest db.Contest
	contest.StartTime = db.Datetime(defaultBeginTime)
	decodeParamVar(r, &contest)
	if contest.Id == 0 {
		panic(ErrBadRequest.WithMessage("contest.id can't be empty or zero"))
	}
	db.UpdContest(r.Context(), contest)
	if contest.OjId > 0 {
		mq.ContestTask(contest.OjId, contest.Id, contest.Cid)
	}
	msgResponse(w, http.StatusOK, "修改比赛成功")
}

func delContest(w http.ResponseWriter, r *http.Request) {
	args := decodeParam(r.Body)
	contestId := args.getInt("contest_id")
	db.DelContest(r.Context(), contestId)
	msgResponse(w, http.StatusOK, "删除比赛成功")
}

func refreshContest(w http.ResponseWriter, r *http.Request) {
	args := struct {
		Id int `json:"id"`
	}{}
	decodeParamVar(r, &args)
	ctx := r.Context()
	c := db.GetContestById(ctx, args.Id)
	if c.OjId == 0 {
		panic(ErrBadRequest.WithMessage("oj_id can't be empty or zero"))
	}
	mq.ContestTask(c.OjId, c.Id, c.Cid)
	msgResponse(w, http.StatusOK, "任务已创建：刷新比赛")
}

func pullContest(w http.ResponseWriter, r *http.Request) {
	type problem struct {
		OJ    string `json:"oj"`
		Pid   string `json:"pid"`
		Index string `json:"index"`
	}
	var arg struct {
		Id           int         `json:"id"`
		OJ           string      `json:"oj"`
		Cid          string      `json:"cid"`
		Name         string      `json:"name"`
		StartTime    db.Datetime `json:"start_time"`
		Duration     int         `json:"duration"`
		MaxSolved    int         `json:"max_solved"`
		Participants int         `json:"participants"`
		Problems     []problem   `json:"problems"`
	}
	decodeParamVar(r, &arg)
	if arg.Id == 0 {
		panic(ErrBadRequest.WithMessage("contest.id can't be empty or zero"))
	}
	ctx := r.Context()

	oj := db.OJMapStoI(db.GetAllOJ(ctx))
	contest := db.Contest{
		Id:           arg.Id,
		OjId:         oj[arg.OJ],
		Cid:          arg.Cid,
		Name:         arg.Name,
		StartTime:    arg.StartTime,
		Duration:     arg.Duration,
		MaxSolved:    arg.MaxSolved,
		Participants: arg.Participants,
		Problems:     make([]db.Problem, 0),
	}
	for _, p := range arg.Problems {
		contest.Problems = append(contest.Problems, db.Problem{
			ContestId: contest.Id,
			OjId:      oj[p.OJ],
			Pid:       p.Pid,
			Index:     p.Index,
		})
	}
	db.PullContest(ctx, contest)
	msgResponse(w, http.StatusOK, "pull contest success")
}

func getContestGroups(w http.ResponseWriter, r *http.Request) {
	isEnable := getParamBool(r, "is_enable", true)
	groups := db.GetContestGroups(r.Context(), isEnable)
	dataResponse(w, groups)
}

func getAllContests(w http.ResponseWriter, r *http.Request) {
	page := decodePage(r)
	dataResponse(w, db.GetContestsByGroup(r.Context(), 0, defaultBeginTime, defaultEndTime, page))
}

func getContestGroupOverview(w http.ResponseWriter, r *http.Request) {
	id := getParamIntURL(r, "id")
	begin := getParamDate(r, "begin_time", defaultBeginTime)
	end := getParamDate(r, "end_time", defaultEndTime).Add(time.Hour * 24).Add(time.Second * -1)
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
