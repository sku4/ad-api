package bot

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sku4/ad-parser/pkg/logger"
)

const (
	statTimeout = time.Second * 20
)

var (
	usersTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "api",
		Subsystem: "users",
		Name:      "total",
	})
	subsTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "api",
		Subsystem: "subscriptions",
		Name:      "total",
	})
)

func (b *Bot) SetUsersMetrics() {
	go func() {
		log := logger.Get()
		ctx, cancel := context.WithTimeout(context.Background(), statTimeout)
		defer cancel()

		stat, err := b.repos.Subscription.GetUsersStat(ctx)
		if err != nil {
			log.Errorf("set user metrics err: %s", err)
			return
		}

		usersTotal.Set(float64(stat.UsersCount))
		subsTotal.Set(float64(stat.SubscriptionCount))
	}()
}
