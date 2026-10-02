# Plan: domain spine → architecture layout

## Context

`design/domain-and-near-term.md` mixes enduring domain model with historical
“near-term slice” framing. After initial-feature-series, rename/move into a
clearer architecture home and fix inbound links.

## Goals

1. Durable spine path/title without “near-term” in the name (unless Q keeps it).
2. All in-repo references updated (AGENTS, ADRs, docs, `.agents` plans).
3. No silent link rot.

## Non-goals

- Rewriting the whole domain essay
- Glossary content (separate plan; may bundle link pass)

## Phases

0. Lock target path + stub policy.  
1. Move + edit title/intro.  
2. Repo-wide link rewrite + verify.  
3. Close-out.

## Success criteria

- [ ] New path is the cited spine in AGENTS/docs.
- [ ] `rg` clean for old basename (except intentional stub/history).
- [ ] Plan removable at close.
