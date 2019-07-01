// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package middleman

import (
	"github.com/Leosocy/IntelliProxy/pkg/loadbalancer"
	"github.com/Leosocy/IntelliProxy/pkg/storage/backend"
	"github.com/Leosocy/IntelliProxy/service/middleman/session"
	"net/http"

	"github.com/elazarl/goproxy"
)

// Server is a middleman between client and real pxy server.
// It run as a https server which always eavesdrop https connections,
// the purpose is to reuse the connection between middleman and the pxy server,
// avoiding TLS handshakes for every request.
//
// And, this is safe because the middleman server is usually deployed
// as a sidecar with crawler program together.
type Server struct {
	sm *session.Manager
	*goproxy.ProxyHttpServer
}

func NewServer(nb backend.NotifyBackend) *Server {
	s := &Server{
		sm:              session.NewManager(nb, loadbalancer.WeightedRoundRobin),
		ProxyHttpServer: goproxy.NewProxyHttpServer(),
	}
	s.Verbose = true
	s.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	s.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (request *http.Request, response *http.Response) {
		if rt, err := s.sm.PickOne(); err == nil {
			ctx.RoundTripper = rt
		}
		return req, nil
	})
	return s
}
