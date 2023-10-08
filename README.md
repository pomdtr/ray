# Raycast CLI

## Installation

```sh
go install github.com/pomdtr/raycast-cli@latest
```

## Usage

Use `raycast [extension] [command]` to run a command.

You can also pass arguments to the command.

```sh
raycast arc new-little-arc https://raycast.com
```

Use stdin to pass a context to the command.

```sh
jq -n '{ key: "value"}' | raycast ...
```

If you want to copy the deeplink instead of opening it, use the `--copy` flag.
You can also use the `--print` flag to print the deeplink to stdout.