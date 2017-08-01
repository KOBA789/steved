package task

import (
	"encoding/json"
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
}

func (t *Task) Spawn(name string, env []string) error {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	_, err = cli.ImagePull(ctx, t.Image, types.ImagePullOptions{})
	if err != nil {
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
		return err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

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
