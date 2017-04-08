package http

import (
	"net/http"
	"net/http/httputil"

	"acko"
)

func (s *shttp) ProxyPass(pth, dst string) {
	director := func(req *http.Request) {
		req.URL.Scheme = "http"
		req.URL.Host = dst
		req.URL.Path = req.URL.Path
	}

	s.mux.Handle(pth, &httputil.ReverseProxy{Director: director})
	acko.LOGGER.Debug("registered_proxypass", "path", pth, "remote", dst)
}
