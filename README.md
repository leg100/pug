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

![Modules demo](https://vhs.charm.sh/vhs-2SDiU03uHVMZ1OweMU5qnw.gif)

## Workspaces

Pug supports workspaces. Invoke plan and apply on workspaces. Change the current workspace for a module.

![Workspaces demo](https://vhs.charm.sh/vhs-2VVSWika2ZVjUeBNGVXzYq.gif)

## Runs

Create multiple plans and apply them in parallel.

![Runs demo](https://vhs.charm.sh/vhs-2XzgTM8B8zMmL5kXSSP8hv.gif)

View the output of plans and applies.

![Run demo](https://vhs.charm.sh/vhs-6SbXJmeccgQG0xoENCH20A.gif)

## State management

Manage state resources. Taint, untaint and delete multiple resources. Select resources for targeted plans.

![State demo](https://vhs.charm.sh/vhs-79IoDj23zTHZnbHKekcz0o.gif)

## FAQ

### Can I use the [provider plugin cache](https://developer.hashicorp.com/terraform/cli/config/config-file#provider-plugin-cache)?

Yes. However, because the plugin cache does not permit concurrent writes, if pug detects the cache is in use it'll automatically only allow one terraform init task to run at a time.
