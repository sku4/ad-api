package server

import (
	"context"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

//nolint:mnd
var (
	InFlightRequests = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "api_tg_bot",
		Subsystem: "http",
		Name:      "in_flight_requests_total",
	})
	SummaryResponseTime = promauto.NewSummary(prometheus.SummaryOpts{
		Namespace: "api_tg_bot",
		Subsystem: "http",
		Name:      "summary_response_time_seconds",
		Objectives: map[float64]float64{
			0.5:  0.1,
			0.9:  0.01,
			0.99: 0.001,
		},
	})
	HistogramResponseTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "api_tg_bot",
			Subsystem: "http",
			Name:      "histogram_response_time_seconds",
			Buckets:   []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2, 4},
		},
		[]string{"cmd"},
	)
)

func Metrics(next IHandler) IHandler {
	hr := Func(func(ctx context.Context, upd tgbotapi.Update) {
		startTime := time.Now()
		next.IncomingMessage(ctx, upd)
		duration := time.Since(startTime)

		SummaryResponseTime.Observe(duration.Seconds())

		var botMessage *tgbotapi.Message
		if upd.Message != nil {
			botMessage = upd.Message
		} else if upd.ChannelPost != nil {
			botMessage = upd.ChannelPost
		}

		cmd := "w/o"
		if botMessage != nil {
			cmd = botMessage.Command()
		}

		HistogramResponseTime.
			WithLabelValues(cmd).
			Observe(duration.Seconds())
	})
	wrappedHandler := InstrumentHandlerInFlight(InFlightRequests, hr)

	return wrappedHandler
}

func InstrumentHandlerInFlight(g prometheus.Gauge, next IHandler) IHandler {
	return Func(func(ctx context.Context, upd tgbotapi.Update) {
		g.Inc()
		defer g.Dec()
		next.IncomingMessage(ctx, upd)
	})
}
