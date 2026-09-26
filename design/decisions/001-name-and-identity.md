# 001. Name and identity (`yerk`)

## Status

Accepted (2026-09-25)

## Context

The host multi-project tool needed a binary name people and agents type often.
Earlier candidates included `pj`, `prjx`, plain compounds (`workyard`), and
letter-string handles (`svpm` / `svepum`). Operator preference settled on a
carefree proper noun rather than a self-describing compound or acronym.

## Decision

- Repo, module, package id, and default binary: **yerk**
- Config directory: `~/.config/yerk` (`$XDG_CONFIG_HOME/yerk`)
- Tool env prefix: `YERK__`
- PRJX remains the spec/vocabulary layer (`.prjx-root`, `PRJX__…`)
- Optional long form from brainstorming (`yerkum`) is not required for the
  public identity

## Consequences

- README and agent prompts lead with behavior, not etymology
- Do not publish as bare `pj` / `svpm`
- Spec docs still say PRJX; the tool *implements* host ops that speak PRJX
