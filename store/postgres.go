package store

import (
	"database/sql"

	_ "github.com/lib/pq" // load the postgres driver
)

// Postgres is a RepoStore backed by PostgreSQL.
type Postgres struct {
	db *sql.DB
}

// NewPostgres returns a RepoStore backed by PostgreSQL. It connects to the
// database using connstr.
func NewPostgres(connstr string) (Repo, error) {
	logger = logger.WithField("store", "postgres")

	logger.Debug("connecting to database")

	db, err := sql.Open("postgres", connstr)
	if err != nil {
		logger.WithField("error", err).Debug("unable to connect to database")
		return nil, err
	}

	return &Postgres{
		db: db,
	}, nil
}

func (pg *Postgres) CreateGitRepo(repo GitRepo) (GitRepo, error) {
	logger.Debugf("creating git repo for %v", repo.Remote)

	return GitRepo{}, nil
}
