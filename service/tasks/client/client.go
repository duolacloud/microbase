package client

import (
	proto "github.com/duolacloud/microbase/proto/tasks"
	"github.com/duolacloud/microbase/service/tasks"
	"github.com/micro/micro/v2/service/client"
	"github.com/micro/micro/v2/service/context"
)

var (
	name = "tasks"
)

type srv struct {
	client proto.TaskService
}

func (m *srv) Create(req *proto.CreateTaskRequest, options ...tasks.CreateTaskOption) (*proto.CreateTaskResponse, error) {
	o := tasks.CreateTaskOptions{}
	for _, option := range options {
		option(&o)
	}

	res, err := m.client.Create(context.DefaultContext, &proto.CreateTaskRequest{}, client.WithAuthToken())
	if err != nil {
		return nil, err
	}

	return res, nil
}

func NewTasks() *srv {
	addr := name

	s := &srv{
		client: proto.NewTaskService(addr, client.DefaultClient),
	}

	return s
}
