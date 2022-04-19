package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Jeffail/gabs/v2"
	"github.com/gorilla/mux"

	"zuccacm-server/utils"
)

var (
	defaultBeginTime = parseTime("2000-01-01")
	defaultEndTime   = parseTime("2100-01-01")
)

// Params get json
type Params gabs.Container

func (params *Params) getInt(path string) int {
	x, err := params.get(path).(json.Number).Int64()
	if err != nil {
		panic(utils.ErrBadRequest)
	}
	return int(x)
}

func (params *Params) getString(path string) string {
	return params.get(path).(string)
}

func (params *Params) get(path string) interface{} {
	p := (*gabs.Container)(params)
	if !p.Exists(path) {
		panic(utils.ErrBadRequest)
	}
	return p.Path(path).Data()
}

func decodeParam(r *http.Request) *Params {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		panic(utils.ErrBadRequest)
	}
	p, err := gabs.ParseJSON(b)
	if err != nil {
		panic(utils.ErrBadRequest)
	}
	return (*Params)(p)
}

func decodeParamVar(r *http.Request, to interface{}) {
	err := json.NewDecoder(r.Body).Decode(to)
	if err != nil {
		panic(utils.ErrBadRequest)
	}
}

func getParam(r *http.Request, key string, defaultValue string) string {
	if !r.URL.Query().Has(key) {
		return defaultValue
	}
	return r.URL.Query().Get(key)
}

func getParamTime(r *http.Request, key string, defaultValue time.Time) time.Time {
	if !r.URL.Query().Has(key) {
		return defaultValue
	}
	return parseTime(r.URL.Query().Get(key))
}

func getParamBool(r *http.Request, key string, defaultValue bool) bool {
	if !r.URL.Query().Has(key) {
		return defaultValue
	}
	v, err := strconv.ParseBool(r.URL.Query().Get(key))
	if err != nil {
		panic(utils.ErrBadRequest)
	}
	return v
}

// getParamIntURL get parameter (type is int) from URL, like '/contest_group/1'
func getParamIntURL(r *http.Request, key string) int {
	vars := mux.Vars(r)
	x, err := strconv.Atoi(vars[key])
	if err != nil {
		panic(utils.ErrBadRequest)
	}
	return x
}

func parseTime(t string) time.Time {
	ret, err := time.ParseInLocation("2006-01-02", t, time.Local)
	if err != nil {
		panic(utils.ErrBadRequest)
	}
	return ret
}
