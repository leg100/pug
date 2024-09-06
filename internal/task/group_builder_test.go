package task

import (
	"testing"

	"github.com/leg100/pug/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTaskCreator struct{}

func (f *fakeTaskCreator) Create(spec Spec) (*Task, error) {
	return (&factory{}).newTask(spec)
}

func TestNewGroupWithDependencies(t *testing.T) {
	vpcID := resource.NewID(resource.Module)
	mysqlID := resource.NewID(resource.Module)
	redisID := resource.NewID(resource.Module)
	backendID := resource.NewID(resource.Module)
	frontendID := resource.NewID(resource.Module)
	mqID := resource.NewID(resource.Module)

	vpcSpec := Spec{ModuleID: &vpcID, Path: "vpc", Dependencies: &Dependencies{}}
	mysqlSpec := Spec{ModuleID: &mysqlID, Path: "mysql", Dependencies: &Dependencies{ModuleIDs: []resource.ID{vpcID}}}
	redisSpec := Spec{ModuleID: &redisID, Path: "redis", Dependencies: &Dependencies{ModuleIDs: []resource.ID{vpcID}}}
	backendSpec := Spec{ModuleID: &backendID, Path: "backend", Dependencies: &Dependencies{ModuleIDs: []resource.ID{vpcID, mysqlID, redisID}}}
	frontendSpec := Spec{ModuleID: &frontendID, Path: "frontend", Dependencies: &Dependencies{ModuleIDs: []resource.ID{vpcID, backendID}}}
	mqSpec := Spec{ModuleID: &mqID, Path: "mq", Dependencies: &Dependencies{}}

	t.Run("normal order", func(t *testing.T) {
		got, err := createDependentTasks(&fakeTaskCreator{}, false,
			vpcSpec,
			mysqlSpec,
			redisSpec,
			backendSpec,
			frontendSpec,
			mqSpec,
		)
		require.NoError(t, err)

		if assert.Len(t, got, 6) {
			vpcTask := hasDependencies(t, got, "vpc") // 0 dependencies
			mysqlTask := hasDependencies(t, got, "mysql", vpcTask)
			redisTask := hasDependencies(t, got, "redis", vpcTask)
			backendTask := hasDependencies(t, got, "backend", vpcTask, mysqlTask, redisTask)
			_ = hasDependencies(t, got, "frontend", vpcTask, backendTask)
			_ = hasDependencies(t, got, "mq")
		}
	})

	t.Run("reverse order", func(t *testing.T) {
		got, err := createDependentTasks(&fakeTaskCreator{}, true,
			vpcSpec,
			mysqlSpec,
			redisSpec,
			backendSpec,
			frontendSpec,
			mqSpec,
		)
		require.NoError(t, err)

		if assert.Len(t, got, 6) {
			frontendTask := hasDependencies(t, got, "frontend") // 0 dependencies
			backendTask := hasDependencies(t, got, "backend", frontendTask)
			mysqlTask := hasDependencies(t, got, "mysql", backendTask)
			redisTask := hasDependencies(t, got, "redis", backendTask)
			_ = hasDependencies(t, got, "vpc", mysqlTask, redisTask, backendTask, frontendTask)
			_ = hasDependencies(t, got, "mq")
		}
	})
}

func hasDependencies(t *testing.T, got []*Task, path string, deps ...resource.ID) resource.ID {
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
