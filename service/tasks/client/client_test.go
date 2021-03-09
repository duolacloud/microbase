package client

import (
	"testing"

	proto "github.com/duolacloud/microbase/proto/tasks"
)

func TestCreateTask(t *testing.T) {
	c := NewTasks()
	res, err := c.Create(&proto.CreateTaskRequest{})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(res)
}
