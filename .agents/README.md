# Agents

Project-local agent context for `yerk` (salotz RFC 22 / RFC 23).

- [context/](./context/): durable pointers and repo map (remote / portable)
- [plans/](./plans/): ephemeral multi-session plans (operator work-process)

Host-local agent context (RFC 23) for this checkout’s **workspace** lives
beside the workspace folder (parent of this replica), e.g. `../.agents/` when
this tree is `…/yerk/main` — not under this remote `.agents/`.

Root policy: [`../AGENTS.md`](../AGENTS.md).
