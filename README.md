<h1> PUG
<a title="This tool is Tool of The Week on Terminal Trove, The $HOME of all things in the terminal" href="https://terminaltrove.com/">
<img align="right" src="https://terminaltrove.com/assets/media/terminal_trove_tool_of_the_week_green_on_dark_grey_bg.png" alt="Terminal Trove Tool of The Week" height="40"></a></h1>

A TUI application for terraform power users.

* Perform tasks in parallel (plan, apply, init, etc)
* Interactively manage state resources (targeted plans, move, delete, etc)
* Supports terraform, tofu and terragrunt.
* Supports workspaces
* Automatically loads workspace variable files
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
  -f, --first-page STRING            The first page to open on startup. (default: modules)
  -d, --debug                        Log bubbletea messages to messages.log
  -v, --version                      Print version.
  -l, --log-level STRING             Logging level. (default: info)
  -c, --config STRING                Path to config file. (default: pug.yaml)
      --disable-reload-after-apply   Disable automatic reload of state following an apply.
```

Environment variables are specified by prefixing the value with `PUG_` and appending the equivalent flag value, replacing hyphens with underscores. For example, to set the max number of tasks to 100, specify `PUG_MAX_TASKS=100`.

The config file by default is expected to be found in the current working directory in which you invoke pug, and by default it's expected to be named `pug.yaml`. Override the default using the flag `-c` or environment variable `PUG_CONFIG`.

## Workspace Variables

Pug automatically loads variables from a .tfvars file. It looks for a file named `<workspace>.tfvars` in the module directory, where `<workspace>` is the name of the workspace. For example, if the workspace is named `dev` then it'll look for `dev.tfvars`. If the file exists then it'll pass the name to `terraform plan`, e.g. for a workspace named `dev`, it'll invoke `terraform plan -vars-file=dev.tfvars`.

## Resource hierarchy

There are several types of resources in pug:

* modules
* workspaces
* states
* tasks

### Modules
 
*Note: what Pug calls a module is equivalent to a [root module](https://developer.hashicorp.com/terraform/language/modules#the-root-module), i.e. a directory containing terraform configuration, including a state backend. It is not to be confused with a [child module](https://developer.hashicorp.com/terraform/language/modules#child-modules).*

A module is a directory of terraform configuration with a backend configuration. When Pug starts up, it looks recursively within the working directory, walking each directory and parsing any terraform configuration it finds. If the configuration contains a [state backend definition](https://developer.hashicorp.com/terraform/language/settings/backends/configuration) then Pug loads the directory as a module.

Each module has zero or more workspaces. Following successful initialization the module has at least one workspace, named `default`. One workspace is set as the *current workspace* for the module. When you run a plan or apply on a module, it is created on its current workspace.

If you add/remove modules outside of Pug, you can instruct Pug to reload modules by pressing `Ctrl-r` on the modules listing.

### Workspaces

A workspace is directly equivalent to a [terraform workspace](https://developer.hashicorp.com/terraform/language/state/workspaces).

When a module is loaded for the first time, Pug automatically creates a task to run `terraform workspace list`, to retrieve the list of workspaces for the module.

If you add/remove workspaces outside of Pug, you can instruct Pug to reload workspaces by pressing `Ctrl-w` on a module.

### States

Each workspace has state. Type `s` on a workspace to see its state, or type `s` on a module to see the state of its current workspace. You can also type `s` on a task, and it'll take you to the state of the task's workspace, or its module's current workspace.

When a workspace is loaded into Pug for the first time, a task is created to invoke `terraform state pull`, to retrieve the workspace's state. The task is also triggered after any task that alters the state, such as an apply or a state action.

Various actions can be carried out on state:

* delete
* taint
* untaint
* targeted plan
* targeted destroy plan
* move (only a single resource at a time)

### Tasks

Each invocation of terraform is represented as a task. A task belongs either to a workspace or a module.

A task is either non-blocking or blocking. If it is blocking then it blocks tasks created after it that belong either to the same resource, or to a child resource. For example, an `init` task, which is a blocking task, runs on module "A". Another `init` task for module "A", created immediately afterwards, would be blocked until the former task has completed. Or a `plan` task created afterwards on workspace "default" on module "A", would also be blocked.

A task starts in the `pending` state. It enters the `queued` state only if it is unblocked (see above). It remains in the `queued` state until there is available capacity, at which point it enters the `running` state. Capacity determines the maximum number of running tasks, and defaults to twice the number of cores on your system and can be overridden using `--max-tasks`.

An exception to this rule are tasks which are classified as *immediate*. Immediate tasks enter the running state regardless of available capacity. At time of writing only the `terraform workspace select` task is classified as such.

A task can further be classed as *exclusive*. These tasks are globally mutually exclusive and cannot run concurrently. The only task classified as such is the `init` task, and only when you have enabled the [provider plugin cache](https://developer.hashicorp.com/terraform/cli/config/config-file#provider-plugin-cache) (the plugin cache does not permit concurrent writes).

A task can be canceled at any stage. If it is `running` then the current terraform process is sent a termination signal. Otherwise, in any other non-terminated state, the task is immediately set as `canceled`.

#### Plans

Press `p` to create a plan. Under the hood, Pug invokes `terraform plan -out <plan-file>`. To apply the plan file, press `a` on the plan task.

Press `d` to create a destroy plan. This is identical to a plan but with a `-destroy` flag.

#### Applies

Press `a` to apply a module or workspace. Pug then requests your confirmation before invoking `terraform apply -auto-approve`.

Alternatively, you can apply a plan file (see above).

## Tofu support

To use tofu, set `--program=tofu`. Ensure it is installed first.

## Terragrunt support

To use terragrunt, set `--program=terragrunt`. Ensure it is installed first.

When `terragrunt` is specified as the program executable, Pug enables "terragrunt mode":

* Modules are detected via the presence of a `terragrunt.hcl` file. (You may want to rename the top-level `terragrunt.hcl` file to something else otherwise it is mis-detected as a module).
* The flag `--terragrunt-non-interactive` is added to commands.
