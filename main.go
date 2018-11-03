package main

import (
	"os"

	"github.com/run-ci/run-server/http"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

func init() {
	lvl, err := logrus.ParseLevel(os.Getenv("RUN_LOG_LEVEL"))
	if err != nil {
		lvl = logrus.InfoLevel
	}

	logrus.SetLevel(lvl)

	logger = logrus.WithField("package", "main")
}

func main() {
	logger.Println("booting server...")

	http.ListenAndServe(":9001")
}
