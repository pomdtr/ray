# Raycast CLI

## Installation

```sh
# using homebrew
brew install pomdtr/tap/ray

# compile from source
go install github.com/pomdtr/ray@latest
```

or download the binary from the [releases page](https://github.com/pomdtr/ray/releases).

See the `raycast completion` command to generate completion scripts for your shell.

## Usage

Use `ray [extension] [command]` to run a command.

You can also pass arguments to the command.

```sh
ray arc new-little-arc https://raycast.com
```

Use stdin to pass a context to the command.

```sh
jq -n '{ key: "value"}' | ray ...
```

If you want to copy the deeplink instead of opening it, use the `--copy` flag.
You can also use the `--print` flag to print the deeplink to stdout.
