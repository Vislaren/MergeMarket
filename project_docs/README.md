# MergeMarket — Project Docs Index

> This folder is the single reference hub for both agents.
> Read the relevant files here before implementing any feature.
> These files never change unless a major architectural decision is made.

---

## What's in This Folder

| File / Folder | What It Contains | Who Reads It |
|---|---|---|
| `architecture/ARCHITECTURE.md` | Full system architecture, service breakdown, infrastructure decisions | Both agents, every session |
| `api/API_CONTRACTS.md` | All request/response schemas between services | Both agents — Agent B codes mocks from this |
| `database/DATABASE_SCHEMA.md` | Full PostgreSQL, TimescaleDB, and Redis schema | Agent A when building services |
| `flows/USER_FLOWS.md` | Step-by-step user journeys through the app | Agent B when building screens |
| `ui/COMPONENT_LIBRARY.md` | Every reusable UI component, its props, states, and usage | Agent B every session |
| `ui/UI_MIGRATION_PROMPT.md` | How to translate Stitch (Firebase) designs into Flutter | Agent B every session |
| `ui/samples/` | UI design images exported from Stitch | Agent B references when building screens |

---

## Reading Order Per Agent

### Agent A
1. `architecture/ARCHITECTURE.md` — understand the full system
2. `api/API_CONTRACTS.md` — know the contracts you must implement
3. `database/DATABASE_SCHEMA.md` — know the schema your services write to

### Agent B
1. `architecture/ARCHITECTURE.md` — understand the full system
2. `api/API_CONTRACTS.md` — build mocks from these contracts
3. `flows/USER_FLOWS.md` — understand what each screen must do
4. `ui/UI_MIGRATION_PROMPT.md` — how to convert Stitch designs to Flutter
5. `ui/COMPONENT_LIBRARY.md` — which components to use for each screen
6. `ui/samples/` — reference the design images while building
