---
name: mistake-to-guardrail
description: >-
  Turn a fixed bug or a repeated agent mistake into a permanent mechanical check — a
  type change, clippy/analyzer rule, ast-grep/AST test, architecture test, behavior
  test, headless UI assertion, pixel golden, or platform probe — and prove the check
  fails on the old code. Use right after fixing a bug, when a review finding repeats,
  when a regression ticket is filed, or when the user says "again", "never again",
  "每次都这样", "别再犯", "固化", "加个 lint / 测试让它不会再犯". This skill covers what
  happens *after* a mistake; pair it with rust-best-practices / go-best-practices for
  writing the code.
---

# Mistake → permanent guardrail

Prose rules ("请记得…") are temporary. After a bug is fixed, encode the failure as a
check that **fails on the old code**, wire it into a command people already run, and
replace the prose with an ID + command. The repo should get cleaner, not longer.

## 1. When this triggers

- A bug fix is committed or staged (before closing the ticket);
- the same review finding appears twice, or a reviewer writes "again / same as last time";
- a regression ticket has a concrete repro;
- the user says any of: *again, recurring, always, never again, 每次都这样, 别再犯,
  固化, 机制化, 加个检查*.

**Not** triggered by: a one-off typo, an unreproducible environment outage, a style
opinion with no incident, or a visual judgement ("这段留白 2px 更好看").

## 2. Workflow (8 steps, each with an exit criterion)

1. **Green baseline on your worktree.** `git worktree add` + private `CARGO_TARGET_DIR`
   when other agents share the tree. You encode a **fixed** bug, never a live one.
   *Exit:* the relevant test command is green before you add anything.
2. **Write the symptom in the user's words** — one sentence, no mechanism words
   (e.g. "Windows 上执行命令时弹出命令行窗口"). This sentence becomes the failure text.
3. **Pick the lowest rung that would have gone red** (see §3) and name why nothing
   cheaper works. *Exit:* mechanism + rationale.
4. **Write it red first.** On a scratch branch, revert the fix (or delete the one guard),
   run the new check, and paste the failing output. If it stays green, the check is
   vacuous — go back to step 3.
5. **Prove both directions.** Inject a synthetic violation → red; flip an allowed case →
   green. Delete the injection. Paste both outputs into the ticket.
6. **Wire it to an existing command.** Go: a step in
   `clients/mygo-native/.github/workflows/ci.yml` or `scripts/verify.sh`. Rust: a
   `bun run gate:rust`-style entry (there is no root gate today — create one only when
   the first rule lands). Write the runtime budget next to it.
7. **Register it** in `docs/agents/guardrails.md` with an ID (`G01`, `G02`, …), class,
   mechanism, command, allowlist (with reasons + review date), cost, incident link.
8. **Replace prose, don't append.** In `AGENTS.md` / `clients/mygo-native/AGENTS.md`,
   collapse the sentence the rule covers to `Gxx` + the command. Net line count of the
   two files must not grow. Allowlist lines carry `# why: … # review: YYYY-MM`.

## 3. Rung ladder — pick the highest rung that still runs per commit

1. **Make it unrepresentable** — type/enum/constructor/capability interface. Zero
   runtime, zero false positives. (Best: the compiler *cannot* express the bug.)
2. **Compiler/lint config** — `clippy.toml` `disallowed-methods`, `go vet` analyzer.
   Cheap, but requires a clean baseline; do not turn on a red gate.
3. **Source-shape rule** — ast-grep (Rust) / stdlib-AST test (Go). Milliseconds, no
   build, reviewable in the diff.
4. **Architecture test** — `go list -deps`, `cargo metadata`, manifest assertions.
5. **Behaviour test** — one test that goes red when the guard is deleted; failure message
   carries the scene (`tt.Texts()`), not just "expected true".
6. **Headless UI assertion** — text/rect/accessibility, deterministic time via
   `Advance`; never sleeps.
7. **Pixel golden** — only when "the paint itself is the contract". Add a `.diff.png`
   artefact on mismatch so blessing is reviewable, not a hash alone.
8. **Platform probe** — real Windows/macOS/WSLg. Release checklist, never per commit.

## 4. Where the rule lives

| Class | Location | Mechanism |
|---|---|---|
| Go single-package shape | `clients/mygo-native/guardrail_test.go` | `go/parser` AST walk |
| Go cross-package | `clients/mygo-native/tools/guardrails/` + `go vet -vettool` | `go/analysis` |
| Go layering | same `guardrail_test.go` | `go list -f '{{.Imports}}'` |
| Go behaviour | owning `*_test.go` | table test / scripted events |
| Go UI structure | `ui.NewTester` assertions | text/rect/a11y |
| Go pixels | `visual_test.go` + `verify_golden.json` | hash manifest + diff image |
| Rust source shape | `guardrails/*.yml` (root) + `sg scan` | ast-grep |
| Rust clippy policy | `clippy.toml` | `disallowed-methods` (narrow) |
| Rust architecture | `crates/<crate>/tests/architecture.rs` | `cargo metadata` / TOML |
| Rust geometry | owning module `#[cfg(test)]` | `TestAppContext` + `debug_bounds` |
| Platform facts | `examples/*_probe`, `docs/VERIFY-*` | evidence, not a gate |

## 5. "Video" is evidence, not a gate

Do not make video a guardrail. It cannot be asserted, is flaky under load, and failure
artefacts are megabytes nobody reviews.

- For animation, assert **frames at named times** with the fake clock
  (`ui.Tester.Advance` on Go, the GPUI test scheduler on Rust).
- For native chrome / WebView / GTK overlays that headless rendering cannot see, attach
  a short capture to the **ticket** as evidence, and pin a decidable proxy in the gate
  (e.g. `HideWindowButtons == true` on Linux, `false` elsewhere).
- Screenshots are the review artefact; pixels are the gate only via a deterministic
  golden with a diff image.

## 6. Stop list — do not encode

- Human judgement: taste, naming, when to split a file.
- Pixel-aesthetic calibration ("3px more"): assert structure, record pixels as evidence.
- Native artefacts the runner cannot see (WSLg GTK focus ring, Windows DLP-encrypted
  screenshots): encode the proxy; keep the visual claim in the ticket.
- Rules needing a suppress hatch that becomes the normal path (`retries`, `--force`,
  auto-blessing, `#[allow]` loops).
- Repo-wide zero-warning baselines while the baseline is red (~100 clippy warnings today).
- One-off release checks (installer, signing, PE subsystem) — release checklist.
- Speculative rules for bugs nobody has observed.
- Blanket bans with legitimate uses (`time.Sleep` deadlines, `unwrap` in examples).

## 7. Anti-patterns seen in practice

- **Green on day one, red on day two**: gate added to a red tree ⇒ everyone ignores it.
- **Test that passes without the feature** (non-nil / grew / no panic): not a guardrail.
- **Allowlist by line number**: rots on every edit; allowlist by function/symbol + reason.
- **Rule + prose**: both kept ⇒ two sources of truth; replace the prose with the ID.
- **Nine rules from one incident**: encode the class, not the instance.
- **Silent skip**: a missing platform key that `Skipf`s hides Windows entirely.

## 8. References

- `references/mechanisms.md` — mechanism matrix with costs for this repo.
- `references/proof.md` — the mutation/red-green protocol, allowlist hygiene.
- `templates/` — AST-rule test, registry row, architecture test.
- Registry: `docs/agents/guardrails.md` (G01+).
- Hand-off: after step 8, close the ticket with the G-ID and the red/green outputs.
