package handler

import (
	"net/http"
	"sort"

	"zuccacm-server/db"
	"zuccacm-server/utils"
)

var userRouter = Router.PathPrefix("/user").Subrouter()

func init() {
	userRouter.HandleFunc("/add", adminOnly(addUser)).Methods("POST")
	userRouter.HandleFunc("/upd", updUser).Methods("POST")
	userRouter.HandleFunc("/upd_admin", adminOnly(updUserAdmin)).Methods("POST")
	userRouter.HandleFunc("/upd_enable", adminOnly(updUserEnable)).Methods("POST")
	userRouter.HandleFunc("/upd_oj", updUserOJ).Methods("POST")

	Router.HandleFunc("/users", getUsers).Methods("GET")
}

func addUser(w http.ResponseWriter, r *http.Request) {
	var user db.User
	decodeParamVar(r, &user)
	db.AddUser(r.Context(), user)
	msgResponse(w, http.StatusOK, "添加用户成功")
}

// upd nickname, id_card, phone, qq, t_shirt
func updUser(w http.ResponseWriter, r *http.Request) {
	var user db.User
	decodeParamVar(r, &user)
	now := getCurrentUser(r)
	if !now.IsAdmin && now.Username != user.Username {
		panic(utils.ErrForbidden)
	}
	db.UpdUser(r.Context(), user)
	msgResponse(w, http.StatusOK, "修改用户信息成功")
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

func updUserOJ(w http.ResponseWriter, r *http.Request) {
	var account db.Account
	decodeParamVar(r, &account)
	db.UpdAccount(r.Context(), account)
	msgResponse(w, http.StatusOK, "修改用户账号成功")
}

// getUsers get all official users if is_enable=false (default is true)
func getUsers(w http.ResponseWriter, r *http.Request) {
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
	userAward := db.GetUserAward(ctx, isEnable)
	for _, x := range userAward {
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
