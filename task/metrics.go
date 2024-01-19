package task

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricRunsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "kobold_run_active",
		Help: "number of active runs",
	})
	metricRunStatus = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "kobold_run_status",
		Help: "run status (task groups)",
	}, []string{"status", "repo"})
	metricMsgRecv = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "kobold_msg_recv",
		Help: "number of messages received",
	}, []string{"channel", "rejected"})
	metricGitFetch = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "kobold_git_fetch",
		Help: "number of git fetches",
	}, []string{"repo"})
	metricGitPush = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "kobold_git_push",
		Help: "number of git pushes",
	}, []string{"repo"})
	metricImageSeen = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "kobold_image_seen",
		Help: "number of images seen",
	}, []string{"ref"})
)
