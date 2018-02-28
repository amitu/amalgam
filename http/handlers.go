package http

import (
	_ "expvar"
	"net/http"
	_ "net/http/pprof"

	"github.com/amitu/amalgam"
	"github.com/amitu/amalgam/django"
	"github.com/juju/errors"
)

func (s *shttp) sessionAPI(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("key")
	value := r.FormValue("value")
	ctx := r.Context()
	errMap := map[string][]amalgam.AError{}

	if value != "" {
		if key[0] == '_' {
			amalgam.LOGGER.Error("can't_modify_private_keys")
			errMap["__all__"] = append(
				errMap["__all__"],
				amalgam.AError{Human: "Oops something went wrong"},
			)

			s.Reject(ctx, w, errMap)
		}
		err := s.SetSession(ctx, key, value)
		if err != nil {
			amalgam.LOGGER.Error(
				"unable_to_set_session", "err", errors.ErrorStack(err),
			)
			errMap["__all__"] = append(
				errMap["__all__"],
				amalgam.AError{Human: "Oops something went wrong"},
			)

			s.Reject(ctx, w, errMap)
			return
		}
		s.Respond(ctx, w, "ok")
		return
	}

	if key == django.KeyUserID {
		user, err := s.GetUser(ctx)
		if err != nil {
			amalgam.LOGGER.Error(
				"unable_to_get_user", "err", errors.ErrorStack(err),
			)
			errMap["__all__"] = append(
				errMap["__all__"],
				amalgam.AError{Human: "Oops something went wrong"},
			)
			s.Reject(ctx, w, errMap)
			return
		}
		s.Respond(ctx, w, user)
		return
	}

	v, err := s.GetSessionString(ctx, key)
	if err != nil {
		amalgam.LOGGER.Error(
			"unable_to_get_session", "err", errors.ErrorStack(err),
		)
		errMap["__all__"] = append(
			errMap["__all__"],
			amalgam.AError{Human: "Oops something went wrong"},
		)
		s.Reject(ctx, w, errMap)
		return
	}

	s.Respond(ctx, w, v)
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
	s.Register("/testUpload", s.testUploadPage)
}
