package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

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

func (s *shttp) GetOrCreateTracker(
	ctx context.Context, r *http.Request,
) (string, error) {
	var tracker string = ""
	cookies := r.Cookies()
	for i := 0; i < len(cookies); i++ {
		cookie := cookies[i]
		if cookie.Name == "trackerid" {
			tracker = cookie.Value
			break
		}
	}

	if tracker != "" {
		return tracker, nil
	}

	// This code should / will get executed only if the user is rBot.
	api_key := r.URL.Query()["api_key"]
	if len(api_key) == 0 {
		return "", nil
	}

	user, err := s.GetUser(ctx)
	if err != nil {
		return "", errors.Trace(err)
	}

	tr, err := amalgam.QueryIntoInt(
		ctx,
		`
			SELECT
				id
			FROM
				acko_tracker
			WHERE
				user_id = $1
			LIMIT 1

		`, user.ID(),
	)
	if err != nil {
		if err.Error() != sql.ErrNoRows.Error() {
			return "", errors.Trace(err)
		}

		tid, err := amalgam.QueryIntoInt(
			ctx,
			`
				INSERT INTO
					acko_tracker(user_id, code_version, landing_page,
						initial_ip, is_mobile, is_app, device, os,
						created_on, browser, browser_version, referer,
						user_agent, updated_on)
				VALUES
					($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
				RETURNING
					id

			`, user.ID(), "", "http://127.0.0.1", "127.0.0.1", false, false,
			"api", "Ubuntu", time.Now(), "", "", "", "", time.Now(),
		)
		if err != nil {
			return "", errors.Trace(err)
		}

		return amalgam.EncodeID(int64(tid), "acko_tracker"), nil
	}

	return amalgam.EncodeID(int64(tr), "acko_tracker"), nil
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

	http.Error(w, string(j), 700)
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

	w.WriteHeader(200)
	w.Write(j)
}
