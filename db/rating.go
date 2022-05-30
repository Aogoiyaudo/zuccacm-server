package db

import (
	"context"
	"time"
)

type Rating struct {
	Username    string    `db:"username"`
	OjId        int       `db:"oj_id"`
	Rating      int       `db:"rating"`
	ContestRank int       `db:"contest_rank"`
	ContestTime time.Time `db:"contest_time"`
	ContestName string    `db:"contest_name"`
	ContestURL  string    `db:"contest_url"`
}

func UpdRating(ctx context.Context, username string, ojId int, ratings []Rating) {
	tx := instance.MustBeginTx(ctx, nil)
	defer tx.Rollback()
	query := `
DELETE FROM rating
WHERE username=? AND oj_id=?`
	mustExecTx(tx, ctx, query, username, ojId)
	query = `
INSERT INTO rating(username, oj_id, rating, contest_rank, contest_time, contest_name, contest_url)
VALUES(:username, :oj_id, :rating, :contest_rank, :contest_time, :contest_name, :contest_url)`
	mustNamedExecTx(tx, ctx, query, ratings)
	mustCommit(tx)
}

func GetRating(ctx context.Context, username string, ojId int) int {
	query := `
SELECT rating FROM rating
WHERE username=? AND oj_id=? AND contest_time=
(
  SELECT MAX(contest_time) FROM rating
  WHERE username=? AND oj_id=? AND contest_rank > 0
)`
	var data struct {
		Rating int `db:"rating"`
	}
	var args []interface{}
	for i := 0; i < 2; i++ {
		args = append(args, username)
		args = append(args, ojId)
	}
	mustGet(ctx, &data, query, args...)
	return data.Rating
}

func GetMaxRating(ctx context.Context, username string, ojId int) int {
	query := `
SELECT MAX(rating) AS rating
FROM rating
WHERE username=? AND oj_id=?`
	var data struct {
		Rating int `db:"rating"`
	}
	mustGet(ctx, &data, query, username, ojId)
	return data.Rating
}

type userRating struct {
	Username  string `db:"username"`
	Rating    int    `db:"rating"`
	MaxRating int    `db:"max_rating"`
}

func GetOfficialUserRatings(ctx context.Context, ojId int) []userRating {
	query := `
SELECT username,
IFNULL((
    SELECT MAX(rating) FROM rating
    WHERE username= official_user.username AND oj_id = ?
), 0) AS max_rating,
IFNULL((
    SELECT rating FROM rating
    WHERE username = official_user.username AND oj_id = ? AND contest_time =
    (
        SELECT MAX(contest_time) FROM rating
        WHERE username = official_user.username AND contest_rank > 0
    )
), 0) AS rating
FROM official_user`
	data := make([]userRating, 0)
	mustSelect(ctx, &data, query, ojId, ojId)
	return data
}
