package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/run-ci/run-server/store"
)

type memStore struct {
	db map[string]store.GitRepo
}

func (st *memStore) CreateGitRepo(repo store.GitRepo) error {
	key := fmt.Sprintf("%v#%v", repo.Remote, repo.Branch)
	st.db[key] = repo

	return nil
}

func (st *memStore) GetGitRepo(remote, branch string) (store.GitRepo, error) {

	return store.GitRepo{}, nil
}

func (st *memStore) GetGitRepos() ([]store.GitRepo, error) {

	return []store.GitRepo{}, nil
}

func TestPostGitRepo(t *testing.T) {
	send := make(chan []byte)
	st := &memStore{
		db: make(map[string]store.GitRepo),
	}
	srv := NewServer(":9001", send, st)

	repo := gitRepoRequest{
		Remote: "test",
		Branch: "master",
	}
	payload, err := json.Marshal(repo)
	if err != nil {
		t.Fatalf("got error when marshaling request payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "http://test/repos/git", bytes.NewBuffer(payload))
	req = req.WithContext(context.WithValue(context.Background(), keyReqID, "test"))
	rw := httptest.NewRecorder()

	srv.postGitRepo(rw, req)

	resp := rw.Result()

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected status %v, got %v", http.StatusAccepted, resp.StatusCode)
	}

	// This will time out if the request to create the poller wasn't sent. This
	// timeout should fail the test.
	rawmsg := <-send
	plrmsg := map[string]string{}
	err = json.Unmarshal(rawmsg, &plrmsg)
	if err != nil {
		t.Fatalf("got error unmarshalling poller message: %v", err)
	}

	if op, ok := plrmsg["op"]; !ok || op != "create" {
		t.Fatalf(`expected "op" to be set to "create", got %v`, op)
	}

	if remote, ok := plrmsg["remote"]; !ok || remote != "test" {
		t.Fatalf(`expected "remote" to be set to "test", got %v`, remote)
	}

	if branch, ok := plrmsg["branch"]; !ok || branch != "master" {
		t.Fatalf(`expected "branch" to be set to "master", got %v`, branch)
	}
}
