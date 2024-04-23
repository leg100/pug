# PUG

A TUI application for terraform power users.

* Perform tasks in parallel (plan, apply, init, etc)
* Manage state resources
* Task queueing.
* Supports tofu as well as terraform
* Supports workspaces
* Backend agnostic

![Applying runs](./demos/output/applied_runs.png)

## Modules

Invoke `init`, `validate`, and `fmt` across multiple modules.

![Modules demo](https://vhs.charm.sh/vhs-1rsDMnWznm105jZPZD3oW5.gif)

## Runs

Create multiple plans and apply them in parallel.

![Runs demo](https://vhs.charm.sh/vhs-61FyNZHAGIN5VnlCOefWl7.gif)

View the output of plans and applies.

![Run demo](https://vhs.charm.sh/vhs-madv068t0GBZOIq7uybMR.gif)

## State management

Manage state resources. Taint, untaint and delete multiple resources. Select resources for targeted plans.

![State demo](https://vhs.charm.sh/vhs-181dbgBQnI6XBy5oIZhWqr.gif)

## FAQ

### Can I use the [provider plugin cache](https://developer.hashicorp.com/terraform/cli/config/config-file#provider-plugin-cache)?

Yes. However, because the plugin cache does not permit concurrent writes, if pug detects the cache is in use it'll automatically only allow one terraform init task to run at a time.
