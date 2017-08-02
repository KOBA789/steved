package task

import (
	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"golang.org/x/net/context"
	"log"
	"os"
	"errors"
)

type Job struct {
	Name string
	Task Task
	Env  []string
}

func buildSlackPayload(title *string, text *string, color *string) slack.Payload {
	attachment := slack.Attachment{
		Title: title,
		Text:  text,
		Color: color,
	}
	return slack.Payload{
		Username:    "steved",
		Attachments: []slack.Attachment{attachment},
	}
}

func (j *Job) buildSlackPayloadStarted() slack.Payload {
	title := "[STARTED] " + j.Name
	color := "#28d7e5"
	return buildSlackPayload(&title, nil, &color)
}

func (j *Job) buildSlackPayloadFailed(err error) slack.Payload {
	title := "[FAILED] " + j.Name
	text := err.Error()
	color := "danger"
	return buildSlackPayload(&title, &text, &color)
}

func (j *Job) buildSlackPayloadSucceed() slack.Payload {
	title := "[SUCCEED] " + j.Name
	color := "good"
	return buildSlackPayload(&title, nil, &color)
}

func (j *Job) notifyToSlack(payload slack.Payload) {
	if j.Task.Slack == nil {
		return
	}
	err := slack.Send(*j.Task.Slack, "", payload)
	if err != nil {
		log.Println(err)
	}
}

func (j *Job) notifyError(err error) {
	j.notifyToSlack(j.buildSlackPayloadFailed(err))
}

func (j *Job) notifyStarting() {
	j.notifyToSlack(j.buildSlackPayloadStarted())
}

func (j *Job) notifySuccess() {
	j.notifyToSlack(j.buildSlackPayloadSucceed())
}

func (j *Job) Run() error {
	j.notifyStarting()
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		j.notifyError(err)
		return err
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: j.Task.Image,
		Cmd:   j.Task.Cmd,
		Env:   j.Env,
	}, &container.HostConfig{
		AutoRemove: true,
	}, nil, "steved-job-"+j.Name)
	if err != nil {
		j.notifyError(err)
		return err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		j.notifyError(err)
		return err
	}

	go func() {
		out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow: true,
			Tail: "all",
		})
		if err != nil {
			log.Println(err)
			return
		}
		stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	}()

	go func() {
		okC, errC := cli.ContainerWait(ctx, resp.ID, "not-running")
		select {
		case ok := <-okC:
			if ok.StatusCode == 0 {
				j.notifySuccess()
			} else {
				j.notifyError(errors.New("Returns non-zero exit code: "+string(ok.StatusCode)))
			}
		case err := <-errC:
			log.Println(err)
			j.notifyError(err)
		}
	}()

	return nil
}
