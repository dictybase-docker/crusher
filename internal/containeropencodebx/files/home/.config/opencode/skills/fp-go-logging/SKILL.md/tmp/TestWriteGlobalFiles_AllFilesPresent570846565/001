---
name: fp-go-logging
description: Use this skill when writing, reviewing, or refactoring Go code that adds logging to fp-go v2 pipelines. Trigger on ChainFirstIOK with IO.Logf, TapSLog, SLog, LogEntryExit, SLogWithCallback, or any structured slog integration in ReaderIOResult or IOResult pipelines.
---

# fp-go logging

Covers structured slog integration in fp-go v2 pipelines: `ChainFirstIOK` with
`IO.Logf` for tap-style logging, `TapSLog`/`SLog` for in-pipeline log entries,
and `LogEntryExit` for automatic function-boundary logging. Keeps logging side
effects isolated to the IO/IOEither layer while the pipeline stays declarative.
