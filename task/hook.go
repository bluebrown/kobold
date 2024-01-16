package task

import (
	"fmt"
	"os"
	"strings"

	"go.starlark.net/starlark"

	"github.com/bluebrown/kobold/starutil"
	"github.com/bluebrown/kobold/store"
)

type StarlarkHookRunner struct {
	hostEnv *starlark.Dict
}

var _ HookRunner = (*StarlarkHookRunner)(nil)

func NewStarlarkPostHook() *StarlarkHookRunner {
	return &StarlarkHookRunner{
		hostEnv: starutil.EnvToStarlarkDict(os.Environ()),
	}
}

func (d *StarlarkHookRunner) Run(group store.TaskGroup, msg string, changes []string, warnings []string) error {
	if group.PostHook == nil {
		return nil
	}

	res, err := starutil.RunMain(starutil.DefaultThread(group.Fingerprint), "post_hook", group.PostHook, d.args(group, msg, changes, warnings), d.hostEnv)
	if err != nil {
		return fmt.Errorf("run main: %w", err)
	}

	if res != starlark.None {
		return fmt.Errorf("post_hook returned %s", res.String())
	}

	return nil
}

func (runner *StarlarkHookRunner) args(group store.TaskGroup, msg string, changes []string, warnings []string) starlark.Tuple {
	title, body, ok := strings.Cut(msg, "\n")
	if !ok {
		title = msg
	}

	body = strings.TrimSpace(body)

	r := starlark.String(group.RepoUri.Repo)

	sb := starlark.String(group.RepoUri.Ref)

	var db starlark.Value
	if group.DestBranch.Valid {
		db = starlark.String(group.DestBranch.String)
	} else {
		db = starlark.String(group.RepoUri.Ref)
	}

	t := starlark.String(title)
	b := starlark.String(body)

	ch := starlark.NewList([]starlark.Value{})
	for _, c := range changes {
		if err := ch.Append(starlark.String(c)); err != nil {
			panic(err)
		}
	}

	warns := starlark.NewList([]starlark.Value{})
	for _, w := range warnings {
		if err := warns.Append(starlark.String(w)); err != nil {
			panic(err)
		}
	}

	return starlark.Tuple([]starlark.Value{r, sb, db, t, b, ch, warns})
}
