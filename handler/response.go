package handler

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func (r *Response) Exec(w http.ResponseWriter) {
	var b []byte
	var err error
	if r.Data == nil {
		b, err = json.Marshal(struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}{r.Code, r.Msg})
	} else {
		b, err = json.Marshal(struct {
			Code int         `json:"code"`
			Data interface{} `json:"data"`
		}{r.Code, r.Data})
	}
	if err != nil {
		panic(err)
	}
	w.WriteHeader(r.Code)
	_, err = w.Write(b)
	if err != nil {
		panic(err)
	}
}

func msgResponse(w http.ResponseWriter, code int, msg string) {
	resp := &Response{
		Code: code,
		Msg:  msg,
	}
	resp.Exec(w)
}

func dataResponse(w http.ResponseWriter, data interface{}) {
	resp := &Response{
		Code: http.StatusOK,
		Data: data,
	}
	resp.Exec(w)
}
