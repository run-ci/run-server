package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type gitTriggerRequest struct {
	Remote string `json:"remote"`
}

type gitTriggerResponse struct {
	Remote string `json:"remote"`
}

func postGitTrigger(rw http.ResponseWriter, req *http.Request) {
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
	var trigger gitTriggerRequest
	err = json.Unmarshal(buf, &trigger)
	if err != nil {
		logger.WithField("error", err).
			Error("unable to unmarshal request body")

		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Infof("adding git trigger for %v", trigger.Remote)

	resp := gitTriggerResponse{
		Remote: trigger.Remote,
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
	}

	rw.WriteHeader(http.StatusCreated)
	rw.Write(buf)
	return
}
