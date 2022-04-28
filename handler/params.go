package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Jeffail/gabs/v2"
	"github.com/gorilla/mux"

	"zuccacm-server/db"
)

var (
	defaultBeginTime = parseDate("2000-01-01")
	defaultEndTime   = parseDate("2100-01-01")
)

func parseDate(t string) time.Time {
	ret, err := time.ParseInLocation("2006-01-02", t, time.Local)
	if err != nil {
		panic(ErrBadRequest.Wrap(err))
	}
	return ret
}

func decodePage(r *http.Request) (p db.Page) {
	p.PageIndex = getParamInt(r, "page_index", 0)
	p.PageSize = getParamInt(r, "page_size", 0)
	return
}

// ----------------- params from req.Body in json format -----------------

type Params gabs.Container

func (params *Params) getInt(path string) int {
	x, err := params.get(path).(json.Number).Int64()
	if err != nil {
		panic(ErrBadRequest.Wrap(err))
	}
	return int(x)
}

func (params *Params) getString(path string) string {
	return params.get(path).(string)
}

func (params *Params) get(path string) interface{} {
	p := (*gabs.Container)(params)
	if !p.Exists(path) {
		panic(ErrBadRequest.New())
	}
	return p.Path(path).Data()
}

func decodeParam(body io.ReadCloser) *Params {
	b, err := io.ReadAll(body)
	if err != nil {
		panic(ErrBadRequest.Wrap(err))
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	p, err := gabs.ParseJSONDecoder(dec)
	if err != nil {
		panic(ErrBadRequest.Wrap(err))
	}
	return (*Params)(p)
}

func decodeParamVar(r *http.Request, to interface{}) {
	err := json.NewDecoder(r.Body).Decode(to)
	if err != nil {
		panic(ErrBadRequest.Wrap(err))
	}
}

// ----------------------- params from URL.Query() -----------------------
// For example, '/users?is_enable=true'
func getParam(r *http.Request, key string, defaultValue string) string {
	if !r.URL.Query().Has(key) {
		return defaultValue
	}
	return r.URL.Query().Get(key)
}

func getParamRequired(r *http.Request, key string) string {
	if !r.URL.Query().Has(key) {
		panic(ErrBadRequest.New())
	}
	return r.URL.Query().Get(key)
}

func getParamInt(r *http.Request, key string, defaultValue int) int {
	if !r.URL.Query().Has(key) {
		return defaultValue
	}
	x, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil {
		panic(err)
	}
	return x
}

func getParamDate(r *http.Request, key string, defaultValue time.Time) time.Time {
	if !r.URL.Query().Has(key) {
		return defaultValue
	}
	return parseDate(r.URL.Query().Get(key))
}

func getParamDateRequired(r *http.Request, key string) time.Time {
	return parseDate(getParamRequired(r, key))
}

func getParamBool(r *http.Request, key string, defaultValue bool) bool {
	if !r.URL.Query().Has(key) {
		return defaultValue
	}
	v, err := strconv.ParseBool(r.URL.Query().Get(key))
	if err != nil {
		panic(ErrBadRequest.Wrap(err))
	}
	return v
}

// ------------------------ params from URL.Path -------------------------
// For example, '/contest/{id}'
func getParamURL(r *http.Request, key string) string {
	vars := mux.Vars(r)
	x, ok := vars[key]
	if !ok {
		panic(ErrBadRequest.New())
	}
	return x
}

func getParamIntURL(r *http.Request, key string) int {
	x, err := strconv.Atoi(getParamURL(r, key))
	if err != nil {
		panic(ErrBadRequest.Wrap(err))
	}
	return x
}
