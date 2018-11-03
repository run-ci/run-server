package http

import (
	"context"
	"net/http"

	"github.com/run-ci/run-server/store"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

type ctxkey int

const (
	keyReqID ctxkey = iota
)

func init() {
	logger = logrus.WithField("package", "http")
}

// Server is a net/http.Server with dependencies like
// the database connection.
type Server struct {
	st store.Repo

	*http.Server
}

// NewServer returns a Server with a reference to `st`, listening
// on `addr`.
func NewServer(addr string, st store.Repo) *Server {
	srv := &Server{
		Server: &http.Server{
			Addr: addr,
		},

		st: st,
	}

	r := mux.NewRouter()
	srv.Handler = r

	r.Handle("/", chain(getRoot, setRequestID, logRequest)).
		Methods(http.MethodGet)

	r.Handle("/repos/git", chain(srv.postGitRepo, setRequestID, logRequest)).
		Methods(http.MethodPost)

	return srv
}

// Middleware is a function that can intercept the handling of an HTTP request
// to do something useful.
type middleware func(http.HandlerFunc) http.HandlerFunc

// Chain builds the final http.Handler from all the middlewares passed to it.
func chain(f http.HandlerFunc, mw ...middleware) http.Handler {
	// Because function calls are placed on a stack, they need to
	// be applied in reverse order from what they are passed in,
	// in order for calls to Chain() to be intuitive.
	for i := len(mw) - 1; i >= 0; i-- {
		f = mw[i](f)
	}

	return f
}

// SetRequestID sets a UUID on the request so that it can be tracked through
// logs, metrics and instrumentation.
func setRequestID(f http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		id := uuid.New().String()

		ctx := context.WithValue(req.Context(), keyReqID, id)
		logger.WithField("request_id", id).
			Debug("setting request ID")

		f(rw, req.WithContext(ctx))
	}
}

// LogRequest logs useful information about the request. It must have a
// "request_id" set on the request context.
func logRequest(f http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		reqid := req.Context().Value(keyReqID).(string)

		logger := logger.WithField("request_id", reqid)

		logger.Infof("%v %v", req.Method, req.URL)

		f(rw, req)
	}
}
