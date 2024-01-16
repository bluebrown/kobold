package task

import (
	"context"
	"fmt"

	"github.com/bluebrown/kobold/store"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusRunning Status = "running"
	StatusSuccess Status = "success"
	StatusFailure Status = "failure"
)

type Decoder interface {
	Decode(name string, script []byte, data []byte) ([]string, error)
}

type HookRunner interface {
	Run(group store.TaskGroup, msg string, changes []string, warnings []string) error
}

type TaskHandler func(ctx context.Context, g store.TaskGroup, hook HookRunner) ([]string, error)

// implement the flag.Value interface
func (t *TaskHandler) String() string {
	return fmt.Sprintf("%T", *t)
}

func (t *TaskHandler) Set(s string) error {
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
