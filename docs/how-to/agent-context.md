# How to dump agent context

Status: **draft**

## Goal

Give an AI agent or tool a **single structured dump** of yerk vocabulary and
“where is this directory?” without scraping help or chaining many commands.

## Prerequisites

- `yerk` on `PATH`
- For `context dir`: path under a cataloged project workspace (or cwd there)

## Steps

1. **Tool-wide context** (vocabulary, commands, XDG paths, how-to hints):

   ```sh
   yerk context
   yerk context --output json
   yerk context --output yaml
   ```

2. **Directory context** (cwd or path → project/replica + placement + short status):

   ```sh
   yerk context dir
   yerk context dir .
   yerk context dir /path/to/checkout
   yerk context dir --output json
   ```

   Prefer replica match when under a checkout. Includes effective style, bound
   flag, live replica names, and overall presence rollup. Change probes are
   off by default; pass `--git` if overall change is needed.

3. **Stability:** JSON/YAML use `apiVersion: yerk/v1` and kinds `ToolContext` /
   `DirContext`. Keys are best-effort stable (ADR 017/019) — do not rely on casual
   renames; breaks bump version or release notes.

4. **When to use finer tools instead:**

   | Need | Command |
   | --- | --- |
   | One resource by id | `yerk get <id> --output json` |
   | Path → URI only | `yerk lookup <path>` |
   | Path → full resource | `yerk lookup <path> --output json` |
   | Full placement stack | `yerk config resolve <project-id>` |
   | Full change / multi-replica | `yerk status <project-id> [--output json]` |

## See also

- [Get project info](./get-project-info.md)
- [Explain placement](./explain-placement.md)
- [Check status](./check-status.md)
- [ADR 017](../../design/decisions/017-agent-context-dumps.md)
