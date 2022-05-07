package handler

import (
	"bytes"
	"net/http"

	"github.com/Jeffail/gabs/v2"
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"

	"zuccacm-server/config"
	"zuccacm-server/db"
)

var (
	ssoURL       = config.Instance.SSO_URL
	sessionStore = sessions.NewCookieStore([]byte(config.Instance.SessionKey))
)

func init() {
	sessionStore.MaxAge(0)
	sessionStore.Options.Secure = true
	sessionStore.Options.SameSite = http.SameSiteNoneMode

	Router.HandleFunc("/session", handlerCurrentUser).Methods("GET")
	Router.HandleFunc("/login", ssoLogin).Methods("POST")
	Router.HandleFunc("/session", loginRequired(logout)).Methods("DELETE")
}

func handlerCurrentUser(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r)
	dataResponse(w, user)
}

func ssoLogin(w http.ResponseWriter, r *http.Request) {
	args := decodeParam(r.Body)
	resp, err := http.Post(ssoURL, "application/json", bytes.NewReader([]byte((*gabs.Container)(args).String())))
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != http.StatusOK {
		panic(ErrLoginFailed.New())
	}
	username := args.getString("username")
	ctx := r.Context()
	user := db.GetUserByUsername(ctx, username)
	if user == nil {
		// create user
		log.WithField("username", username).Warn("valid user but not found, creating user...")
		db.AddUser(ctx, db.User{Username: username, Nickname: username, IsAdmin: false, IsEnable: true})
	}
	session := mustGetSession(r)
	session.Values["username"] = username
	mustSaveSession(session, r, w)
	msgResponse(w, http.StatusOK, "登录成功")
}

func logout(w http.ResponseWriter, r *http.Request) {
	session := mustGetSession(r)
	for key := range session.Values {
		delete(session.Values, key)
	}
	mustSaveSession(session, r, w)
	msgResponse(w, http.StatusOK, "登出成功")
}

func mustGetSession(r *http.Request) *sessions.Session {
	session, err := sessionStore.Get(r, "mainsite-session")
	if err != nil {
		panic(err)
	}
	return session
}

func mustSaveSession(session *sessions.Session, r *http.Request, w http.ResponseWriter) {
	err := session.Save(r, w)
	if err != nil {
		panic(err)
	}
}

func getCurrentUser(r *http.Request) *db.User {
	session := mustGetSession(r)
	username := session.Values["username"]
	if username == nil {
		panic(ErrNotLogged.New())
	}
	user := db.GetUserByUsername(r.Context(), username.(string))
	return user
}
