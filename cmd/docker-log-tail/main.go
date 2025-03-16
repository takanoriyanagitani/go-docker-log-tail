package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/docker/docker/api/types/container"
	dlt "github.com/takanoriyanagitani/go-docker-log-tail"
	. "github.com/takanoriyanagitani/go-docker-log-tail/util"
)

var envValByKey func(string) IO[string] = Lift(
	func(key string) (string, error) {
		val, found := os.LookupEnv(key)
		switch found {
		case true:
			return val, nil
		default:
			return "", fmt.Errorf("env var %s missing", key)
		}
	},
)

var showStdout IO[string] = envValByKey("ENV_SHOW_STDOUT").Or("true")

var showStderr IO[string] = envValByKey("ENV_SHOW_STDERR").Or("true")

var timestamps IO[string] = envValByKey("ENV_SHOW_TIME").Or("true")

var follow IO[string] = envValByKey("ENV_DO_FOLLOW").Or("true")

var details IO[string] = envValByKey("ENV_SHOW_DETAILS").Or("true")

var since IO[string] = envValByKey("ENV_SINCE").Or("")

var until IO[string] = envValByKey("ENV_UNTIL").Or("")

var tail IO[string] = envValByKey("ENV_TAIL").Or("10")

var containerId IO[string] = func(_ context.Context) (string, error) {
	var sz int = len(os.Args)
	switch sz {
	case 0:
		return "", fmt.Errorf("unexpected args: %v", os.Args)
	case 1:
		return "", fmt.Errorf("container id missing: %v", os.Args)
	case 2:
		return os.Args[1], nil
	default:
		return "", fmt.Errorf("too many arguments: %v", sz)
	}
}

var cfg IO[dlt.ShowLogOption] = Bind(
	AllMap(map[string]IO[string]{
		"stdout":     showStdout,
		"stderr":     showStderr,
		"timestamps": timestamps,
		"follow":     follow,
		"details":    details,
		"since":      since,
		"until":      until,
		"tail":       tail,
	}),
	func(m map[string]string) IO[dlt.ShowLogOption] {
		var stdout IO[bool] = Bind(Of(m["stdout"]), Lift(strconv.ParseBool))
		var stderr IO[bool] = Bind(Of(m["stderr"]), Lift(strconv.ParseBool))
		var timestamps IO[bool] = Bind(Of(m["timestamps"]), Lift(strconv.ParseBool))
		var follow IO[bool] = Bind(Of(m["follow"]), Lift(strconv.ParseBool))
		var details IO[bool] = Bind(Of(m["details"]), Lift(strconv.ParseBool))
		return Bind(
			AllMap(map[string]IO[bool]{
				"stdout":     stdout,
				"stderr":     stderr,
				"timestamps": timestamps,
				"follow":     follow,
				"details":    details,
			}),
			Lift(func(mb map[string]bool) (dlt.ShowLogOption, error) {
				return dlt.ShowLogOption{
					LogsOptions: container.LogsOptions{
						ShowStdout: mb["stdout"],
						ShowStderr: mb["stderr"],
						Since:      m["since"],
						Until:      m["until"],
						Timestamps: mb["timestamps"],
						Follow:     mb["follow"],
						Tail:       m["tail"],
						Details:    mb["details"],
					},
				}, nil
			}),
		)
	},
)

var logs2stdoe IO[Void] = Bind(
	containerId,
	func(id string) IO[Void] {
		return Bind(
			cfg,
			func(cfg dlt.ShowLogOption) IO[Void] {
				return func(ctx context.Context) (Void, error) {
					cli, e := dlt.DockerClientDefault()
					if nil != e {
						return Empty, e
					}
					defer cli.Close()

					e = cfg.DemuxLogToStd(
						ctx,
						cli,
						id,
					)
					return Empty, e
				}
			},
		)
	},
)

var sub IO[Void] = func(ctx context.Context) (Void, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	return logs2stdoe(ctx)
}

func main() {
	_, e := sub(context.Background())
	if nil != e {
		log.Printf("%v\n", e)
	}
}
