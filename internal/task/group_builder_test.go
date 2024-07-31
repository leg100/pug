package task

import (
	"testing"

	"github.com/leg100/pug/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTaskCreator struct{}

func (f *fakeTaskCreator) Create(spec Spec) (*Task, error) {
	return (&factory{}).newTask(spec), nil
}

func TestNewGroupWithDependencies(t *testing.T) {
	// Create module dependency tree
	vpc := resource.New(resource.Module, resource.GlobalResource)
	mysql := resource.New(resource.Module, resource.GlobalResource).WithDependencies(vpc.ID)
	redis := resource.New(resource.Module, resource.GlobalResource).WithDependencies(vpc.ID)
	backend := resource.New(resource.Module, resource.GlobalResource).WithDependencies(vpc.ID, redis.ID, mysql.ID)
	frontend := resource.New(resource.Module, resource.GlobalResource).WithDependencies(backend.ID, vpc.ID)
	mq := resource.New(resource.Module, resource.GlobalResource)

	vpcSpec := Spec{Parent: vpc, Path: "vpc"}
	mysqlSpec := Spec{Parent: mysql, Path: "mysql"}
	redisSpec := Spec{Parent: redis, Path: "redis"}
	backendSpec := Spec{Parent: backend, Path: "backend"}
	frontendSpec := Spec{Parent: frontend, Path: "frontend"}
	mqSpec := Spec{Parent: mq, Path: "mq"}

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
