# Domain Documentation

This repository uses a **multi-context** layout for domain documentation.

## Layout

```
/
├── CONTEXT-MAP.md          # Index pointing to per-context docs
├── packages/
│   ├── core/
│   │   ├── CONTEXT.md      # Core package domain context
│   │   └── docs/adr/       # Core package ADRs
│   ├── frontend/
│   │   ├── CONTEXT.md      # Frontend domain context
│   │   └── docs/adr/       # Frontend ADRs
│   └── backend/
│       ├── CONTEXT.md      # Backend domain context
│       └── docs/adr/       # Backend ADRs
└── docs/adr/               # Project-wide ADRs
```

## Consumer Rules

Skills that read domain docs (`improve-codebase-architecture`, `diagnose`, `tdd`) will:

1. Read `CONTEXT-MAP.md` first to understand the project structure
2. Navigate to the relevant sub-project's `CONTEXT.md` based on the files being worked on
3. Read ADRs from the sub-project's `docs/adr/` directory
4. Fall back to root-level `docs/adr/` for project-wide decisions

## Creating Context Files

Each sub-project should have:
- `CONTEXT.md` — domain language, business rules, key concepts
- `docs/adr/` — architectural decision records specific to that sub-project
