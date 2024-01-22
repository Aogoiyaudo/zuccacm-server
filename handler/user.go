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
	userRouter.HandleFunc("/upd_grade_group", adminOnly(updGradeGroup)).Methods("POST")
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

func updGradeGroup(w http.ResponseWriter, r *http.Request) {
	var args struct {
		Username string `json:"username"`
		Group    int    `json:"group"`
	}
	decodeParamVar(r, &args)

	db.UpdUserGradeGroup(r.Context(), args.Username, args.Group)
	msgResponse(w, http.StatusOK, "修改用户所属组成功")
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
	username := getParamURL(r, "username")
	ctx := r.Context()

	oj := db.OJMapStoI(db.GetAllOJ(ctx))
	cf := oj["codeforces"]
	u := db.MustGetUser(ctx, username)

	data := struct {
		Username    string     `json:"username"`
		Nickname    string     `json:"nickname"`
		CfRating    int        `json:"cf_rating"`
		CfMaxRating int        `json:"cf_max_rating"`
		IsEnable    bool       `json:"is_enable"`
		IsAdmin     bool       `json:"is_admin"`
		Medals      [3]int     `json:"medals"`
		Awards      []db.Award `json:"awards"`
	}{
		Username:    u.Username,
		Nickname:    u.Nickname,
		CfRating:    db.GetRating(ctx, username, cf),
		CfMaxRating: db.GetMaxRating(ctx, username, cf),
		IsEnable:    u.IsEnable,
		IsAdmin:     u.IsAdmin,
		Awards:      db.GetAwardsByUsername(ctx, username),
	}

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
	type RT struct {
		Solve  db.Submission `json:"submission"`
		OjName string        `json:"oj_name"`
	}
	type Data struct {
		Date       string `json:"date"`
		Solved     int    `json:"solved"`
		Submission int    `json:"submission"`
		Solves     []RT   `json:"solves"`
	}
	ctx := r.Context()
	username := getParamURL(r, "username")
	begin, end := getParamDateInterval(r)
	submissions := db.GetSubmissionByUsername(ctx, username, begin, end)
	if begin == defaultBeginTime {
		end = time.Now()
		if len(submissions) == 0 {
			begin = end
		} else {
			begin = time.Time(submissions[0].CreateTime)
		}
	}

	n := utils.SubDays(begin, end) + 1
	data := make([]Data, n)

	for i := range data {
		data[i].Date = db.Datetime(begin.AddDate(0, 0, i)).Date()
		data[i].Solves = make([]RT, 0)
	}
	for _, s := range submissions {
		i := utils.SubDays(begin, time.Time(s.CreateTime))
		data[i].Submission++
	}
	submissions = db.GetAcceptedSubmissionByUsername(ctx, username, begin, end)
	for _, s := range submissions {
		i := utils.SubDays(begin, time.Time(s.CreateTime))
		data[i].Solved++
		oj := db.GetOJById(s.OjId)
		data[i].Solves = append(data[i].Solves, RT{
			Solve:  s,
			OjName: oj.OjName,
		})
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
	begin, end := getParamDateInterval(r)
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
	ctx := r.Context()

	users := db.GetUsers(ctx, isEnable, isOfficial, page)
	grade := db.GetUserGroup(ctx)
	type user struct {
		Username   string `json:"username"`
		Nickname   string `json:"nickname"`
		IsEnable   bool   `json:"is_enable"`
		IsAdmin    bool   `json:"is_admin"`
		GradeGroup string `json:"grade_group"`
	}
	data := make([]user, 0)
	for _, u := range users {
		data = append(data, user{
			Username:   u.Username,
			Nickname:   u.Nickname,
			IsEnable:   u.IsEnable,
			IsAdmin:    u.IsAdmin,
			GradeGroup: grade[u.Username].GroupName,
		})
	}
	dataResponse(w, data)
}

// getMembers get all official users if is_enable=false (default is true)
// official users are those who are in team_groups with is_grade=true
func getMembers(w http.ResponseWriter, r *http.Request) {
	type user struct {
		Username    string   `json:"username"`
		Nickname    string   `json:"nickname"`
		CfRating    int      `json:"cf_rating"`
		CfMaxRating int      `json:"cf_max_rating"`
		Awards      []string `json:"awards"`
		Medals      [3]int   `json:"medals"`
	}
	type group struct {
		GroupId   int    `json:"group_id"`
		GroupName string `json:"group_name"`
		Users     []user `json:"users"`
	}
	isEnable := getParamBool(r, "is_enable", false)
	ctx := r.Context()

	oj := db.OJMapStoI(db.GetAllOJ(ctx))
	ratings := db.GetOfficialUserRatings(ctx, oj["codeforces"])
	type rating struct {
		rating    int
		maxRating int
	}
	cf := make(map[string]rating)
	for _, x := range ratings {
		cf[x.Username] = rating{
			rating:    x.Rating,
			maxRating: x.MaxRating,
		}
	}

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
				Username:    u.Username,
				Nickname:    u.Nickname,
				CfRating:    cf[u.Username].rating,
				CfMaxRating: cf[u.Username].maxRating,
				Awards:      make([]string, 0),
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
