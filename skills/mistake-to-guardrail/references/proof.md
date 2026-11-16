# Proof protocol — the check must fail on the old code

A guardrail is only real if you have seen it red for the **original** bug and green
after the fix. Do this on a scratch branch; never leave the revert in the final diff.

## Red

```bash
git switch -c guardrail-proof/<slug>
# Option A: revert the fix commit, keep the new check
git revert --no-commit <fix-sha>
# Option B: hand-delete only the guard line (e.g. hideConsole(cmd))
go test -run 'TestSpawnSitesHideConsole' ./...   # MUST fail, with the symptom sentence
git checkout -- <files>                          # restore
```

If it stays green, the check is vacuous: it asserts a proxy that the fix did not change.
Delete it and go back to the rung ladder.

## Green + non-vacuity

1. Restore the fix → same command green.
2. Inject a synthetic violation in a scratch file (new `exec.Command` without
   `hideConsole`, a raw hex in a view, an undeclared async site) → red with `file:line`.
3. Confirm an **allowed** case stays green (the allowlisted Linux opener, the palette
   function). Both directions prove the allowlist is not hiding everything.

## Allowlist hygiene

```go
// why: xdg-open/gio/kde-open are Linux desktop launchers, no console window exists.
// review: 2027-05
"openURLInBrowser": true,
```

- Allowlist by **function/symbol**, never by line number (edits rot line numbers).
- Every entry needs `why` + `review`; an entry without a reason fails review.
- If the allowlist grows with each incident, the rule class is wrong — encode the class
  or drop the rule.

## Failure messages

The message must contain the **symptom sentence** from step 2 of the skill, plus
`file:line`. A reviewer should understand the incident without opening the ticket.

## Wiring

- Go: a step in `clients/mygo-native/.github/workflows/ci.yml`; the rule itself runs in
  `go test ./...` (no extra command to remember).
- Rust: one `bun run gate:rust` entry (`sg scan` + targeted clippy/test). Do not add
  `-D warnings` until the crate is at zero.
- Print the runtime budget next to the gate so a slow rule is a deliberate trade.
