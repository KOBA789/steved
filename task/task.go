package task

import (
	"encoding/json"
	"io/ioutil"
)

type TaskMap map[string]Task

type Task struct {
	Image string
	Cmd   []string
	Slack *string
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

func Spawn(name string, env []string) (*Job, bool, error) {
	tm, err := LoadTaskMap()
	task, ok := tm[name]
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}

	job := Job{
		Name: name,
		Task: task,
		Env: env,
	}
	return &job, true, nil
}
