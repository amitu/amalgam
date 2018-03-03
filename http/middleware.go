package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/amitu/amalgam"
	"github.com/amitu/amalgam/django"
	"github.com/getsentry/raven-go"
	"github.com/inconshreveable/log15"
	"github.com/jmoiron/sqlx"
	"github.com/juju/errors"
)

var (
	ErrNoSession              = errors.New("no session in request")
	ErrBadSession             = errors.New("bad session")
	ErrSessionItemIsNotString = errors.New("session item is not string")
)

func (s *shttp) GetSession(ctx context.Context) (django.Session, error) {
	sessionid, err := amalgam.Ctx2SessionKey(ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return s.sessions.GetSessionBySessionKey(ctx, sessionid)
}

func (s *shttp) SetSession(
	ctx context.Context, key string, value interface{},
) error {
	session, err := s.GetSession(ctx)
	if err != nil && errors.Cause(err) != sql.ErrNoRows {
		return errors.Trace(err)
	}

	return errors.Trace(session.SetValue(ctx, key, value))
}

func (s *shttp) GetSessionString(ctx context.Context, key string) (string, error) {
	session, err := s.GetSession(ctx)
	if err != nil {
		return "", errors.Trace(err)
	}

	return session.GetString(key)
}

func (s *shttp) GetSessionInt64(ctx context.Context, key string) (int64, error) {
	session, err := s.GetSession(ctx)
	if err != nil {
		return 0, errors.Trace(err)
	}

	return session.GetInt64(key)
}

func (s *shttp) GetUser(ctx context.Context) (django.User, error) {
	session, err := s.GetSession(ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}

	user, err := session.GetUser(ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return user, nil
}

type CodeWriter struct {
	*sqlx.Tx
	code     int
	hasError bool
	http.ResponseWriter
}

func (c *CodeWriter) WriteHeader(code int) {
	if code > 399 {
		err := c.Tx.Rollback()
		if err != nil {
			amalgam.LOGGER.Crit(
				"failed_to_rollback_transaction", "err", errors.ErrorStack(err),
			)
		}
	} else {
		err := c.Tx.Commit()
		if err != nil {
			amalgam.LOGGER.Crit(
				"failed_to_commit_transaction", "err", errors.ErrorStack(err),
			)
			c.hasError = true
			err := c.Tx.Rollback()
			if err != nil {
				amalgam.LOGGER.Crit(
					"failed_to_rollback_transaction",
					"err", errors.ErrorStack(err),
				)
			}
		}
	}
	c.code = 200
	c.ResponseWriter.WriteHeader(code)
}

func (c *CodeWriter) Write(resp []byte) (int, error) {
	if c.hasError {
		errMap := map[string][]amalgam.AError{}
		errMap["__all__"] = append(
			errMap["__all__"],
			amalgam.AError{Human: "Oops something went wrong"},
		)
		body, _ := json.Marshal(&EResult{Errors: errMap, Success: false})
		return c.ResponseWriter.Write(body)
	}

	return c.ResponseWriter.Write(resp)
}

func (s *shttp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	if colon := strings.LastIndex(clientIP, ":"); colon != -1 {
		clientIP = clientIP[:colon]
	}

	db, err := amalgam.Ctx2Db(s.ctx)
	if err != nil {
		amalgam.LOGGER.Crit(
			"failed_to_get_db", "err", errors.ErrorStack(err),
		)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	tx, err := db.Beginx()
	if err != nil {
		amalgam.LOGGER.Crit(
			"failed_to_create_transaction", "err", errors.ErrorStack(err),
		)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	ctx := context.WithValue(r.Context(), amalgam.KeyDBTransaction, tx)
	ctx = context.WithValue(ctx, amalgam.KeyDB, db)

	w2 := &CodeWriter{tx, 200, false, w}

	start := time.Now()
	logger := amalgam.LOGGER.New(
		"url", r.RequestURI, "method", r.Method, "ip", clientIP,
	)
	logger.Debug("http_started")
	logger = logger.New(
		"time", log15.Lazy{func() interface{} {
			return time.Since(start)
		}},
		"code", log15.Lazy{func() interface{} {
			return w2.code
		}},
	)

	defer func() {
		if err := recover(); err != nil {
			err2, ok := err.(error)
			if ok {
				logger.Error(
					"server_error", "err", errors.ErrorStack(err2),
				)
			} else {
				logger.Error(
					"server_uerror", "err", err, "ip", clientIP,
				)
			}

			err3 := tx.Rollback()
			if err3 != nil {
				logger.Crit("server_tx_error", "err", errors.ErrorStack(err3))
			}

			errMap := map[string][]amalgam.AError{}
			errMap["__all__"] = append(
				errMap["__all__"],
				amalgam.AError{Human: "Oops something went wrong!"},
			)

			res := &EResult{Errors: errMap, Success: false}
			m, err := json.Marshal(res)
			http.Error(w, string(m), 500)

			if (!amalgam.Debug) && (amalgam.Sentry) != "" {
				// raven/sentry stuff
				rvalStr := fmt.Sprint(err)
				packet := raven.NewPacket(
					rvalStr,
					raven.NewException(
						errors.New(rvalStr),
						raven.NewStacktrace(2, 3, nil),
					),
					raven.NewHttp(r),
				)
				raven.Capture(packet, nil)
			}
		}
	}()

	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-XSS-Protection", "1; mode=block")

	errMap := map[string][]amalgam.AError{}

	if amalgam.UseSession {
		sessionid, err := r.Cookie("sessionid")
		if err != nil {
			if err == http.ErrNoCookie {
				sid, err := s.sessions.CreateSession(ctx)
				if err != nil {
					logger.Crit(
						"session_creation_error",
						"err",
						errors.ErrorStack(errors.Trace(err)),
					)
					errMap["__all__"] = append(
						errMap["__all__"],
						amalgam.AError{Human: "Oops something went wrong"},
					)
					s.Reject(w, errMap)
					return
				}

				sessionid = &http.Cookie{
					Name: "sessionid", Value: sid.SessionKey(), Path: "/",
				}
				http.SetCookie(w, sessionid)
			}
		}

		ctx = context.WithValue(ctx, amalgam.KeySession, sessionid.Value)
		s.mux.ServeHTTP(w2, r.WithContext(ctx))

		logger.Debug("http_served")
	}
}
