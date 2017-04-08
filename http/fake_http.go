package http

import (
	"context"
	"net/http"

	"acko/django"
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

func (f *fhttp) Reject(w http.ResponseWriter, reason string) {
	panic("not implemented")
}

func (f *fhttp) Respond(w http.ResponseWriter, result interface{}) {
	panic("not implemented")
}

func (f *fhttp) GetUser(ctx context.Context) (django.User, error) {
	panic("not implemented")
	return nil, nil
}

func NewFakeHTTPService() HTTPService {
	return &fhttp{}
}
