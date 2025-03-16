package logtail

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

func Demux(
	wout io.Writer,
	werr io.Writer,
	mux io.Reader,
) (written int64, err error) {
	return stdcopy.StdCopy(wout, werr, mux)
}

func DockerClientDefault() (*client.Client, error) {
	return client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
}

type ShowLogOption struct {
	container.LogsOptions
}

func (o ShowLogOption) DemuxLog(
	ctx context.Context,
	client *client.Client,
	containerId string,
	wout io.Writer,
	werr io.Writer,
) error {
	rcloser, e := client.ContainerLogs(
		ctx,
		containerId,
		o.LogsOptions,
	)
	if nil != e {
		return e
	}
	defer rcloser.Close()

	_, e = Demux(
		wout,
		werr,
		rcloser,
	)

	return e
}

func (o ShowLogOption) DemuxLogToStd(
	ctx context.Context,
	client *client.Client,
	containerId string,
) error {
	return o.DemuxLog(
		ctx,
		client,
		containerId,
		os.Stdout,
		os.Stderr,
	)
}
