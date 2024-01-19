package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/bluebrown/kobold/krm"
	"github.com/bluebrown/kobold/store"
)

// the task handler is the final point of execution. after decoding, debouncing
// and aggregating the events, this handler is resonbible for the actual work
func KoboldHandler(ctx context.Context, cache string, g store.TaskGroup, runner HookRunner) ([]string, error) {
	var (
		changes     []string
		warnings    []string
		pipelineErr error
		hookErr     error
		msg         string
	)

	// TODO: this should be emitted to the monitoring system
	defer func(ts time.Time) {
		slog.InfoContext(ctx,
			"pipeline run completed",
			"fingerprint", g.Fingerprint,
			"changes", len(changes),
			"warnings", len(warnings),
			"pipelineErr", pipelineErr,
			"hookErr", hookErr,
			"elapsed", time.Since(ts).String(),
		)
	}(time.Now())

	if g.DestBranch.Valid {
		g.DestBranch.String = g.DestBranch.String + "-" + g.Fingerprint
	}

	pipline := krm.Pipeline{
		RepoURI:   g.RepoUri.String(),
		SrcBranch: g.RepoUri.Ref,
		DstBranch: g.DestBranch.String,
		CachePath: cache,
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	msg, changes, warnings, pipelineErr = pipline.Run(ctx, g.Msgs)
	if pipelineErr != nil {
		return nil, fmt.Errorf("pipeline: %w", pipelineErr)
	}

	if runner == nil || len(changes) == 0 {
		return warnings, nil
	}

	hookErr = runner.Run(g, msg, changes, warnings)
	if hookErr != nil {
		return warnings, fmt.Errorf("hook: %w", hookErr)
	}

	return warnings, nil
}

var _ TaskHandler = KoboldHandler

func PrintHandler(ctx context.Context, hostPath string, g store.TaskGroup, runner HookRunner) ([]string, error) {
	b, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal task group: %w", err)
	}
	fmt.Println(string(b))
	return nil, nil
}

var _ TaskHandler = PrintHandler

func ThrowHandler(ctx context.Context, hostPath string, g store.TaskGroup, runner HookRunner) ([]string, error) {
	return nil, fmt.Errorf("throw handler error")
}

var _ TaskHandler = ThrowHandler
