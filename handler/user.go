package handler

import (
	"net/http"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"

	"zuccacm-server/db"
	"zuccacm-server/enum/tshirt"
	"zuccacm-server/mq"
	"zuccacm-server/utils"
)

var userRouter = Router.PathPrefix("/user").Subrouter()

func init() {
	userRouter.HandleFunc("/add", adminOnly(addUser)).Methods("POST")
	userRouter.HandleFunc("/upd", userSelfOrAdminOnly(updUser)).Methods("POST")
	userRouter.HandleFunc("/upd_oj", userSelfOrAdminOnly(updUserAccount)).Methods("POST")
	userRouter.HandleFunc("/upd_admin", adminOnly(updUserAdmin)).Methods("POST")
	userRouter.HandleFunc("/upd_enable", adminOnly(updUserEnable)).Methods("POST")
	userRouter.HandleFunc("/upd_rating", updRating).Methods("POST")
	userRouter.HandleFunc("/upd_groups", adminOnly(updGroups)).Methods("POST")
	userRouter.HandleFunc("/refresh_rating", refreshUserRating).Methods("POST")

	userRouter.HandleFunc("/{username}", getUser).Methods("GET")
	userRouter.HandleFunc("/{username}/profile", userSelfOrAdminOnly(getUserProfile)).Methods("GET")
	userRouter.HandleFunc("/{username}/accounts", getUserAccounts).Methods("GET")
	userRouter.HandleFunc("/{username}/submissions", getUserSubmissions).Methods("GET")
	userRouter.HandleFunc("/{username}/contests", getUserContests).Methods("GET")
	userRouter.HandleFunc("/{username}/groups", getGroupsByUser).Methods("GET")

	Router.HandleFunc("/users", getUsers).Methods("GET")
	Router.HandleFunc("/members", getMembers).Methods("GET")
}

func addUser(w http.ResponseWriter, r *http.Request) {
	var user db.User
	decodeParamVar(r, &user)
	if _, err := tshirt.Parse(user.TShirt); err != nil {
		log.WithField("err", err).Warn("T-shirt size parse failed, use empty string instead")
		user.TShirt = ""
	}
	db.AddUser(r.Context(), user)
	msgResponse(w, http.StatusOK, "添加用户成功")
}

// updUser update db.User basic info (nickname, id_card, phone, qq, t_shirt)
func updUser(w http.ResponseWriter, r *http.Request) {
	var user db.User
	decodeParamVar(r, &user)
	if _, err := tshirt.Parse(user.TShirt); err != nil {
		log.WithField("err", err).Warn("T-shirt size parse failed, use empty string instead")
		user.TShirt = ""
	}
	db.UpdUser(r.Context(), user)
	msgResponse(w, http.StatusOK, "修改用户信息成功")
}

func updUserAccount(w http.ResponseWriter, r *http.Request) {
	var account db.Account
	decodeParamVar(r, &account)
	db.UpdAccount(r.Context(), account)
	msgResponse(w, http.StatusOK, "修改用户账号成功")
}

func updUserAdmin(w http.ResponseWriter, r *http.Request) {
	var user db.User
	decodeParamVar(r, &user)
	db.UpdUserAdmin(r.Context(), user)
	msgResponse(w, http.StatusOK, "修改用户权限成功")
}

func updUserEnable(w http.ResponseWriter, r *http.Request) {
	var user db.User
	decodeParamVar(r, &user)
	db.UpdUserEnable(r.Context(), user)
	msgResponse(w, http.StatusOK, "修改用户状态成功")
}

func updRating(w http.ResponseWriter, r *http.Request) {
	args := make([]struct {
		Account string `json:"username"`
		OJ      string `json:"oj"`
		Rating  int    `json:"rating"`
	}, 0)
	decodeParamVar(r, &args)
	ctx := r.Context()

	oj := db.OJMapStoI(db.GetAllEnableOJ(ctx))
	mp := db.GetAllAccountsMap(ctx)
	users := make([]db.User, 0)
	for _, arg := range args {
		ac := db.Account{OjId: oj[arg.OJ], Account: arg.Account}
		if _, ok := mp[ac]; !ok {
			log.WithFields(log.Fields{
				"account": arg.Account,
				"oj":      arg.OJ,
				"rating":  arg.Rating,
			}).Error("account not found")
			continue
		}
		users = append(users, db.User{
			Username: mp[ac],
			CfRating: arg.Rating,
		})
	}
	db.UpdUserRating(ctx, users)
	msgResponse(w, http.StatusOK, "upd user rating success")
}

func updGroups(w http.ResponseWriter, r *http.Request) {
	var args struct {
		Username string `json:"username"`
		Groups   []int  `json:"groups"`
	}
	decodeParamVar(r, &args)

	db.UpdUserGroups(r.Context(), args.Username, args.Groups)
	msgResponse(w, http.StatusOK, "修改用户所属组成功")
}

func refreshUserRating(w http.ResponseWriter, r *http.Request) {
	args := decodeParam(r.Body)
	ojId := args.getInt("oj_id")
	tmp := args.get("username").([]interface{})
	var username []string
	for _, u := range tmp {
		username = append(username, u.(string))
	}
	mq.ExecTask(mq.Topic(ojId), mq.RatingTask(username))
	msgResponse(w, http.StatusOK, "任务已创建：刷新Rating")
}

// getUser return user's basic info and awards
func getUser(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Username string     `json:"username"`
		Nickname string     `json:"nickname"`
		CfRating int        `json:"cf_rating"`
		IsEnable bool       `json:"is_enable"`
		IsAdmin  bool       `json:"is_admin"`
		Medals   [3]int     `json:"medals"`
		Awards   []db.Award `json:"awards"`
	}
	username := getParamURL(r, "username")
	ctx := r.Context()

	u := db.MustGetUser(ctx, username)
	data.Username = u.Username
	data.Nickname = u.Nickname
	data.CfRating = u.CfRating
	data.IsEnable = u.IsEnable
	data.IsAdmin = u.IsAdmin
	data.Awards = db.GetAwardsByUsername(ctx, username)
	for _, x := range data.Awards {
		if x.Medal > 0 {
			data.Medals[x.Medal-1]++
		}
	}
	dataResponse(w, data)
}

func getUserProfile(w http.ResponseWriter, r *http.Request) {
	username := getParamURL(r, "username")
	user := db.GetUserByUsername(r.Context(), username)
	dataResponse(w, user)
}

func getUserAccounts(w http.ResponseWriter, r *http.Request) {
	type account struct {
		OjId    int    `json:"oj_id" db:"oj_id"`
		OjName  string `json:"oj_name" db:"oj_name"`
		Account string `json:"account" db:"account"`
	}
	ctx := r.Context()
	username := getParamURL(r, "username")

	oj := db.GetAllEnableOJ(ctx)
	data := make([]account, len(oj))
	mp := make(map[int]int)
	for i, x := range oj {
		data[i] = account{
			OjId:   x.OjId,
			OjName: x.OjName,
		}
		mp[x.OjId] = i
	}
	ac := db.GetAccountsByUsername(ctx, username)
	for _, x := range ac {
		data[mp[x.OjId]].Account = x.Account
	}
	dataResponse(w, data)
}

func getUserSubmissions(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Date       string `json:"date"`
		Solved     int    `json:"solved"`
		Submission int    `json:"submission"`
	}
	ctx := r.Context()
	username := getParamURL(r, "username")
	begin := getParamDateRequired(r, "begin_time")
	end := getParamDateRequired(r, "end_time").Add(time.Hour * 24).Add(time.Second * -1)
	n := utils.SubDays(begin, end) + 1
	data := make([]Data, n)
	for i := range data {
		data[i].Date = db.Datetime(begin.AddDate(0, 0, i)).Date()
	}
	submissions := db.GetAcceptedSubmissionByUsername(ctx, username, begin, end)
	for _, s := range submissions {
		i := utils.SubDays(begin, time.Time(s.CreateTime))
		data[i].Solved++
	}
	submissions = db.GetSubmissionByUsername(ctx, username, begin, end)
	for _, s := range submissions {
		i := utils.SubDays(begin, time.Time(s.CreateTime))
		data[i].Submission++
	}
	dataResponse(w, data)
}

func getUserContests(w http.ResponseWriter, r *http.Request) {
	type Row struct {
		ContestId      int             `json:"contest_id"`
		ContestName    string          `json:"contest_name"`
		StartTime      db.Datetime     `json:"start_time"`
		Duration       int             `json:"duration"`
		Solved         int             `json:"solved"`
		Problems       []db.Problem    `json:"problems"`
		ProblemResults []problemResult `json:"problem_results"`
	}
	data := struct {
		MaxProblems int   `json:"max_problems"`
		Contests    []Row `json:"contests"`
	}{0, make([]Row, 0)}
	ctx := r.Context()
	username := getParamURL(r, "username")
	begin := getParamDate(r, "begin_time", defaultBeginTime)
	end := getParamDate(r, "end_time", defaultEndTime).Add(time.Hour * 24).Add(time.Second * -1)
	groupId := getParamInt(r, "group_id", 0)
	contests := db.GetContestsByUser(ctx, username, begin, end, groupId)
	for _, c := range contests {
		data.Contests = append(data.Contests, Row{
			ContestId:      c.Id,
			ContestName:    c.Name,
			StartTime:      c.StartTime,
			Duration:       c.Duration,
			Problems:       c.Problems,
			ProblemResults: make([]problemResult, len(c.Problems)),
		})
		data.MaxProblems = utils.Max(data.MaxProblems, len(c.Problems))
	}
	submissions := db.GetSubmissionByUsername(ctx, username, defaultBeginTime, defaultEndTime)
	type Key struct {
		OjId int
		Pid  string
	}
	mp := make(map[Key][]submissionInfo)
	for _, s := range submissions {
		key := Key{s.OjId, s.Pid}
		mp[key] = append(mp[key], submissionInfo{s.IsAccepted, s.CreateTime})
	}
	for i, c := range data.Contests {
		for j, p := range c.Problems {
			data.Contests[i].ProblemResults[j] = calcProblemResult(mp[Key{p.OjId, p.Pid}], c.StartTime, c.Duration)
			if data.Contests[i].ProblemResults[j].AcceptedTime != -1 {
				data.Contests[i].Solved++
			}
		}
	}
	dataResponse(w, data)
}

func getGroupsByUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := getParamURL(r, "username")
	groups := make([]db.TeamGroup, 0)
	if !r.URL.Query().Has("is_grade") {
		groups = append(groups, db.GetGroupsByUser(ctx, username, true)...)
		groups = append(groups, db.GetGroupsByUser(ctx, username, false)...)
	} else {
		isGrade := getParamBool(r, "is_grade", false)
		groups = append(groups, db.GetGroupsByUser(ctx, username, isGrade)...)
	}
	type group struct {
		GroupId   int    `json:"group_id"`
		GroupName string `json:"group_name"`
		IsGrade   bool   `json:"is_grade"`
	}
	data := make([]group, 0)
	for _, g := range groups {
		data = append(data, group{
			GroupId:   g.GroupId,
			GroupName: g.GroupName,
			IsGrade:   g.IsGrade,
		})
	}
	dataResponse(w, data)
}

// getUsers return all users with basic info
func getUsers(w http.ResponseWriter, r *http.Request) {
	isEnable := getParamBool(r, "is_enable", false)
	isOfficial := getParamBool(r, "is_official", false)
	page := decodePage(r)
	users := db.GetUsers(r.Context(), isEnable, isOfficial, page)
	type user struct {
		Username string `json:"username"`
		Nickname string `json:"nickname"`
		CfRating int    `json:"cf_rating"`
		IsEnable bool   `json:"is_enable"`
		IsAdmin  bool   `json:"is_admin"`
	}
	data := make([]user, 0)
	for _, u := range users {
		data = append(data, user{
			Username: u.Username,
			Nickname: u.Nickname,
			CfRating: u.CfRating,
			IsEnable: u.IsEnable,
			IsAdmin:  u.IsAdmin,
		})
	}
	dataResponse(w, data)
}

// getMembers get all official users if is_enable=false (default is true)
// official users are those who are in team_groups with is_grade=true
func getMembers(w http.ResponseWriter, r *http.Request) {
	type user struct {
		Username string   `json:"username"`
		Nickname string   `json:"nickname"`
		CfRating int      `json:"cf_rating"`
		Awards   []string `json:"awards"`
		Medals   [3]int   `json:"medals"`
	}
	type group struct {
		GroupId   int    `json:"group_id"`
		GroupName string `json:"group_name"`
		Users     []user `json:"users"`
	}
	isEnable := getParamBool(r, "is_enable", false)
	ctx := r.Context()

	mpUser := make(map[string]*user)
	mpGroup := make(map[int]*group)
	userGroup := db.GetOfficialUsers(ctx, isEnable)
	for _, x := range userGroup {
		mpGroup[x.GroupId] = &group{
			GroupId:   x.GroupId,
			GroupName: x.GroupName,
			Users:     make([]user, 0),
		}
		for _, u := range x.Users {
			mpUser[u.Username] = &user{
				Username: u.Username,
				Nickname: u.Nickname,
				CfRating: u.CfRating,
				Awards:   make([]string, 0),
			}
		}
	}
	userAward := db.GetAwardsAll(ctx, isEnable)
	for _, x := range userAward {
		if _, ok := mpUser[x.Username]; !ok {
			log.WithFields(log.Fields{
				"username": x.Username,
				"award":    x.Award,
			}).Warn("unofficial user with award")
			continue
		}
		if x.Medal > 0 {
			mpUser[x.Username].Medals[x.Medal-1]++
		}
		if len(x.Award) > 0 {
			mpUser[x.Username].Awards = append(mpUser[x.Username].Awards, x.Award)
		}
	}
	for _, x := range userGroup {
		for _, u := range x.Users {
			mpGroup[x.GroupId].Users = append(mpGroup[x.GroupId].Users, *mpUser[u.Username])
		}
	}
	var data []group
	for _, x := range mpGroup {
		data = append(data, *x)
	}
	sort.SliceStable(data, func(i, j int) bool {
		return data[i].GroupName > data[j].GroupName
	})
	for k := range data {
		sort.SliceStable(data[k].Users, func(i, j int) bool {
			return data[k].Users[i].Username < data[k].Users[j].Username
		})
	}
	dataResponse(w, data)
}
