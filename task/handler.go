package task

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/bluebrown/kobold/git"
	"github.com/bluebrown/kobold/krm"
	"github.com/bluebrown/kobold/store/model"
	"github.com/prometheus/client_golang/prometheus"
)

// the task handler is the final point of execution. after decoding, debouncing
// and aggregating the events, this handler is resonbible for the actual work
func KoboldHandler(ctx context.Context, cache string, g model.TaskGroup, runner HookRunner) ([]string, error) {
	var (
		changes  []string
		warnings []string
		msg      string
	)

	if err := git.Switch(ctx, cache, g.RepoUri.Ref); err != nil {
		return nil, fmt.Errorf("git switch: %#q => %#q: %w", g.RepoUri.Repo, g.RepoUri.Ref, err)
	}

	changes, warnings, err := krm.Pipeline(ctx, filepath.Join(cache, g.RepoUri.Pkg), g.Msgs...)
	if err != nil {
		return nil, fmt.Errorf("krm pipeline: %w", err)
	}

	if len(changes) < 1 {
		return nil, nil
	}

	if g.DestBranch.Valid {
		g.DestBranch.String = g.DestBranch.String + "-" + g.Fingerprint
		if err := git.CheckoutB(ctx, cache, g.DestBranch.String); err != nil {
			return nil, fmt.Errorf("git checkout -b: %w", err)
		}
	} else {
		g.DestBranch.String = g.RepoUri.Ref
		g.DestBranch.Valid = true
	}

	msg = "chore(kobold): update image refs"

	if err := git.Publish(ctx, cache, g.DestBranch.String, msg); err != nil {
		return nil, fmt.Errorf("git publish: %w", err)
	}

	metricGitPush.With(prometheus.Labels{"repo": g.RepoUri.Repo}).Inc()

	if runner == nil || len(changes) == 0 {
		return warnings, nil
	}

	if err := runner.Run(g, msg, changes, warnings); err != nil {
		return warnings, fmt.Errorf("hook: %w", err)
	}

	return warnings, nil
}

var _ Handler = KoboldHandler

func PrintHandler(ctx context.Context, hostPath string, g model.TaskGroup, runner HookRunner) ([]string, error) {
	b, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal task group: %w", err)
	}
	fmt.Println(string(b))
	return nil, nil
}

var _ Handler = PrintHandler

func ThrowHandler(ctx context.Context, hostPath string, g model.TaskGroup, runner HookRunner) ([]string, error) {
	return nil, fmt.Errorf("throw handler error")
}

var _ Handler = ThrowHandler
