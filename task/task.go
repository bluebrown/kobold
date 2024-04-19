package task

import (
	"context"
	"fmt"

	"github.com/bluebrown/kobold/krm"
	"github.com/bluebrown/kobold/store/model"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusRunning Status = "running"
	StatusSuccess Status = "success"
	StatusFailure Status = "failure"
)

type DecoderRunner interface {
	Decode(name string, script []byte, data []byte) ([]string, error)
}

type HookRunner interface {
	Run(group model.TaskGroup, msg string, changes []krm.Change, warnings []string) error
}

type Handler func(ctx context.Context, hostPath string, g model.TaskGroup, hook HookRunner) ([]string, error)

func (t *Handler) String() string {
	return fmt.Sprintf("%T", *t)
}

func (t *Handler) Set(s string) error {
	switch s {
	case "kobold":
		*t = KoboldHandler
	case "error":
		*t = ThrowHandler
	case "print":
		*t = PrintHandler
	default:
		return fmt.Errorf("unknown task handler: %s", s)
	}
	return nil
}
