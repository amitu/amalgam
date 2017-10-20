package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/amitu/amalgam"
	"github.com/amitu/amalgam/django"
	"github.com/juju/errors"
)

type shttp struct {
	mux      *http.ServeMux
	addr     string
	proxies  map[string]string
	ctx      context.Context
	sessions django.SessionStore
}

func NewHTTPService(
	addr string, ctx context.Context, sessions django.SessionStore,
) HTTPService {
	h := &shttp{
		http.NewServeMux(), addr, make(map[string]string), ctx, sessions,
	}
	h.register()
	return h
}

func (s *shttp) ListenAndServe(listen string) {
	http.ListenAndServe(listen, s)
}

func (s *shttp) Redirect(w http.ResponseWriter, r *http.Request, url string, code int) {
	http.Redirect(w, r, url, code)
}

func (s *shttp) Register(pattern string, fn http.HandlerFunc) {
	amalgam.LOGGER.Debug("registering pattern", "pattern", pattern)
	s.mux.HandleFunc(pattern, fn)
}

type EResult struct {
	Result  interface{}                 `json:"result,omitempty"`
	Errors  map[string][]amalgam.AError `json:"errors,omitempty"`
	Success bool                        `json:"success"`
}

func (s *shttp) Reject(
	w http.ResponseWriter,
	reason map[string][]amalgam.AError,
) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	j, err := json.Marshal(&EResult{Errors: reason, Success: false})
	if err != nil {
		amalgam.LOGGER.Error(
			"reject_json_failed", "err", errors.ErrorStack(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Error(w, string(j), http.StatusOK)
}

func (s *shttp) Respond(w http.ResponseWriter, result interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	j, err := json.Marshal(&EResult{Result: result, Success: true})
	if err != nil {
		amalgam.LOGGER.Error(
			"respond_json_failed", "err", errors.ErrorStack(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(j)
}
