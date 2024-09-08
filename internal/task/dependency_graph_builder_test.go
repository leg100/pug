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

	vpcSpec := Spec{ModuleID: &vpcID, Dependencies: &Dependencies{}}
	mysqlSpec := Spec{ModuleID: &mysqlID, Dependencies: &Dependencies{ModuleIDs: []resource.ID{vpcID}}}
	redisSpec := Spec{ModuleID: &redisID, Dependencies: &Dependencies{ModuleIDs: []resource.ID{vpcID}}}
	backendSpec := Spec{ModuleID: &backendID, Dependencies: &Dependencies{ModuleIDs: []resource.ID{vpcID, mysqlID, redisID}}}
	frontendSpec := Spec{ModuleID: &frontendID, Dependencies: &Dependencies{ModuleIDs: []resource.ID{vpcID, backendID}}}
	mqSpec := Spec{ModuleID: &mqID, Dependencies: &Dependencies{}}

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
			vpcTask := hasDependencies(t, got, vpcID) // 0 dependencies
			mysqlTask := hasDependencies(t, got, mysqlID, vpcTask)
			redisTask := hasDependencies(t, got, redisID, vpcTask)
			backendTask := hasDependencies(t, got, backendID, vpcTask, mysqlTask, redisTask)
			_ = hasDependencies(t, got, frontendID, vpcTask, backendTask)
			_ = hasDependencies(t, got, mqID)
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
			frontendTask := hasDependencies(t, got, frontendID) // 0 dependencies
			backendTask := hasDependencies(t, got, backendID, frontendTask)
			mysqlTask := hasDependencies(t, got, mysqlID, backendTask)
			redisTask := hasDependencies(t, got, redisID, backendTask)
			_ = hasDependencies(t, got, vpcID, mysqlTask, redisTask, backendTask, frontendTask)
			_ = hasDependencies(t, got, mqID)
		}
	})
}

func hasDependencies(t *testing.T, got []*Task, want resource.ID, deps ...resource.ID) resource.ID {
	for _, task := range got {
		if task.ModuleID != nil && *task.ModuleID == want {
			// Module matches, so now check dependencies
			if assert.Len(t, task.DependsOn, len(deps)) {
				for _, dep := range deps {
					assert.Contains(t, task.DependsOn, dep)
				}
				return task.ID
			}
		}
	}
	t.Fatalf("%s not found in %v", want, got)
	return resource.ID{}
}
