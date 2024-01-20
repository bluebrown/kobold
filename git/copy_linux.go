package git

import (
	"context"
	"fmt"
	"os/exec"
)

func copy(ctx context.Context, src, dst string) error {
	cmd := exec.CommandContext(ctx, "cp", "-r", src, dst)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cp: %w: %s", err, string(b))
	}
	return nil
}
