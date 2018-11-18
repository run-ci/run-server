package http

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/run-ci/run-server/store"
)

type gitRepoRequest struct {
	Remote string `json:"remote"`
}

type gitRepoResponse struct {
	Remote string `json:"remote"`
}

func (srv *Server) postGitRepo(rw http.ResponseWriter, req *http.Request) {
	reqID := req.Context().Value(keyReqID).(string)
	logger := logger.WithField("request_id", reqID)

	logger.Debug("reading request body")
	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.WithField("error", err).
			Error("unable to read request body")

		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Debug("unmarshaling request body")
	var repo gitRepoRequest
	err = json.Unmarshal(buf, &repo)
	if err != nil {
		logger.WithField("error", err).
			Error("unable to unmarshal request body")

		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Infof("adding git repo for %v", repo.Remote)
	err = srv.st.CreateGitRepo(store.GitRepo{
		Remote: repo.Remote,
	})
	if err != nil {
		logger.WithField("error", err).
			Error("unable to save git repo in database")

		rw.WriteHeader(http.StatusInternalServerError)
		buf, err := json.Marshal(map[string]string{
			"error": err.Error(),
		})
		if err != nil {
			return
		}
		rw.Write(buf)
		return
	}

	// create poller
	err = srv.pcl.createPoller(req.Context(), repo.Remote)
	if err != nil {
		logger.WithError(err).Warn("unable to create poller for git repo")
	}

	resp := gitRepoResponse{
		Remote: repo.Remote,
	}
	buf, err = json.Marshal(resp)
	if err != nil {
		logger.WithField("error", err).
			Error("unable to marshal response body")

		rw.WriteHeader(http.StatusCreated)
		buf, err := json.Marshal(map[string]string{
			"error": err.Error(),
		})
		if err != nil {
			return
		}
		rw.Write(buf)
		return
	}

	rw.WriteHeader(http.StatusCreated)
	rw.Write(buf)
	return
}

func (srv *Server) getGitRepo(rw http.ResponseWriter, req *http.Request) {
	reqID := req.Context().Value(keyReqID).(string)
	logger := logger.WithField("request_id", reqID)

	if _, ok := req.URL.Query()["remote"]; !ok {
		logger.Info("missing 'remote' argument, fetching all repos")

		repos, err := srv.st.GetGitRepos()
		if err != nil {
			logger.WithField("error", err).Error("unable to get git repos from database")

			rw.WriteHeader(http.StatusInternalServerError)
			buf, err := json.Marshal(map[string]string{
				"error": err.Error(),
			})
			if err != nil {
				return
			}
			rw.Write(buf)
			return
		}

		resp := []gitRepoResponse{}
		for _, repo := range repos {
			resp = append(resp, gitRepoResponse{Remote: repo.Remote})
		}

		buf, err := json.Marshal(resp)
		if err != nil {
			logger.WithField("error", err).Error("unable to marshal response body")
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write(buf)
		return
	}
	remote := req.URL.Query()["remote"][0]

	logger = logger.WithField("remote", remote)
	logger.Debug("getting repo")

	repo, err := srv.st.GetGitRepo(remote)
	if err == sql.ErrNoRows {
		logger.WithField("error", err).Error("repo not found in database")
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		logger.WithField("error", err).Error("unable to fetch repo from database")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := gitRepoResponse{
		Remote: repo.Remote,
	}
	buf, err := json.Marshal(resp)
	if err != nil {
		logger.WithField("error", err).Error("unable to marshal response body")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write(buf)
	return
}
