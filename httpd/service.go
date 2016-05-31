package httpd

import (
	"encoding/base64"
	"net"
	"net/http"
	"strings"

	"github.com/adamveld12/gittp"
	. "github.com/adamveld12/goku"
	"github.com/adamveld12/muxwrap"
)

func init() {
	RegisterService(newHTTPd)
}

func newHTTPd(config Configuration, backend Backend) Service {
	cfg := gittp.ServerConfig{
		Path:        config.GitPath,
		PreReceive:  gittp.UseGithubRepoNames,
		PostReceive: newPushHandler(config.Hostname, config.DockerSock, config.Debug),
		Debug:       config.Debug,
	}

	if config.MasterOnly {
		cfg.PreReceive = gittp.CombinePreHooks(gittp.UseGithubRepoNames, gittp.MasterOnly)
	}

	hl := NewLog("[http]", config.Debug)
	gittpHandler, _ := gittp.NewGitServer(cfg)

	gitHandler := muxwrap.New( /* basicAuth(handleAuth) */ )
	gitHandler.Handle("/", gittpHandler.ServeHTTP)

	return &httpService{
		Log:        hl,
		addr:       config.HTTP,
		gitHandler: gitHandler,
		backend:    backend,
	}
}

type httpService struct {
	Log
	addr       string
	gitHandler http.Handler
	backend    Backend
	api        http.Handler
	l          net.Listener
}

func (h *httpService) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	h.Tracef("%v %v", req.Method, req.URL)

	if strings.HasPrefix("/api/v1", req.URL.Path) {
		h.api.ServeHTTP(res, req)
	} else {
		h.gitHandler.ServeHTTP(res, req)
	}
}

func (h *httpService) Start() error {
	h.Trace("listening for git on ", h.addr)
	l, err := net.Listen("tcp", h.addr)
	if err != nil {
		return err
	}

	h.l = l
	go func(h *httpService) {
		s := http.Server{Handler: h}

		if err := s.Serve(h.l); err != nil {
			// TODO if we get an error here, the daemon is hosed so we'll just panic
			h.Fatal(err)
		}
	}(h)

	return nil
}

func (h *httpService) Stop() error {
	h.l.Close()
	return nil
}

func basicAuth(auth func(username, pass string) error) muxwrap.Middleware {
	return muxwrap.Middleware(func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.Header().Set("WWW-Authenticate", `Basic realm="Goku"`)

			s := strings.SplitN(req.Header.Get("Authorization"), " ", 2)
			if len(s) != 2 {
				http.Error(res, "Not authorized", 401)
				return
			}

			b, err := base64.StdEncoding.DecodeString(s[1])
			if err != nil {
				http.Error(res, err.Error(), 401)
				return
			}

			pair := strings.SplitN(string(b), ":", 2)
			if len(pair) != 2 {
				http.Error(res, "Not authorized", 401)
				return
			}

			if err := auth(pair[0], pair[1]); err != nil {
				http.Error(res, "Not authorized", 401)
				return
			}

			next.ServeHTTP(res, req)
		})

	})
}
