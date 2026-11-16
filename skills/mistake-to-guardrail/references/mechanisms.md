# Guardrail mechanisms — cost & coverage in this repo

Measured baseline (2026-11-16): `go vet` fails at the first compile error in ~0.24 s;
`go test ./...` ≈ 4–7 s; `cargo test --workspace` ≈ 1m57 s / 1148 tests (docs/EVAL.md B11);
ast-grep and stdlib-AST tests are milliseconds. Clippy has ~30 core/server + ~70 app
existing warnings, so a `-D warnings` gate is red on day one.

| Failure class | Cheapest mechanism that fires | Second rung | Do not bother |
|---|---|---|---|
| Type/compile — Rust | move invariant into a type (newtype/enum) | `trybuild` compile-fail | `#![deny]` before the crate is clean |
| Type/compile — Go | typed capability interface instead of a flag | `go vet` analyzer | blanket `a.mode` ban |
| Custom lint — Rust | `clippy.toml` `disallowed-methods` | ast-grep (`guardrails/*.yml`) | dylint (maintenance) |
| Custom lint — Go | stdlib `go/parser` test in `guardrail_test.go` | `go/analysis` + `-vettool` | golangci-lint for one rule |
| Architecture | `go list -f '{{.Imports}}'` test; `cargo metadata` test | depguard wrapper crates | prose "don't import X" |
| Behaviour | one test, red when the guard is deleted | table test per entry point | nextest retries |
| UI structure | `ui.Tester` text/rect/a11y; `debug_bounds(id)` on Rust | — | asserting pixels for layout |
| UI pixels | hash golden (`TestVisualHarness`) | A/B seed diff | video |
| Platform/OS | decidable option assertion (`WindowOptions`, spawn flags) | real probe on the OS | per-commit Windows runner |
| Video | — | — | always (evidence only) |

Cost legend: cheap = ms–seconds, no new dependency; medium = adds a dependency or CI
job; expensive = real OS/GPU, minutes, non-deterministic.

Repo facts that shape the choice:

- Rust has **no CI** and no green clippy baseline → start with cheap source-shape
  rules + one Linux-only core/server job; do not pretend prose is a gate.
- Visual goldens: `verify_golden.json` has 16 darwin-arm64 / 13 linux-amd64 / 0 windows
  keys; `visual_test.go` `Skipf`s missing keys, so Windows reports green without being
  checked. Fix the skip (Fatal for automated screens + document opt-outs) before
  trusting it.
- `features.json` is only checked by `scripts/verify.sh`, not CI → add the step.
- Shared worktree: gates run on someone else's half-edit. Use `git worktree` +
  private `CARGO_TARGET_DIR` / `GOCACHE`.
