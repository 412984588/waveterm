# Repository Guidelines

## Project Structure & Module Organization

- `frontend/` — React UI and client-side state.
- `emain/` — Electron main process and preload scripts.
- `cmd/` — Go entrypoints (e.g., wavesrv, wsh, tools).
- `pkg/` — shared Go packages (backend, RPC, utilities).
- `tsunami/` — auxiliary services and frontend for Tsunami.
- `docs/` — docsite source (with its own `package.json`).
- `tests/` and `testdriver/` — test fixtures and drivers.
- `assets/`, `public/`, `static/` — bundled static assets.
- `schema/`, `dist/`, `build/` — generated artifacts.

## Build, Test, and Development Commands

- `task init` — install Go/Node dependencies for the repo.
- `task dev` — run Electron via Vite dev server (HMR).
- `task start` — run Electron standalone (no hot reload).
- `task package` — production build + packaging (artifacts in `make/`).
- `npm test` — run Vitest.
- `npm run coverage` — Vitest with coverage.
- `task docsite` — run docs dev server (from `docs/`).

## Coding Style & Naming Conventions

- Go: run `gofmt` on touched files; keep packages in `pkg/` with clear, lower_snake filenames.
- TS/React: ESLint + Prettier (120-char print width, spaces not tabs). Follow existing patterns in `frontend/app`.
- Keep changes focused; avoid drive‑by refactors.

## Testing Guidelines

- Frontend/unit tests use Vitest (`npm test`).
- Go tests live alongside code as `*_test.go` (run `go test ./...` when touching backend logic).
- Prefer deterministic tests; avoid external network calls unless required.
- Coverage: no explicit threshold; use `npm run coverage` when requested or when adding test-heavy changes.

## Commit & Pull Request Guidelines

- Commit messages are short and imperative; optional scopes like `[wave-orch]` appear in history. Merged PRs often add `(#1234)`.
- One PR = one logical change. Discuss larger changes in Discord first (see `CONTRIBUTING.md`).
- Include a clear description, reproduction steps, and screenshots for UI changes. Sign the CLA on your first PR.

## Agent-Specific Notes (Wave‑Orch)

- If working on Wave‑Orch, follow `CLAUDE.md` and `prd.txt`, and keep `docs/wave-orch/` plan/status files updated.
