# PUG

A TUI application for terraform power users.

* Perform tasks in parallel (plan, apply, init, etc)
* Manage state resources
* Task queueing
* Supports tofu as well as terraform
* Supports workspaces
* Backend agnostic

![Applying runs](./demos/output/applied_runs.png)

## Modules

Invoke `init`, `validate`, and `fmt` across multiple modules.

![Modules demo](https://vhs.charm.sh/vhs-3Cy7AzQGztpAUvNmekM7f9.gif)

## Runs

Create multiple plans and apply them in parallel.

![Runs demo](https://vhs.charm.sh/vhs-141kKh9q7Ikije2rN8EBK1.gif)

View the output of plans and applies.

![Run demo](https://vhs.charm.sh/vhs-3qwIobBxxLGB6bC5OR5kYd.gif)

## State management

Manage state resources. Taint, untaint and delete multiple resources. Select resources for targeted plans.

![State demo](https://vhs.charm.sh/vhs-77uh1e8YRXR5Aux8mSU1Z3.gif)

## FAQ

### Can I use the [provider plugin cache](https://developer.hashicorp.com/terraform/cli/config/config-file#provider-plugin-cache)?

Yes. However, because the plugin cache does not permit concurrent writes, if pug detects the cache is in use it'll automatically only allow one terraform init task to run at a time.
