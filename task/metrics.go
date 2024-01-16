package task

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricRun = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "kobold_run",
		Help: "run status (task groups)",
	}, []string{"status", "repo"})
	metricMsgRecv = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "kobold_recv",
		Help: "number of messages received",
	}, []string{"channel", "rejected"})
)
