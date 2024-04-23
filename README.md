# PUG

A TUI application for terraform power users.

* Perform tasks in parallel (plan, apply, init, state ops etc)
* Built-in queuing of tasks.
* Manage state resources
* Supports tofu as well as terraform
* Supports workspaces
* Backend agnostic

![Applying runs](./demos/applying_runs.png)

## Modules

*Note: a pug "module" is more accurately a [root module](https://developer.hashicorp.com/terraform/language/modules#the-root-module)*.

![Modules demo](https://vhs.charm.sh/vhs-25Rrh8wNPvkuQ3gdDUlELu.gif)

## Runs

![Runs demo](https://vhs.charm.sh/vhs-3NlDZpoxnp6o31lzWAy1PJ.gif)

![Run demo](https://vhs.charm.sh/vhs-5MIcWiERfQPPYyUUksgNRb.gif)

## State management

![State demo](https://vhs.charm.sh/vhs-X34EQIErI8F2gRvA2cD63.gif)

## FAQ

### Can I use the [provider plugin cache](https://developer.hashicorp.com/terraform/cli/config/config-file#provider-plugin-cache)?

Yes. However, because the plugin cache does not permit concurrent writes, if pug detects the cache is in use it'll automatically only allow one terraform init task to run at a time.
