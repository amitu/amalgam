package http

import (
	"context"
	"net/http"

	"github.com/amitu/amalgam"
	"github.com/amitu/amalgam/django"
)

type fhttp struct {
}

func (f *fhttp) ListenAndServe(string) {
	panic("not implemented")
}

func (f *fhttp) ProxyPass(path, dst string) {
	panic("not implemented")
}

func (f *fhttp) Register(pattern string, fn http.HandlerFunc) {
	panic("not implemented")
}

func (f *fhttp) Redirect(w http.ResponseWriter, r *http.Request, url string, code int) {
	panic("not implemented")
}

func (f *fhttp) Reject(
	ctx context.Context, w http.ResponseWriter, reason map[string][]amalgam.AError,
) {
	panic("not implemented")
}

func (f *fhttp) Respond(
	ctx context.Context, w http.ResponseWriter, result interface{},
) {
	panic("not implemented")
}

func (f *fhttp) GetUser(ctx context.Context) (django.User, error) {
	panic("not implemented")
	return nil, nil
}

func (f *fhttp) GetOrCreateTracker(
	ctx context.Context, r *http.Request,
) (string, error) {
	panic("not implemented")
}

func (f *fhttp) Success(w http.ResponseWriter, result interface{}, api string) {
	panic("not implemented")
}

func NewFakeHTTPService() HTTPService {
	return &fhttp{}
}
