package module

import (
	"bytes"
	"testing"

	"github.com/leg100/pug/internal"
	"github.com/leg100/pug/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadTerragruntDependenciesFromDigraph(t *testing.T) {
	// Setup modules and load into table
	workdir, err := internal.NewWorkdir("/home/louis/co/pug/internal/module/testdata/terragrunt/")
	require.NoError(t, err)
	vpc := New(workdir, Options{Path: "root/vpc"})
	redis := New(workdir, Options{Path: "root/redis"})
	mysql := New(workdir, Options{Path: "root/mysql"})
	frontend := New(workdir, Options{Path: "root/frontend-app"})
	backend := New(workdir, Options{Path: "root/backend-app"})
	svc := &Service{
		table:   &fakeModuleTable{modules: []*Module{vpc, redis, mysql, backend, frontend}},
		workdir: workdir,
	}

	// Check it can handle both relative and absolute paths - `terragrunt
	// graph-dependencies` outputs the latter if there are "external"
	// dependencies.
	var (
		relativePaths = `
digraph {
        "root/backend-app" ;
        "root/backend-app" -> "root/mysql";
        "root/backend-app" -> "root/redis";
        "root/backend-app" -> "root/vpc";
        "root/frontend-app" ;
        "root/frontend-app" -> "root/backend-app";
        "root/frontend-app" -> "root/vpc";
        "root/mysql" ;
        "root/mysql" -> "root/vpc";
        "root/redis" ;
        "root/redis" -> "root/vpc";
        "root/vpc" ;
}
`
		absolutePaths = `
digraph {
        "/home/louis/co/pug/internal/module/testdata/terragrunt/root/backend-app" ;
        "/home/louis/co/pug/internal/module/testdata/terragrunt/root/backend-app" -> "/home/louis/co/pug/internal/module/testdata/terragrunt/root/mysql";
        "/home/louis/co/pug/internal/module/testdata/terragrunt/root/backend-app" -> "/home/louis/co/pug/internal/module/testdata/terragrunt/root/redis";
        "/home/louis/co/pug/internal/module/testdata/terragrunt/root/backend-app" -> "/home/louis/co/pug/internal/module/testdata/terragrunt/root/vpc";
        "/home/louis/co/pug/internal/module/testdata/terragrunt/root/mysql" ;
        "/home/louis/co/pug/internal/module/testdata/terragrunt/root/mysql" -> "/home/louis/co/pug/internal/module/testdata/terragrunt/root/vpc";
        "/home/louis/co/pug/internal/module/testdata/terragrunt/root/redis" ;
        "/home/louis/co/pug/internal/module/testdata/terragrunt/root/redis" -> "/home/louis/co/pug/internal/module/testdata/terragrunt/root/vpc";
        "/home/louis/co/pug/internal/module/testdata/terragrunt/root/vpc" ;
}
`
	)

	for _, output := range []string{relativePaths, absolutePaths} {
		err := svc.loadTerragruntDependenciesFromDigraph(bytes.NewBufferString(output))
		require.NoError(t, err)

		// vpc
		assert.Len(t, vpc.Dependencies(), 0)
		// redis
		if assert.Len(t, redis.Dependencies(), 1) {
			assert.Equal(t, vpc.ID, redis.Dependencies()[0].GetID())
		}
		// mysql
		if assert.Len(t, mysql.Dependencies(), 1) {
			assert.Equal(t, vpc.ID, mysql.Dependencies()[0].GetID())
		}
		// backend
		if assert.Len(t, backend.Dependencies(), 3) {
			assert.Contains(t, backend.Dependencies(), vpc.ID)
			assert.Contains(t, backend.Dependencies(), redis.ID)
			assert.Contains(t, backend.Dependencies(), mysql.ID)
		}
		// frontend
		if assert.Len(t, frontend.Dependencies(), 2) {
			assert.Contains(t, frontend.Dependencies(), vpc.ID)
			assert.Contains(t, frontend.Dependencies(), backend.ID)
		}
	}
}

type fakeModuleTable struct {
	modules []*Module

	moduleTable
}

func (f *fakeModuleTable) List() []*Module {
	return f.modules
}

func (f *fakeModuleTable) Update(id resource.ID, updater func(*Module) error) (*Module, error) {
	for _, mod := range f.modules {
		if mod.ID == id {
			if err := updater(mod); err != nil {
				return nil, err
			}
			return mod, nil
		}
	}
	return nil, resource.ErrNotFound
}
