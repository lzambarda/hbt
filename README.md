# hbt

A heuristic command suggestion system for zsh.
TAB anywhere and you can cycle through the commands you previously ran in the same location.

## Rationale

Storing each command into a directed graph where each node is a directory, and the edges are commands connecting nodes.

Each new shell session creates a "walker" which keeps track of the "path" of directories and commands.

The built graph can then be used to suggest possible commands to execute.

## Installation

1. Download or build `hbt`
2. Amend the environment variables in `zsh/hbt.zsh` to match where you store hbt (fancy the PATH?)
3. Make sure that `zsh/hbt.zsh` is loaded by your shell.

## Development

There is a `--debug` flag (or `HBT_DEBUG` env var) which can be used to print extra information.
By default the TCP server runs in the foreground. If you want to work on the same terminal you can use something like `nohup`.
Check out [`zsh/hbt.zsh`](./zsh/hbt.zsh) for an implementation of it.

You can change `HBT_BIN_PATH` to point to the binary in this repo.
It is usually better to run the built binary otherwise `hbt_stop` won't be able to find the running process and terminate it.

## Usage

The zsh bit of hbt talks to a locally spawned TCP server handled by a go binary.
Hbt will track every command that you type and store it into a graph.
Upon pressing TAB with an empty prompty buffer, it will try to hint at a good command, according to your typing history. Shrugs otherwise (seriously).

It internally uses some functions from [zsh-autosuggestions](https://github.com/zsh-users/zsh-autosuggestions).

Since a ~~picture~~ GIF is worth a thousand words:

![demo](./docs/demo.gif)

### Manual interaction with hbt

```bash
hbtsrv --help
```

[`zsh/hbt.zsh`](./zsh/hbt.zsh) provides some functions you can use to interact with a running hbt server.

Otherwise you can use the `cli` command to manually execute certain commands without interacting with a server (cache and graph will be the same as the server's).

Example:

```bash
hbtsrv --debug cli hint $$ $(pwd)
```

See [server/server.go](server/server.go) for the list of supported commands.

## Why the mix of go and shell functions?

At first I wanted to developed the whole thing in go, but for tracking and hinting I couldn't find an implementation faster than pure shell commands.
Given that the functions use zsh hooks which are executed at every command, I didn't want this to have a too significant impact on the terminal experience.

## Todo

- [x] Naive graph implementation
- [x] Create custom marshaller which supports cycles and correctly restores pointers

  This hasn't really been done the proper way, but it works.

- [x] Tests
- [x] Migrate what can be migrated from zsh to go
- [ ] Benchmarking
- [ ] R/B tree / ngram tree implementation???
- [x] Partial path search
- [ ] Better error catching
- [ ] More dynamic graph parameters (env variables or flags)
- [ ] Do not store sensistive information (is it even possible to detect it?)
- [ ] Identify "workflows" by using the walker model (for the naive implementation)
