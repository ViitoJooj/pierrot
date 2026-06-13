# Contributing to Pierrot

First off, thanks for taking the time to contribute! ❤️

Pierrot is a JavaScript-free web framework written in **Go**: a compiler for the
`.pierrot` language that builds static sites. There's no Node, no npm, no
`package.json` — just a Go toolchain and a single binary. This guide explains how
to get the project running locally and how to send changes.

> If you like the project but don't have time to contribute, that's fine too —
> starring the repo, reporting bugs, or telling a friend all help.

## Table of contents

- [Code of conduct](#code-of-conduct)
- [I have a question](#i-have-a-question)
- [Reporting bugs](#reporting-bugs)
- [Suggesting enhancements](#suggesting-enhancements)
- [Development setup](#development-setup)
- [Project layout](#project-layout)
- [Making a change](#making-a-change)
- [Styleguide](#styleguide)
- [Pull requests](#pull-requests)
- [License](#license)

## Code of conduct

Be respectful and constructive. Assume good faith, keep discussion on the
technical merits, and help newcomers. Report unacceptable behavior by opening an
issue or contacting the maintainer.

## I have a question

Before asking, please read the [documentation](./docs/arquitetura.md) and search
the existing [issues](https://github.com/ViitoJooj/pierrot/issues). If nothing
covers your question, open a new issue with as much context as you can.

## Reporting bugs

A good bug report is reproducible. Before filing, please:

- Make sure you're on the latest version (`go install github.com/pierrot/cmd/pierrot@latest`
  or the latest [release](https://github.com/ViitoJooj/pierrot/releases/)).
- Confirm it's a bug, not a usage question (the [docs](./docs/arquitetura.md) may
  cover it).
- Search the [issue tracker](https://github.com/ViitoJooj/pierrot/issues) for an
  existing report.

When you open an [issue](https://github.com/ViitoJooj/pierrot/issues/new),
include:

- **OS, platform and arch** (Windows, Linux, macOS; x86, ARM).
- **Go version** (`go version`) and Pierrot version.
- The smallest `.pierrot` snippet or project that reproduces the problem.
- Expected vs. actual behavior, and the full error output (the dev-server overlay
  message or the `pierrot build` log).

> Don't report security issues in public. Send them privately to the maintainer
> instead.

## Suggesting enhancements

Enhancements are tracked as [issues](https://github.com/ViitoJooj/pierrot/issues).
Use a clear title, describe the current behavior and what you'd like instead, and
explain why it fits Pierrot's scope. Pierrot deliberately stays small — features
that add a JS toolchain dependency, a virtual DOM, or `eval` are unlikely to land.

## Development setup

**Prerequisites:** [Go 1.26+](https://go.dev/dl/) (see `go.mod`). That's it — no
Node, no npm.

```bash
# 1. clone
git clone https://github.com/ViitoJooj/pierrot
cd pierrot

# 2. build and install the binary into ~/.pierrot/bin/pierrot.exe
make install

# 3. build everything (compile check)
make build

# 4. run the tests
make test
```

The `Makefile` targets:

| Target | What it does |
|--------|--------------|
| `make install` | `go build` the CLI into `~/.pierrot/bin` (add it to your `PATH`) |
| `make build` | `go build ./...` — compiles every package |
| `make test` | `go test ./...` |
| `make clean` | removes the installed binary |

You can also run the CLI directly without installing:

```bash
go run ./cmd dev      # from inside a project folder
go run ./cmd build
```

### Trying changes against the demo site

`www/` is a real Pierrot project (the official landing page). Use it to exercise
your changes end to end:

```bash
cd www
go run ../cmd dev     # http://localhost:3000, live reload
go run ../cmd build   # writes static output to www/build
```

## Project layout

```txt
cmd/                    # main.go — CLI entry point
internal/
├── cli/                # cobra commands: init, dev, build, vscode
├── workers/            # the work behind each command
│   ├── create_project.go   # `pierrot init` scaffold
│   ├── dev_server.go       # dev server, render pipeline, browser runtime
│   ├── build.go            # static-site build
│   ├── template.go         # @for/@if/@render/@bind template compiler
│   ├── props.go            # component prop expansion
│   └── vscode.go           # `pierrot vscode install`
├── readers/            # parser.go (.pierrot parsing), settings.go (config)
├── files/              # embedded scaffold templates for `pierrot init`
└── assets/             # embedded VS Code extension (.vsix)
docs/                   # documentation (start at docs/arquitetura.md)
www/                    # the official site, built with Pierrot
vscode-pierrot/         # the VS Code extension source
```

The compiler is a short pipeline of pure functions. If you're touching it, read
[`docs/arquitetura.md`](./docs/arquitetura.md#architecture) first — it maps each
stage to the file that implements it.

## Making a change

- **Fixing the framework / CLI** → Go code under `cmd/` and `internal/`.
- **Changing the `init` scaffold** → the embedded templates in `internal/files/`.
- **Documentation** → `docs/` (English, cross-linked; the hub is
  `docs/arquitetura.md`).
- **VS Code extension** → `vscode-pierrot/`.

When you change template/parser behavior, verify against the `www/` site so you
don't regress a real project.

## Styleguide

- **Go:** run `gofmt` (or `go fmt ./...`) before committing. Keep functions small
  and prefer the existing style in the package you're editing.
- **Comments:** the codebase comments the *why*, often in Portuguese — match the
  surrounding file's language and density rather than rewriting it.
- **Tests:** add or update `go test` coverage when you change behavior.

### Commit messages

- Write a short imperative subject (e.g. `fix prop expansion inside @for`).
- Add a body only when the *why* isn't obvious from the diff.
- Keep one logical change per commit.

## Pull requests

1. Fork and branch off `main`.
2. Make your change; run `make build` and `make test`.
3. If you changed the framework, confirm `www/` still builds (`pierrot build`).
4. Open a PR describing what changed and why. Link any related issue.

> By contributing, you confirm you authored the content and that it may be
> released under the project's license.

## License

By contributing, you agree that your contributions are licensed under the
[MIT License](./LICENSE.md), the same license that covers the project.
