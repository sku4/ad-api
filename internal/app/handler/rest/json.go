package rest

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"github.com/sku4/ad-parser/pkg/logger"
)

var (
	ErrInternal         = &sentinelAPI{Status: http.StatusInternalServerError, Message: "INTERNAL"}
	ErrChatIDNotSet     = &sentinelAPI{Status: http.StatusForbidden, Message: "CHAT_ID_NOT_SET"}
	StatusOK            = &sentinelAPI{Status: http.StatusOK, Message: "OK"}
	StatusAlreadyExists = &sentinelAPI{Status: http.StatusOK, Message: "ALREADY_EXISTS"}
)

type APIMessage interface {
	APIMessage() (int, string)
}

type sentinelAPI struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (e sentinelAPI) Error() string {
	return e.Message
}

func (e sentinelAPI) APIMessage() (int, string) {
	return e.Status, e.Message
}

func JSONHandleError(w http.ResponseWriter, err error) {
	log := logger.Get()

	var apiErr APIMessage
	if !errors.As(err, &apiErr) {
		apiErr = ErrInternal
		log.Warnf("handle err: %s", err)
	}

	w.Header().Set("Content-Type", "application/json")
	status, _ := apiErr.APIMessage()
	w.WriteHeader(status)

	if err = json.NewEncoder(w).Encode(apiErr); err != nil {
		log.Errorf("handle err: encoder error: %s", err)
		return
	}
}

func JSONHandleOK(w http.ResponseWriter) {
	log := logger.Get()

	var apiOk APIMessage = StatusOK

	w.Header().Set("Content-Type", "application/json")
	status, _ := apiOk.APIMessage()
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(apiOk); err != nil {
		log.Errorf("handle ok: encoder error: %s", err)
		return
	}
}

func JSONHandleResp(w http.ResponseWriter, data []byte) {
	log := logger.Get()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, errWrite := w.Write(data)
	if errWrite != nil {
		log.Errorf("handle resp: write error: %s", errWrite)
		return
	}
}
