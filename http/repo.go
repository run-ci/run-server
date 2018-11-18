package http

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/run-ci/run-server/store"
	"github.com/sirupsen/logrus"
)

type gitRepoRequest struct {
	Remote string `json:"remote"`
	Branch string `json:"branch"`
}

type gitRepoResponse struct {
	Remote string `json:"remote"`
	Branch string `json:"branch"`
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

	if repo.Branch == "" {
		repo.Branch = "master"
	}

	logger = logger.WithFields(logrus.Fields{
		"remote": repo.Remote,
		"branch": repo.Branch,
	})

	logger.Info("adding git repo")
	err = srv.st.CreateGitRepo(store.GitRepo{
		Remote: repo.Remote,
		Branch: repo.Branch,
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

	err = srv.pcl.createPoller(req.Context(), repo.Remote, repo.Branch)
	if err != nil {
		logger.WithError(err).Warn("unable to create poller for git repo")
	}

	resp := gitRepoResponse{
		Remote: repo.Remote,
		Branch: repo.Branch,
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
			resp = append(resp, gitRepoResponse{
				Remote: repo.Remote,
				Branch: repo.Branch,
			})
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

	branch := "master"
	if _, ok := req.URL.Query()["branch"]; ok {
		branch = req.URL.Query()["branch"][0]
	}

	logger.Infof("using %v as branch", branch)

	logger = logger.WithFields(logrus.Fields{
		"remote": remote,
		"branch": branch,
	})

	logger.Debug("getting repo")

	repo, err := srv.st.GetGitRepo(remote, branch)
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
