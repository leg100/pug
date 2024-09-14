<h1> PUG
<a title="This tool is Tool of The Week on Terminal Trove, The $HOME of all things in the terminal" href="https://terminaltrove.com/">
<img align="right" src="https://terminaltrove.com/assets/media/terminal_trove_tool_of_the_week_green_on_dark_grey_bg.png" alt="Terminal Trove Tool of The Week" height="40"></a></h1>

A terminal user interface for terraform power users.

* Perform tasks in parallel (plan, apply, init, etc)
* Interactively manage state resources (targeted plans, move, delete, etc)
* Supports terraform, [tofu](#tofu-support) and [terragrunt](#terragrunt-support)
* Supports [terragrunt dependencies](#terragrunt-support)
* Supports workspaces
* Calculate costs using [infracost](#infracost-integration)
* Automatically loads [workspace variable files](#workspace-variables)
* Backend agnostic (s3, cloud, etc)

![Demo](./demo/demo.gif)

## Install instructions

With `go`:

```
go install github.com/leg100/pug@latest
```

Homebrew:

```
brew install leg100/tap/pug
```

Or download and unzip a [GitHub release](https://github.com/leg100/pug/releases) for your system and architecture.

## Getting started

Pug requires `terraform` to be installed on your system.

The first time you run `pug`, it'll recursively search sub-directories in the current working directory for terraform root modules.

To get started with some pre-existing root modules, clone this repo, change into the `./demos/getting_started` directory, and start pug:

```bash
git clone https://github.com/leg100/pug.git
cd pug
cd demos/getting_started
pug
```

## Configuration

Pug can be configured with - in order of precedence - flags, environment variables, and a config file.

Flags:

```bash
> pug -h
NAME
  pug

FLAGS
  -p, --program STRING               The default program to use with pug. (default: terraform)
  -w, --workdir STRING               The working directory containing modules. (default: .)
  -t, --max-tasks INT                The maximum number of parallel tasks. (default: 32)
      --data-dir STRING              Directory in which to store plan files. (default: /home/louis/.pug)
  -e, --env STRING                   Environment variable to pass to terraform process. Can set more than once.
  -a, --arg STRING                   CLI arg to pass to terraform process. Can set more than once.
  -f, --first-page STRING            The first page to open on startup. (default: modules)
  -d, --debug                        Log bubbletea messages to messages.log
  -v, --version                      Print version.
  -c, --config STRING                Path to config file. (default: /home/louis/.pug.yaml)
      --disable-reload-after-apply   Disable automatic reload of state following an apply.
  -l, --log-level STRING             Logging level (valid: info,debug,error,warn). (default: info)
```

Environment variables are specified by prefixing the value with `PUG_` and appending the equivalent flag value, replacing hyphens with underscores, e.g. `--max-tasks 100` is set via `PUG_MAX_TASKS=100`.

The config file by default is expected to be found at `$HOME/.pug.yaml`. Override the default using the flag `-c` or environment variable `PUG_CONFIG`. The config uses YAML format. Set values in the config file by removing the `--` prefix from the equivalent flag value, e.g. `--max-tasks 100` is set like so in the config file:

```yaml
max-tasks: 100
```

## Workspace Variables

Pug automatically loads variables from a .tfvars file. It looks for a file named `<workspace>.tfvars` in the module directory, where `<workspace>` is the name of the workspace. For example, if the workspace is named `dev` then it'll look for `dev.tfvars`. If the file exists then it'll pass the name to `terraform plan`, e.g. for a workspace named `dev`, it'll invoke `terraform plan -vars-file=dev.tfvars`.

## Pages

### Modules

![Modules screenshot](./demo/modules.png)
 
Press `m` to go to the modules page.

*Note: what Pug calls a module is equivalent to a [root module](https://developer.hashicorp.com/terraform/language/modules#the-root-module), i.e. a directory containing terraform configuration, including a state backend. It is not to be confused with a [child module](https://developer.hashicorp.com/terraform/language/modules#child-modules).*

#### Key bindings

| Key | Description | Multi-select |
|--|--|--|
|`i`|Run `terraform init`|&check;|
|`u`|Run `terraform init -upgrade`|&check;|
|`f`|Run `terraform fmt`|&check;|
|`v`|Run `terraform validate`|&check;|
|`p`|Run `terraform plan`|&check;|
|`P`|Run `terraform plan -destroy`|&check;|
|`a`|Run `terraform apply`|&check;|
|`d`|Run `terraform apply -destroy`|&check;|
|`e`|Open module in editor|&cross;|
|`x`|Run any program|&check;|
|`Ctrl+r`|Reload all modules|-|
|`Ctrl+w`|Reload module's workspaces|&check;|

### Workspaces

![Workspaces screenshot](./demo/workspaces.png)

Press `w` to go to the workspaces page.

*Note: A workspace is directly equivalent to a [terraform workspace](https://developer.hashicorp.com/terraform/language/state/workspaces).*

#### Key bindings

| Key | Description | Multi-select |
|--|--|--|
|`i`|Run `terraform init`|&check;|
|`u`|Run `terraform init -upgrade`|&check;|
|`f`|Run `terraform fmt`|&check;|
|`v`|Run `terraform validate`|&check;|
|`p`|Run `terraform plan`|&check;|
|`P`|Run `terraform plan -destroy`|&check;|
|`a`|Run `terraform apply`|&check;|
|`d`|Run `terraform apply -destroy`|&check;|
|`C`|Run `terraform workspace select`|&cross;|
|`$`|Run `infracost breakdown`|&check;|

### State

![State screenshot](./demo/state.png)

Press `s` to go to the state page, listing a workspace's resources.

#### Key bindings

| Key | Description | Multi-select |
|--|--|--|
|`p`|Run `terraform plan -target`|&check;|
|`P`|Run `terraform plan -destroy -target`|&check;|
|`a`|Run `terraform apply -target`|&check;|
|`d`|Run `terraform apply -destroy -target`|&check;|
|`D`|Run `terraform state rm`|&check;|
|`M`|Run `terraform state mv`|&cross;|
|`Ctrl+t`|Run `terraform taint`|&check;|
|`U`|Run `terraform untaint`|&check;|
|`Ctrl+r`|Run `terraform state pull`|-|
|`S`|Toggle split screen|-|
|`+`|Increase split screen top pane|-|
|`-`|Decrease split screen top pane|-|
|`tab`|Switch split screen pane focus|-|

### Tasks

![Tasks screenshot](./demo/tasks.png)

Press `t` to go to the tasks page.

#### Key bindings

| Key | Description | Multi-select |
|--|--|--|
|`c`|Cancel task|&check;|
|`r`|Retry task|&check;|
|`Enter`|Full screen task output|&cross;|
|`S`|Toggle split screen|-|
|`+`|Increase split screen top pane|-|
|`-`|Decrease split screen top pane|-|
|`tab`|Switch split screen pane focus|-|
|`I`|Toggle task info sidebar|-|

### Task Group

![Task group screenshot](./demo/task_group.png)

Creating multiple tasks, via a selection, creates a task group, and takes you to the task group page.

#### Key bindings

| Key | Description | Multi-select |
|--|--|--|
|`c`|Cancel task|&check;|
|`r`|Retry task|&check;|
|`Enter`|Full screen task output|&cross;|
|`S`|Toggle split screen|-|
|`+`|Increase split screen top pane|-|
|`-`|Decrease split screen top pane|-|
|`tab`|Switch split screen pane focus|-|
|`I`|Toggle task info sidebar|-|

### Task Groups Listing

![Task groups screenshot](./demo/task_groups.png)

Press `T` to go to the tasks groups page, which lists all task groups.

### Logs

![Logs screenshot](./demo/logs.png)

Press `l` to go to the logs page.

## Common Key bindings

### Global

These keys are valid on any page.

| Key | Description |
|--|--|
|`?`|Open help pane|
|`Ctrl+c`|Quit|
|`Esc`|Go to previous page|
|`m`|Go to modules page|
|`w`|Go to workspaces page|
|`s`|Go to state page\*|
|`t`|Go to tasks page|
|`T`|Go to task groups page|
|`l`|Go to logs|
|`Ctrl+s`|Toggle auto-scrolling of terraform output|

\* Only where the workspace can be ascertained.

### Selections

Items can be added or removed from a selection. Once selected, actions are carried out on the selected items if the action supports multiple selection.

| Key | Description |
|--|--|
|`<space>`|Toggle selection|
|`Ctrl+a`|Select all|
|`Ctrl+\`|Clear selection|
|`Ctrl+<space>`|Select range|

### Filtering

![Filter mode screenshot](./demo/filter.png)

Items can be filtered to those containing a sub-string.

| Key | Description |
|--|--|
|`/`|Open and focus filter prompt|
|`Enter`|Unfocus filter prompt|
|`Esc`|Clear and close filter prompt|

### Navigation

Common vim key bindings are supported for navigation.

| Key | Description |
|--|--|
|`Up/k`|Up one row|
|`Down/j`|Down one row|
|`PgUp`|Up one page|
|`PgDown`|Down one page|
|`Ctrl+u`|Up half page|
|`Ctrl+d`|Down half page|
|`Home/g`|Go to top|
|`End/G`|Go to bottom|

## Reference

### Module

A module is a directory of terraform configuration with a backend configuration. When Pug starts up, it looks recursively within the working directory, walking each directory and parsing any terraform configuration it finds. If the configuration contains a [state backend definition](https://developer.hashicorp.com/terraform/language/settings/backends/configuration) then Pug loads the directory as a module.

Each module has zero or more workspaces. Following successful initialization the module has at least one workspace, named `default`. One workspace is set as the *current workspace* for the module. When you run a plan or apply on a module, it is created on its current workspace.

If you add/remove modules outside of Pug, you can instruct Pug to reload modules by pressing `Ctrl-r` on the modules listing.

### Workspace

Workspaces are parsed from the output of `terraform workspace list`, which is automatically run when:

* When a module is loaded into pug for the first time. Note the task may fail if the module is not correct initialized, and needs `terraform init` to be run.
* Following a `terraform init` task, but only if the module doesn't have a current workspace yet.

### Task

Each invocation of terraform is represented as a task.

A task is either non-blocking or blocking. Blocking tasks block their workspace or module, and prevent from further tasks from being enqueued until the blocking task has finished. For example, an `init` task, a blocking task, runs on module "A". Another `init` task for module "A", created immediately afterwards, would be blocked until the former task has completed. Or a `plan` task created afterwards on workspace "default" on module "A", would also be blocked. Blocking tasks in this manner prevent concurrent writes to resources that don't permit concurrent writes, such as the terraform state.

A task starts in the `pending` state. It enters the `queued` state only if it is unblocked (see above). It remains in the `queued` state until there is available capacity, at which point it enters the `running` state. Capacity determines the maximum number of running tasks, and defaults to twice the number of cores on your system and can be overridden using `--max-tasks`.

An exception to this rule are tasks which are classified as *immediate*. Immediate tasks enter the running state regardless of available capacity. At time of writing only the `terraform workspace select` task is classified as such.

A task can further be classed as *exclusive*. These tasks are globally mutually exclusive and cannot run concurrently. The only task classified as such is the `init` task, and only when you have enabled the [provider plugin cache](https://developer.hashicorp.com/terraform/cli/config/config-file#provider-plugin-cache) (the plugin cache does not permit concurrent writes).

A task can be canceled at any stage. If it is `running` then the current terraform process is sent a termination signal. Otherwise, in any other non-terminated state, the task is immediately set as `canceled`.

### State

When a workspace is loaded into Pug for the first time, a task is created to invoke `terraform state pull`, which retrieves workspace's state, and then the state is loaded into Pug. The task is also triggered after any task that alters the state, such as an apply or moving a resource in the state.

## Infracost integration

NOTE: Requires `infracost` to be installed on your machine, along with configured API key.

Pug integrates with [infracost](https://www.infracost.io/) to provide cost estimation. Select workspaces on the workspace page and press `$` to calculate their costs:

![Infracost output screenshot](./demo/infracost_output.png)

Once the task has finished, the costs are visible on the workspaces page:

![Worksapces with costs screenshot](./demo/workspaces_with_cost.png)

## Tofu support

To use tofu, set `--program=tofu`. Ensure it is installed first.

## Terragrunt support

To use terragrunt, set `--program=terragrunt`. Ensure it is installed first.

When `terragrunt` is specified as the program executable, Pug enables "terragrunt mode":

* Modules are detected via the presence of a `terragrunt.hcl` file. (You may want to rename the top-level `terragrunt.hcl` file to something else otherwise it is mis-detected as a module).
* Module dependencies are supported. After modules are loaded, a task invokes `terragrunt graph-dependencies`, from which dependencies are parsed and configured in Pug. If you apply multiple modules Pug ensures their dependencies are respected, applying modules in topological order. If you apply a *destroy* plan for multiple modules, modules are applied in reverse topological order.
* The flag `--terragrunt-non-interactive` is added to commands.

## Multiple terraform versions

You may want to use a specific version of terraform for each module. To do so, it's recommended to use either [asdf](https://asdf-vm.com/) or [mise](https://mise.jdx.dev/), specifying the terraform version in a `.tool-versions` file in each module. Whenever you run `terraform`, directly or via Pug, the specific version for that module is used.

However, you first need to instruct `asdf` or `mise` to install specific versions of terraform. Pug's arbitrary execution feature can be used to perform this task for multiple modules.

For example, select modules and press `x`, and when prompted type `asdf install terraform`:

![Execute asdf install terraform in each module](./demo/execute_asdf_install_terraform.png)

Press enter to run that command in each module's directory:

![Executing asdf install terraform in each module](./demo/asdf_install_terraform_task_group.png)

You've now installed a version of terraform for each version specified in `.tool-versions` files.
