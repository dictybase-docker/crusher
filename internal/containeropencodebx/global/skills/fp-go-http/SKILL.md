---
name: fp-go-http
description: Use this skill when writing, reviewing, or refactoring Go code that makes HTTP requests using fp-go v2. Trigger on fp-go HTTP client usage, ReaderIOResult pipelines with HTTP calls, parallel HTTP requests with TraverseArray or TraverseTuple2, or the http/builder API for custom headers, query params, or JSON bodies.
---

# fp-go http

Covers HTTP request construction and execution in fp-go v2: the `http/builder`
API for custom headers/query params/JSON bodies, `ReaderIOResult` pipelines that
thread configuration into HTTP calls, and parallel execution via
`TraverseArray`/`TraverseTuple2`. Pair with `fp-go-logging` for request/response
tracing.
