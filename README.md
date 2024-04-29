# PUG

A TUI application for terraform power users.

* Perform tasks in parallel (plan, apply, init, etc)
* Manage state resources
* Task queueing
* Supports workspaces
* Backend agnostic

![Applying runs](./demos/runs/applied_runs.png)

## Modules

Invoke `init`, `validate`, and `fmt` across multiple modules.

![Modules demo](https://vhs.charm.sh/vhs-224dkO2QdUANY0xFpvDbu5.gif)

## Workspaces

Pug supports workspaces. Invoke plan and apply on workspaces. Change the current workspace for a module.

![Workspaces demo](https://vhs.charm.sh/vhs-6fmPs3if1bgzNBxh2MDaan.gif)

## Runs

Create multiple plans and apply them in parallel.

![Runs demo](https://vhs.charm.sh/vhs-7sehbg4FPreF7IJ3Ljt4mx.gif)

View the output of plans and applies.

![Run demo](https://vhs.charm.sh/vhs-3wheZYKZS8bIT2516ucd9i.gif)

## State management

Manage state resources. Taint, untaint and delete multiple resources. Select resources for targeted plans.

![State demo](https://vhs.charm.sh/vhs-2f4bV5JJmPI2cAqyclFZyn.gif)

## Tasks

All invocations of terraform are represented as a task.

![Tasks demo](https://vhs.charm.sh/vhs-2MCPUcm85YRkI4QrZ3dv5b.gif)

## Install instructions

With `go`:

```
go install github.com/leg100/pug@latest
```

Homebrew:

```
brew install leg100/tap/pug
```

Or download and unzip a [GitHub release](https://github.com/leg100/pug/releases) for your platform.

## Getting started

The first time you run `pug`, it'll recursively search sub-directories for terraform root modules.

For each module it finds, it'll attempt to run `terraform workspace list`, to search for workspaces belonging to the module.

For each workspace it finds, it'll attempt to run `terraform show -json`, to retrieve the workspace's state.

## Resource hierarchy

There are several types of resources in pug:

* modules
* workspaces
* runs
* tasks

A task can be belong to a run, a workspace, or a module. A run belongs to a workspace. And a workspace belongs to a module.
 
## Scheduling

Each invocation of terraform is represented as a task. A task belongs either to a run, a workspace, or a module.

A task is either non-blocking or blocking. If it is blocking then it blocks tasks created after it that belong either to the same resource, or to a child resource. For example, an `init` task, which is a blocking task, runs on module "A". Another `init` task for module "A", created immediately afterwards, would be blocked until the former task has completed. Or a `plan` task created afterwards on workspace "default" on module "A", would also be blocked.

A task starts in the `pending` state. It enters the `queued` state only if it is unblocked (see above). It remains in the `queued` state until there is available capacity, at which point it enters the `running` state. Capacity determines the maximum number of running tasks, and defaults to twice the number of cores on your system and can be overridden using `--max-tasks`.

An exception to this rule are tasks which are classified as *immediate*. Immediate tasks enter the running state regardless of available capacity. At time of writing only the `terraform workspace select` task is classified as such.

A task can further be classed as *exclusive*. These tasks are globally mutually exclusive and cannot run concurrently. The only task classified as such is the `init` task, and only when you have enabled the [provider plugin cache](https://developer.hashicorp.com/terraform/cli/config/config-file#provider-plugin-cache) (the plugin cache does not permit concurrent writes).
