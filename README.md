# hbt

A heuristic command suggestion system for zsh.

## Installation

1. Download or build `hbt`
2. Amend the environment variables in `zsh/hbt.zsh` to match where you store hbt (fancy the PATH?)
3. Make sure that `zsh/hbt.zsh` is loaded by your shell.

## Usage

The zsh bit of hbt talks to a locally spawned TCP server handled by a go binary.
Hbt will track every command that you type and store it into a graph.
Upon pressing TAB with an empty prompty buffer, it will try to hint at a good command, according to your typing history. Shrugs otherwise (seriously).

## Todo

- [x] Naive graph implementation
- [x] Create custom marshaller which supports cycles and correctly restores pointers

  This hasn't really been done the proper way, but it works.

- [x] Tests
- [ ] Migrate what can be migrated from zsh to go
- [ ] Benchmarking
- [ ] R/B tree implementation???
- [x] Partial path search
- [ ] Better error catching
- [ ] More dynamic graph parameters (env variables or flags)
