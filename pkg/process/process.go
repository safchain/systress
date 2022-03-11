package process

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/safchain/systress/pkg/utils"
)

type ExecOpts struct {
	Wait     time.Duration
	ArgsLen  int64
	ArgsSize int64
	EnvsLen  int64
	EnvsSize int64
}

func ExecAndWait(ctx context.Context, opts ExecOpts, name string, arg ...string) error {
	for i := 0; i != int(opts.ArgsLen); i++ {
		arg = append(arg, utils.RandString(int(opts.ArgsSize)))
	}

	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.Env = []string{}

	for i := 0; i != int(opts.EnvsLen); i++ {
		env := fmt.Sprintf("%s=%s", utils.RandString(int(opts.EnvsSize/2)), utils.RandString(int(opts.EnvsSize/2)))
		cmd.Env = append(cmd.Env, env)
	}

	if err := exec.CommandContext(ctx, name, arg...).Run(); err != nil {
		return err
	}

	time.Sleep(opts.Wait)

	return nil
}
