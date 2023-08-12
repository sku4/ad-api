package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sku4/ad-api/internal/app"
	"github.com/sku4/ad-api/model"
	modelAd "github.com/sku4/ad-parser/pkg/ad/model"
	"github.com/sku4/ad-parser/pkg/logger"
)

func (h *Handler) AddSub(w http.ResponseWriter, r *http.Request) {
	log := logger.Get()

	var err error
	isPrivate := false
	if isPrivateStr := r.URL.Query().Get(app.ParamIsPrivate); isPrivateStr != "" {
		isPrivate, err = strconv.ParseBool(isPrivateStr)
		if err != nil {
			JSONHandleError(w, err)
			return
		}
	}

	body, err := h.services.Bot.WebAppAddSubscriptionIndex(r.Context(), isPrivate, r.URL.RawQuery)
	if err != nil {
		log.Warnf("add sub: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(body)
	if err != nil {
		log.Warnf("add sub: write error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

type streetsResp struct {
	Status  int             `json:"status"`
	Message string          `json:"message"`
	Result  []*model.Street `json:"result"`
}

func (h *Handler) Streets(w http.ResponseWriter, r *http.Request) {
	streets, err := h.services.WebAppStreets(r.Context())
	if err != nil {
		JSONHandleError(w, err)
		return
	}

	streetsJSON, err := json.Marshal(streetsResp{
		Status:  http.StatusOK,
		Result:  streets,
		Message: "OK",
	})
	if err != nil {
		JSONHandleError(w, err)
		return
	}

	JSONHandleResp(w, streetsJSON)
}

func (h *Handler) SubscriptionAdd(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get(app.ParamChatID) == "" {
		JSONHandleError(w, ErrChatIDNotSet)
		return
	}

	var subTnt *modelAd.SubscriptionTnt
	if err := json.NewDecoder(r.Body).Decode(&subTnt); err != nil {
		JSONHandleError(w, err)
		return
	}

	if !subTnt.Valid() {
		JSONHandleError(w, ErrInternal)
		return
	}

	chatID, err := strconv.ParseInt(r.URL.Query().Get(app.ParamChatID), 10, 0)
	if err != nil {
		JSONHandleError(w, err)
		return
	}
	subTnt.TelegramID = chatID

	if err = h.services.WebAppAddSubscription(r.Context(), subTnt); err != nil {
		if errors.Is(err, model.ErrSubAlreadyExists) {
			JSONHandleError(w, StatusAlreadyExists)
		} else {
			JSONHandleError(w, err)
		}
		return
	}

	JSONHandleOK(w)
}
