package task

import (
	"encoding/json"
	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
	"io/ioutil"
)

type TaskMap map[string]Task

type Task struct {
	Image string
	Cmd   []string
	Slack *string
}

func (t *Task) buildSlackPayload(title *string, text *string, color *string) slack.Payload {
	attachment := slack.Attachment {
		Title: title,
		Text: text,
		Color: color,
	}
	return slack.Payload {
		Username: "steved",
		Attachments: []slack.Attachment{attachment},
	}
}

func (t *Task) buildSlackPayloadStarted(name string) slack.Payload {
	title := "[STARTED] " + name
	color := "#28d7e5"
	return t.buildSlackPayload(&title, nil, &color)
}

func (t *Task) buildSlackPayloadFailed(name string, err error) slack.Payload {
	title := "[FAILED] " + name
	text := err.Error()
	color := "danger"
	return t.buildSlackPayload(&title, &text, &color)
}

func (t *Task) buildSlackPayloadSucceed(name string) slack.Payload {
	title := "[SUCCEED] " + name
	color := "good"
	return t.buildSlackPayload(&title, nil, &color)
}

func (t *Task) NotifyToSlack(payload slack.Payload) {
	if t.Slack == nil {
		return
	}
	slack.Send(*t.Slack, "", payload)
}

func (t *Task) Spawn(name string, env []string) error {
	t.NotifyToSlack(t.buildSlackPayloadStarted(name))
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		t.NotifyToSlack(t.buildSlackPayloadFailed(name, err))
		return err
	}

	_, err = cli.ImagePull(ctx, t.Image, types.ImagePullOptions{})
	if err != nil {
		t.NotifyToSlack(t.buildSlackPayloadFailed(name, err))
		return err
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: t.Image,
		Cmd:   t.Cmd,
		Env:   env,
	}, &container.HostConfig{
		AutoRemove: true,
	}, nil, "steved-job-"+name)
	if err != nil {
		t.NotifyToSlack(t.buildSlackPayloadFailed(name, err))
		return err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		t.NotifyToSlack(t.buildSlackPayloadFailed(name, err))
		return err
	}

	t.NotifyToSlack(t.buildSlackPayloadSucceed(name))
	return nil
}

func LoadTaskMap() (TaskMap, error) {
	file, err := ioutil.ReadFile("./tasks.json")
	if err != nil {
		return TaskMap{}, err
	}
	taskMap := TaskMap{}
	json.Unmarshal(file, &taskMap)
	return taskMap, nil
}

func GetTask(name string) (Task, bool, error) {
	taskMap, err := LoadTaskMap()
	if err != nil {
		return Task{}, false, err
	}

	task, ok := taskMap[name]
	if !ok {
		return Task{}, false, nil
	}

	return task, true, nil
}
