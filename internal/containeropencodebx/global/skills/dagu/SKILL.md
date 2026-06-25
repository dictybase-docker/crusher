---
name: dagu
description: Writes, validates, and debugs Dagu DAG workflow definitions in YAML. Covers all executor types, DAG YAML schema, CLI commands, environment variables, and critical pitfalls. Use when creating, editing, or troubleshooting Dagu .yaml DAG files. Do not use for general YAML editing.
---

# dagu

Covers Dagu DAG workflow definitions: the YAML schema for all executor types
(command, docker, http, mail, ssh), the `schedule`, `params`, `env`, and
`depends` fields, the `dagu start`/`dagu status`/`dagu retry` CLI commands, and
critical pitfalls such as YAML timestamp parsing of cron expressions and the
difference between `command` and `executor`.
