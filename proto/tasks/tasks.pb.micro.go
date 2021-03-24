// Code generated by protoc-gen-micro. DO NOT EDIT.
// source: proto/tasks/tasks.proto

package tasks

import (
	fmt "fmt"
	_ "github.com/duolacloud/microbase/proto/api"
	proto "github.com/golang/protobuf/proto"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	math "math"
)

import (
	context "context"
	api "github.com/micro/go-micro/v2/api"
	client "github.com/micro/go-micro/v2/client"
	server "github.com/micro/go-micro/v2/server"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ api.Endpoint
var _ context.Context
var _ client.Option
var _ server.Option

// Api Endpoints for TaskService service

func NewTaskServiceEndpoints() []*api.Endpoint {
	return []*api.Endpoint{}
}

// Client API for TaskService service

type TaskService interface {
	Create(ctx context.Context, in *CreateTaskRequest, opts ...client.CallOption) (*CreateTaskResponse, error)
	Update(ctx context.Context, in *UpdateTaskRequest, opts ...client.CallOption) (*UpdateTaskResponse, error)
	Get(ctx context.Context, in *GetTaskRequest, opts ...client.CallOption) (*GetTasksResponse, error)
	Delete(ctx context.Context, in *UpdateTaskRequest, opts ...client.CallOption) (*UpdateTaskResponse, error)
	List(ctx context.Context, in *ListTasksRequest, opts ...client.CallOption) (*ListTasksResponse, error)
}

type taskService struct {
	c    client.Client
	name string
}

func NewTaskService(name string, c client.Client) TaskService {
	return &taskService{
		c:    c,
		name: name,
	}
}

func (c *taskService) Create(ctx context.Context, in *CreateTaskRequest, opts ...client.CallOption) (*CreateTaskResponse, error) {
	req := c.c.NewRequest(c.name, "TaskService.Create", in)
	out := new(CreateTaskResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskService) Update(ctx context.Context, in *UpdateTaskRequest, opts ...client.CallOption) (*UpdateTaskResponse, error) {
	req := c.c.NewRequest(c.name, "TaskService.Update", in)
	out := new(UpdateTaskResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskService) Get(ctx context.Context, in *GetTaskRequest, opts ...client.CallOption) (*GetTasksResponse, error) {
	req := c.c.NewRequest(c.name, "TaskService.Get", in)
	out := new(GetTasksResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskService) Delete(ctx context.Context, in *UpdateTaskRequest, opts ...client.CallOption) (*UpdateTaskResponse, error) {
	req := c.c.NewRequest(c.name, "TaskService.Delete", in)
	out := new(UpdateTaskResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taskService) List(ctx context.Context, in *ListTasksRequest, opts ...client.CallOption) (*ListTasksResponse, error) {
	req := c.c.NewRequest(c.name, "TaskService.List", in)
	out := new(ListTasksResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for TaskService service

type TaskServiceHandler interface {
	Create(context.Context, *CreateTaskRequest, *CreateTaskResponse) error
	Update(context.Context, *UpdateTaskRequest, *UpdateTaskResponse) error
	Get(context.Context, *GetTaskRequest, *GetTasksResponse) error
	Delete(context.Context, *UpdateTaskRequest, *UpdateTaskResponse) error
	List(context.Context, *ListTasksRequest, *ListTasksResponse) error
}

func RegisterTaskServiceHandler(s server.Server, hdlr TaskServiceHandler, opts ...server.HandlerOption) error {
	type taskService interface {
		Create(ctx context.Context, in *CreateTaskRequest, out *CreateTaskResponse) error
		Update(ctx context.Context, in *UpdateTaskRequest, out *UpdateTaskResponse) error
		Get(ctx context.Context, in *GetTaskRequest, out *GetTasksResponse) error
		Delete(ctx context.Context, in *UpdateTaskRequest, out *UpdateTaskResponse) error
		List(ctx context.Context, in *ListTasksRequest, out *ListTasksResponse) error
	}
	type TaskService struct {
		taskService
	}
	h := &taskServiceHandler{hdlr}
	return s.Handle(s.NewHandler(&TaskService{h}, opts...))
}

type taskServiceHandler struct {
	TaskServiceHandler
}

func (h *taskServiceHandler) Create(ctx context.Context, in *CreateTaskRequest, out *CreateTaskResponse) error {
	return h.TaskServiceHandler.Create(ctx, in, out)
}

func (h *taskServiceHandler) Update(ctx context.Context, in *UpdateTaskRequest, out *UpdateTaskResponse) error {
	return h.TaskServiceHandler.Update(ctx, in, out)
}

func (h *taskServiceHandler) Get(ctx context.Context, in *GetTaskRequest, out *GetTasksResponse) error {
	return h.TaskServiceHandler.Get(ctx, in, out)
}

func (h *taskServiceHandler) Delete(ctx context.Context, in *UpdateTaskRequest, out *UpdateTaskResponse) error {
	return h.TaskServiceHandler.Delete(ctx, in, out)
}

func (h *taskServiceHandler) List(ctx context.Context, in *ListTasksRequest, out *ListTasksResponse) error {
	return h.TaskServiceHandler.List(ctx, in, out)
}
