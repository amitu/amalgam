package http

import (
	"net/http"
	"net/http/httputil"

	amalgam "github.com/amitu/amalgam"
)

func (s *shttp) ProxyPass(pth, dst string) {
	director := func(req *http.Request) {
		req.URL.Scheme = "http"
		req.URL.Host = dst
		req.URL.Path = req.URL.Path
	}

	s.mux.Handle(pth, &httputil.ReverseProxy{Director: director})
	amalgam.LOGGER.Debug("registered_proxypass", "path", pth, "remote", dst)
}
