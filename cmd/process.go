package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/spf13/cobra"

	datadog "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/safchain/systress/pkg/process"
)

var (
	count    int64
	depth    int64
	wait     int64
	argsLen  int64
	argsSize int64
	envsLen  int64
	envsSize int64
	child    bool
)

var processCmd = &cobra.Command{
	Use:   "process",
	Short: "process tool box",
	Long:  `Bunch of tool to generate load on a system in a timed box manner`,
}

var forkExecCmd = &cobra.Command{
	Use:   "fork-exec",
	Short: "generate fork exec",
	Long:  `generate a chain of fork/exec`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if depth == 0 {
			return
		}

		executable, err := os.Executable()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		opts := process.ExecOpts{
			Wait:     time.Duration(wait) * time.Millisecond,
			ArgsLen:  argsLen,
			ArgsSize: argsSize,
			EnvsLen:  envsLen,
			EnvsSize: envsSize,
		}

		fmt.Printf("Start for a duration of %d\n", duration)

		if !child {
			ctx, _ := context.WithTimeout(context.Background(), time.Duration(duration)*time.Second)
			ctx = datadog.NewDefaultContext(ctx)

			configuration := datadog.NewConfiguration()
			apiClient := datadog.NewAPIClient(configuration)

			var lock sync.Mutex
			body := datadog.MetricsPayload{
				Series: []datadog.Series{
					{
						Metric: "cws.systress.process.count",
						Type:   datadog.PtrString("count"),
						Points: [][]*float64{},
						Tags: &[]string{
							"benchmark: cws-process",
						},
					},
				},
			}

			var total int64
		LOOP:
			for {
				select {
				case <-ctx.Done():
					break LOOP
				default:
					var wg sync.WaitGroup
					for i := int64(0); i != count; i++ {
						wg.Add(1)

						go func() {
							lock.Lock()
							total += depth
							lock.Unlock()

							err = process.ExecAndWait(ctx, opts, executable, "--duration", strconv.FormatInt(duration, 10), "process", "fork-exec", "--wait", strconv.FormatInt(wait, 10), "--depth", strconv.FormatInt(depth-1, 10), "--child")
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
								os.Exit(1)
							}
							wg.Done()

							lock.Lock()
							points := body.Series[0].Points
							body.Series[0].Points = append(points, []*float64{
								datadog.PtrFloat64(float64(time.Now().Unix())),
								datadog.PtrFloat64(float64(total)),
							})
							lock.Unlock()
						}()
					}

					wg.Wait()

					_, _, err := apiClient.MetricsApi.SubmitMetrics(ctx, body, *datadog.NewSubmitMetricsOptionalParameters())
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error when calling `MetricsApi.SubmitMetrics`: %v\n", err)
					}

					lock.Lock()
					body.Series[0].Points = body.Series[0].Points[0:0]
					lock.Unlock()
				}
			}
		} else {
			err = process.ExecAndWait(context.Background(), opts, executable, "--duration", strconv.FormatInt(duration, 10), "process", "fork-exec", "--wait", strconv.FormatInt(wait, 10), "--depth", strconv.FormatInt(depth-1, 10), "--child")
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	processCmd.AddCommand(forkExecCmd)

	forkExecCmd.PersistentFlags().Int64VarP(&wait, "wait", "", 0, "specify duration(ms) to wait after exec")
	forkExecCmd.PersistentFlags().BoolVarP(&child, "child", "", false, "internal used")
	forkExecCmd.PersistentFlags().Int64VarP(&count, "count", "", 1, "number of first level child")
	forkExecCmd.PersistentFlags().Int64VarP(&depth, "depth", "", 1, "specify the number of child")
	forkExecCmd.PersistentFlags().Int64VarP(&argsLen, "args-len", "", 16, "specify number of arguments")
	forkExecCmd.PersistentFlags().Int64VarP(&argsSize, "args-size", "", 16, "specify the size of each argument")
	forkExecCmd.PersistentFlags().Int64VarP(&envsLen, "envs-len", "", 8, "specify the number of environment variable")
	forkExecCmd.PersistentFlags().Int64VarP(&envsSize, "envs-size", "", 16, "specify the size of each environment variable")
}
