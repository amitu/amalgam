package http

import (
	_ "expvar"
	"net/http"
	_ "net/http/pprof"

	"acko/django"

	"github.com/juju/errors"
)

func (s *shttp) sessionAPI(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("key")
	value := r.FormValue("value")
	ctx := r.Context()

	if value != "" {
		if key[0] == '_' {
			s.Reject(w, "cant modify private keys")
		}
		err := s.SetSession(ctx, key, value)
		if err != nil {
			s.Reject(w, errors.ErrorStack(err))
			return
		}
		s.Respond(w, "ok")
		return
	}

	if key == django.KeyUserID {
		user, err := s.GetUser(ctx)
		if err != nil {
			s.Reject(w, errors.ErrorStack(err))
			return
		}
		s.Respond(w, user)
		return
	}

	v, err := s.GetSessionString(ctx, key)
	if err != nil {
		s.Reject(w, errors.ErrorStack(err))
		return
	}

	s.Respond(w, v)
}

func (s *shttp) elmPage(w http.ResponseWriter, r *http.Request) {
	_, err := s.GetUser(r.Context())
	if err != nil {
		s.Redirect(w, r, "admin/", http.StatusFound)
		return
	}
	w.Write(
		[]byte(`
			<!DOCTYPE html>
			<html>
				<head>
					<meta charset="utf-8" />
					<meta content="width=device-width,
						  initial-scale=1.0" name="viewport" />
					<title>r2d2</title>
					<link href="/static/style.css" rel="stylesheet"
					      type="text/css" />
				</head>
				<body data-csrf="asd"><script src="/static/elm.js"></script></body>
				<h1>
			</html>
		`),
	)
}

func (s *shttp) testUploadPage(w http.ResponseWriter, _ *http.Request) {
	w.Write(
		[]byte(`
			<!DOCTYPE html>
			<html>
				<body>
					<form method="POST" action="/claims/attachment/" enctype="multipart/form-data">
					Task ID <input type="text" name="id">
					Field Name <input type="text" name="field">
					Start <input type="text" value="0" name="start">
					UUID <input type="text" value="0" name="uuid">
					File: <input type="file" name="file">
					<input type="submit">
					</form>
				</body>
			</html>
		`),
	)
}

func (s *shttp) register() {
	s.mux.Handle("/debug/", http.DefaultServeMux)
	s.Register("/_session", s.sessionAPI)
	s.Register("/", s.elmPage)
	s.Register("/testUpload", s.testUploadPage)
}
