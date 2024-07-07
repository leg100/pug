package task

import (
	"testing"

	"github.com/leg100/pug/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTaskCreator struct{}

func (f *fakeTaskCreator) Create(spec CreateOptions) (*Task, error) {
	return (&factory{}).newTask(spec)
}

func TestNewGroupWithDependencies(t *testing.T) {
	// Create module dependency tree
	vpc := resource.New(resource.Module, resource.GlobalResource)
	mysql := resource.New(resource.Module, resource.GlobalResource).WithDependencies(vpc)
	redis := resource.New(resource.Module, resource.GlobalResource).WithDependencies(vpc)
	backend := resource.New(resource.Module, resource.GlobalResource).WithDependencies(vpc, redis, mysql)
	frontend := resource.New(resource.Module, resource.GlobalResource).WithDependencies(backend, vpc)
	mq := resource.New(resource.Module, resource.GlobalResource)

	vpcSpec := CreateOptions{Parent: vpc, Path: "vpc"}
	mysqlSpec := CreateOptions{Parent: mysql, Path: "mysql"}
	redisSpec := CreateOptions{Parent: redis, Path: "redis"}
	backendSpec := CreateOptions{Parent: backend, Path: "backend"}
	frontendSpec := CreateOptions{Parent: frontend, Path: "frontend"}
	mqSpec := CreateOptions{Parent: mq, Path: "mq"}

	got, err := NewGroupWithDependencies(&fakeTaskCreator{}, "apply",
		vpcSpec,
		mysqlSpec,
		redisSpec,
		backendSpec,
		frontendSpec,
		mqSpec,
	)
	require.NoError(t, err)

	if assert.Len(t, got.Tasks, 6) {
		vpcTask := hasTask(t, got.Tasks, "vpc")
		mysqlTask := hasTask(t, got.Tasks, "mysql", vpcTask)
		redisTask := hasTask(t, got.Tasks, "redis", vpcTask)
		backendTask := hasTask(t, got.Tasks, "backend", vpcTask, mysqlTask, redisTask)
		_ = hasTask(t, got.Tasks, "frontend", vpcTask, backendTask)
		_ = hasTask(t, got.Tasks, "mq")
	}
}

func hasTask(t *testing.T, got []*Task, path string, deps ...resource.ID) resource.ID {
	for _, task := range got {
		if task.Path == path {
			// Module matches, so now check dependencies
			if assert.Len(t, task.DependsOn, len(deps)) {
				for _, dep := range deps {
					assert.Contains(t, task.DependsOn, dep)
				}
				return task.ID
			}
		}
	}
	t.Fatalf("%s not found in %v", path, got)
	return resource.ID{}
}
