# Objective

Build a standalone planning brief that can be handed to any agentic AI system to implement a Go CLI application in this repository. The brief must capture all user feedback given so far and keep the implementation constrained to a `container build` wrapper built with `github.com/urfave/cli/v3` and `github.com/IBM/fp-go/v2`.

The objective includes all of the following:

- current narrowed CLI surface:
  - Dockerfile path, default `Dockerfile`
  - repeatable tag list, default `latest`
  - fixed build context `.` with no user-facing flag
- source paths explicitly requested for fp-go code and semantics review:
  - `/Users/sba964/Projects/devenv/golang/learn-golang/grpc/plasmid/goldenbraid`
  - `/Users/sba964/Projects/devenv/golang/learn-golang/fp-go-concepts/v2`
  - `/Users/sba964/Projects/devenv/golang/learn-golang/llm/adk/devops-engineer`
- mandatory fp-go-concepts/v2 subpaths to read before planning or coding:
  - `/Users/sba964/Projects/devenv/golang/learn-golang/fp-go-concepts/v2/either/`
  - `/Users/sba964/Projects/devenv/golang/learn-golang/fp-go-concepts/v2/ioeither/`
  - `/Users/sba964/Projects/devenv/golang/learn-golang/fp-go-concepts/v2/error-handling-examples/`
  - `/Users/sba964/Projects/devenv/golang/learn-golang/fp-go-concepts/v2/file/`
- strict functional-style instructions that must govern the implementation:
  - no imperative branching in application code samples
  - validation, defaults, optional behavior, and terminal control expressed through `Either`, `Option`, and `IOEither` combinators
  - effects isolated inside `IOEither`
  - direct argv construction for `exec.CommandContext`, never shell-string construction
- important findings discovered while developing the plan:
  - `command-reference.md` is the local source of truth for `container build`
  - `Dockerfile` already exists at the repository root
  - the scanned Either examples rely on combinators such as `E.FromPredicate`, `E.SequenceArray`, `E.FromOption`, `E.GetOrElse`, and `E.Fold`
  - planning guidance is grounded in inspected local code and local command documentation
  - the current repository still has no Go source files yet

The concrete command in scope remains a build command that drives `container build`.
