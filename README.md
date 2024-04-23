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

Invoke `init`, `validate`, and `fmt` across multiple modules.

![Modules demo](https://vhs.charm.sh/vhs-yP3fHxUgxvKfimnTk9GLk.gif)

## Runs

Create multiple plans and apply them in parallel.

![Runs demo](https://vhs.charm.sh/vhs-5cktxirUmHmO7R37AYRYib.gif)

View the output of current and past runs.

![Run demo](https://vhs.charm.sh/vhs-6rzYrbsZN98dPUUMyVkWPN.gif)

## State management

Manage state resources. Taint, untaint and delete multiple resources. Select resources for targeted plans.

![State demo](https://vhs.charm.sh/vhs-oPUY2OAFmb4nvtVl1xY88.gif)

## FAQ

### Can I use the [provider plugin cache](https://developer.hashicorp.com/terraform/cli/config/config-file#provider-plugin-cache)?

Yes. However, because the plugin cache does not permit concurrent writes, if pug detects the cache is in use it'll automatically only allow one terraform init task to run at a time.
