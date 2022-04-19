package handler

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"zuccacm-server/config"
	"zuccacm-server/utils"
)

var Router = mux.NewRouter()

func init() {
	Router.Use(corsMiddleware)
	Router.Use(baseMiddleware)
	Router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		addCORSHeader(w, r)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
		} else {
			msgResponse(w, http.StatusNotFound, "404 not found")
		}
	})
	Router.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		addCORSHeader(w, r)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
		} else {
			msgResponse(w, http.StatusMethodNotAllowed, "405 method not allowed")
		}
	})
}

func addCORSHeader(w http.ResponseWriter, r *http.Request) {
	if len(r.Header["Origin"]) > 0 {
		w.Header().Set("Access-Control-Allow-Origin", r.Header["Origin"][0]) // 允许访问所有域，可以换成具体url，注意仅具体url才能带cookie信息
	}
	w.Header().Add("Access-Control-Allow-Headers", "Origin, Content-Type, AccessToken, X-CSRF-Token, Authorization, Token") //header的类型
	w.Header().Add("Access-Control-Allow-Credentials", "true")                                                              //设置为true，允许ajax异步请求带cookie信息
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")                                       //允许请求方法
	w.Header().Set("content-type", "application/json;charset=UTF-8")                                                        //返回数据格式是json
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.Header["Origin"]) > 0 {
			w.Header().Set("Access-Control-Allow-Origin", r.Header["Origin"][0]) // 允许访问所有域，可以换成具体url，注意仅具体url才能带cookie信息
		}
		addCORSHeader(w, r)
		next.ServeHTTP(w, r)
	})
}

func stackInfo() string {
	info := "\nERROR STACK:\n"
	pc := make([]uintptr, 20)
	n := runtime.Callers(0, pc)
	frames := runtime.CallersFrames(pc[:n])
	for n > 0 {
		frame, _ := frames.Next()
		if utils.IsLocalFile(frame.File, config.RootDir) {
			file := fmt.Sprintf("%s:%d", utils.SimplePath(frame.File, config.RootDir), frame.Line)
			function := fmt.Sprintf("[%s]", utils.SimplePath(frame.Function, config.RootDir))
			info += fmt.Sprintf("%-25s %s\n", file, function)
		}
		n--
	}
	return info
}

// baseMiddleware logging and handle panic
func baseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.WithField("stack", stackInfo()).Error(err)
				resp := &Response{}
				if err == sql.ErrNoRows {
					err = utils.ErrNotFound
				}
				switch err {
				case utils.ErrBadRequest:
					resp.Code = http.StatusBadRequest
				case utils.ErrForbidden:
					resp.Code = http.StatusForbidden
				case utils.ErrNotLogged, utils.ErrLoginFailed:
					resp.Code = http.StatusUnauthorized
				case utils.ErrNotFound:
					resp.Code = http.StatusNotFound
				default:
					resp.Code = http.StatusInternalServerError
				}
				switch err.(type) {
				case utils.ErrorMessage:
					resp.Msg = err.(error).Error()
				default:
					resp.Msg = "服务器内部错误"
				}
				resp.Exec(w)
			}
		}()
		log.Info(r.RequestURI)
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func loginRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//getCurrentUser(r)
		next(w, r)
	}
}

func adminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//user := getCurrentUser(r)
		//if !user.IsAdmin {
		//	panic(ErrForbidden)
		//}
		next(w, r)
	}
}
