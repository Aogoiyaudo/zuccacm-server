package handler

import (
	"net/http"
	"time"

	"zuccacm-server/db"
)

var ratingRouter = Router.PathPrefix("/rating").Subrouter()

func init() {
	ratingRouter.HandleFunc("/upd", updRating).Methods("POST")
}

func updRating(w http.ResponseWriter, r *http.Request) {
	var data []struct {
		OJ       string `json:"oj"`
		Username string `json:"username"`
		Ratings  []struct {
			Rating      int         `json:"rating"`
			ContestRank int         `json:"contest_rank"`
			ContestTime db.Datetime `json:"contest_time"`
			ContestName string      `json:"contest_name"`
			ContestURL  string      `json:"contest_url"`
		} `json:"ratings"`
	}
	decodeParamVar(r, &data)
	ctx := r.Context()

	oj := db.OJMapStoI(db.GetAllEnableOJ(ctx))
	mp := db.GetAllAccountsMap(ctx)
	for _, x := range data {
		ratings := make([]db.Rating, 0)
		ojId := oj[x.OJ]
		username := mp[db.Account{OjId: ojId, Account: x.Username}]
		for _, y := range x.Ratings {
			ratings = append(ratings, db.Rating{
				OjId:        ojId,
				Username:    username,
				Rating:      y.Rating,
				ContestRank: y.ContestRank,
				ContestTime: time.Time(y.ContestTime),
				ContestName: y.ContestName,
				ContestURL:  y.ContestURL,
			})
		}
		db.UpdRating(ctx, username, ojId, ratings)
	}
	msgResponse(w, http.StatusOK, "upd user rating success")
}
