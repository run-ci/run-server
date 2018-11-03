package main

import (
	"fmt"
	"os"

	"github.com/run-ci/run-server/http"
	"github.com/run-ci/run-server/store"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

var pgconnstr string

func init() {
	lvl, err := logrus.ParseLevel(os.Getenv("RUN_LOG_LEVEL"))
	if err != nil {
		lvl = logrus.InfoLevel
	}

	logrus.SetLevel(lvl)

	logger = logrus.WithField("package", "main")

	pguser := os.Getenv("RUN_POSTGRES_USER")
	if pguser == "" {
		logger.Fatal("need RUN_POSTGRES_USER")
	}

	pgpass := os.Getenv("RUN_POSTGRES_PASS")
	if pgpass == "" {
		logger.Fatal("need RUN_POSTGRES_PASS")
	}

	pghref := os.Getenv("RUN_POSTGRES_HREF")
	if pghref == "" {
		logger.Fatal("need RUN_POSTGRES_HREF")
	}

	pgdb := os.Getenv("RUN_POSTGRES_DB")
	if pgdb == "" {
		logger.Fatal("need RUN_POSTGRES_DB")
	}

	pgssl := os.Getenv("RUN_POSTGRES_SSL")
	if pgssl == "" {
		logger.Info("RUN_POSTGRES_SSL not set - defaulting to verify-full")
		pgssl = "verify-full"
	}

	pgconnstr = fmt.Sprintf("postgres://%v:%v@%v/%v?sslmode=%v",
		pguser, pgpass, pghref, pgdb, pgssl)
}

func main() {
	logger.Info("booting server...")

	logger.Info("connecting to database")
	st, err := store.NewPostgres(pgconnstr)
	if err != nil {
		logger.WithField("error", err).Fatal("unable to connect to postgres")
	}

	srv := http.NewServer(":9001", st)

	if err := srv.ListenAndServe(); err != nil {
		logger.WithField("error", err).Fatal("shutting down server")
	}
}
