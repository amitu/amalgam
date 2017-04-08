package http

import (
	"context"
	"net/http"

	"acko/django"
)

type HTTPService interface {
	ProxyPass(path, dst string)
	Register(string, http.HandlerFunc)
	Reject(w http.ResponseWriter, reason string)
	Respond(w http.ResponseWriter, result interface{})
	ListenAndServe(string)
	Redirect(w http.ResponseWriter, r *http.Request, url string, code int)
	GetUser(ctx context.Context) (django.User, error)
}
