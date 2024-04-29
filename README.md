# PUG

A TUI application for terraform power users.

* Perform tasks in parallel (plan, apply, init, etc)
* Manage state resources
* Task queueing
* Supports tofu as well as terraform
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

### Go

```
go install github.com/leg100/pug@latest
```

### Homebrew

```
brew install leg100/tap/pug
```

To upgrade:

```
brew upgrade pug
```

### Github releases

Download and unzip a [GitHub release](https://github.com/leg100/pug/releases) for your platform.

## FAQ

### Can I use the [provider plugin cache](https://developer.hashicorp.com/terraform/cli/config/config-file#provider-plugin-cache)?

Yes. However, because the plugin cache does not permit concurrent writes, if pug detects the cache is in use it'll automatically only allow one terraform init task to run at a time.
