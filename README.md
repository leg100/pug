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

![Modules demo](https://vhs.charm.sh/vhs-3cHPiBdOWUrFPLPuT6b7Xo.gif)

## Workspaces

Pug supports workspaces. Invoke plan and apply on workspaces. Change the current workspace for a module.

![Workspaces demo](https://vhs.charm.sh/vhs-6PFGiwyBoDHAmbrqEYMJxh.gif)

## Runs

Create multiple plans and apply them in parallel.

![Runs demo](https://vhs.charm.sh/vhs-1aCXy8f3hLX90p6g1PdRb2.gif)

View the output of plans and applies.

![Run demo](https://vhs.charm.sh/vhs-5iTYJKBzZPX9fdTxlbBHus.gif)

## State management

Manage state resources. Taint, untaint and delete multiple resources. Select resources for targeted plans.

![State demo](https://vhs.charm.sh/vhs-7Lxu2duPBx5gdg82JBOXMO.gif)

## Tasks

All invocations of terraform are represented as a task.

## FAQ

### Can I use the [provider plugin cache](https://developer.hashicorp.com/terraform/cli/config/config-file#provider-plugin-cache)?

Yes. However, because the plugin cache does not permit concurrent writes, if pug detects the cache is in use it'll automatically only allow one terraform init task to run at a time.
