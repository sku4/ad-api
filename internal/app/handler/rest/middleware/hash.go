package middleware

import (
	"net/http"

	"github.com/sku4/ad-api/configs"
	"github.com/sku4/ad-api/internal/app"
	"github.com/sku4/ad-api/pkg/secret"
)

func ValidHash(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		cfg := configs.Get(r.Context())

		if r.URL.Query().Get(app.ParamChatID) != "" {
			u := r.URL.Query()
			u.Del(app.ParamHash)
			err := secret.IsValid(cfg.Telegram.BotToken, u.Encode(), r.URL.Query().Get(app.ParamHash))
			if err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
