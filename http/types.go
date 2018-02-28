package http

import (
	"context"
	"net/http"

	"github.com/amitu/amalgam"
	"github.com/amitu/amalgam/django"
)

type HTTPService interface {
	ProxyPass(path, dst string)
	Register(string, http.HandlerFunc)
	Reject(w http.ResponseWriter, reason map[string][]amalgam.AError)
	Respond(w http.ResponseWriter, result interface{})
	ListenAndServe(string)
	Redirect(w http.ResponseWriter, r *http.Request, url string, code int)
	GetUser(ctx context.Context) (django.User, error)
	GetOrCreateTracker(context.Context, *http.Request) (string, error)
}
