package httpd

import (
	"net"
	"net/http"
	"strings"

	"github.com/adamveld12/gittp"
	. "github.com/adamveld12/goku"
	"github.com/adamveld12/muxwrap"
)

func New(config Configuration, backend Backend) (*HttpService, error) {
	cfg := gittp.ServerConfig{
		Path:        config.GitPath,
		PreReceive:  gittp.UseGithubRepoNames,
		PostReceive: NewPushHandler(config),
		Debug:       true,
	}

	if config.MasterOnly {
		cfg.PreReceive = gittp.CombinePreHooks(gittp.UseGithubRepoNames, gittp.MasterOnly)
	}

	hl := NewLog("[http]", config.Debug)
	gittpHandler, err := gittp.NewGitServer(cfg)
	if err != nil {
		hl.Error(err)
		return nil, err
	}

	hl.Trace("setting up git handlers")
	gitHandler := muxwrap.New( /* BasicAuth(handleAuth) */ )
	gitHandler.Handle("/", gittpHandler.ServeHTTP)

	return &HttpService{
		Log:        hl,
		config:     config,
		gitHandler: gitHandler,
		backend:    backend,
	}, nil
}

type HttpService struct {
	Log
	config     Configuration
	gitHandler http.Handler
	backend    Backend
	api        http.Handler
	l          net.Listener
}

func (h *HttpService) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	h.Tracef("%v %v", req.Method, req.URL)

	if strings.HasPrefix("/api/v1", req.URL.Path) {
		h.api.ServeHTTP(res, req)
	} else {
		h.gitHandler.ServeHTTP(res, req)
	}
}

func (h *HttpService) Start() error {
	addr := h.config.HTTP

	h.Trace("starting http server")
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	h.l = l
	go func(h *HttpService) {
		s := http.Server{Handler: h}
		h.Trace("serving git on ", addr)

		if err := s.Serve(h.l); err != nil {
			// TODO if we get an error here, the daemon is hosed so we'll just panic
			h.Fatal(err)
		}
	}(h)

	return nil
}

func (h *HttpService) Stop() error {
	if h.l != nil {
		h.l.Close()
	}
	return nil
}
