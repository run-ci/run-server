package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

type pollerRequest struct {
	Remote string `json:"remote"`
	Branch string `json:"branch"`
}

type pollClient struct {
	url string

	client *http.Client
}

func (c *pollClient) createPoller(ctx context.Context, remote string) error {
	reqID := ctx.Value(keyReqID).(string)

	logger := logger.WithFields(logrus.Fields{
		"remote":     remote,
		"request_id": reqID,
	})

	if c.url == "" {
		return errors.New("poller server url not set")
	}

	logger.Debug("creating poller")

	endpoint := fmt.Sprintf("%v/pollers", c.url)

	data := pollerRequest{
		Remote: remote,
	}

	buf, err := json.Marshal(data)
	if err != nil {
		logger.WithError(err).Debug("unable to marshal poller request data")
		return err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(buf))
	if err != nil {
		logger.WithError(err).Debug("unable to create http request")
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		logger.WithError(err).Debug("error requesting poller create")
		return err
	}

	if resp.StatusCode != http.StatusAccepted {
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.WithError(err).Debug("unable to read poller server error response")
			return err
		}

		errbody := map[string]string{}
		err = json.Unmarshal(buf, &errbody)
		if err != nil {
			logger.WithError(err).Debug("unable to unmarshal poller server error response")
			return err
		}

		return errors.New(errbody["error"])
	}

	logger.Debug("poller created")

	return nil
}
