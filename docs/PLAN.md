---
PLAN: "feat: sqlt honours storage.ILike (case-insensitive match on SQLite)"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 13490748600497650720
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to
> proceed or finalize — nobody answers, and a session that waits ends without your work. Commit after every stage.
> Never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash` on your own work.
> **Publishing:** you publish only by OPENING THE PULL REQUEST. Submit after EACH stage that compiles.
> **Never ask a yes/no question about progress**; ask only when a signature in this plan is wrong, and name the symbol.
> **Before you submit:** `git status` must be clean and `git diff --stat main` must list the files of the plan.

# Plan: `sqlt` answers `storage.ILike`

`storage` v0.2.3 added `storage.ILike(field, value) Condition` whose operator string is `"ILIKE"`, and its conformance clause (`conformance/conformance.go`, function checking `ILike %AN%`) must pass on every backend. SQLite has no `ILIKE`; its `LIKE` already ignores ASCII case.

## Tools
```bash
go install webtyp.com/devflow/cmd/gotest@latest
```
`go.mod` already pins `webtyp.com/storage v0.2.3`; do not change versions.

## Stage 1 — translate (`translate.go`)
In the `switch op` of the where-builder (≈ line 336, the `default:` branch builds `fmt.Sprintf("%s %s ?", c.Field(), op)`), add before `default`:

```go
case operatorILike:
	clause = fmt.Sprintf("%s LIKE ?", c.Field()) // SQLite LIKE ignores ASCII case
	args = append(args, c.Value())
```
and declare once in the same file `const operatorILike = "ILIKE"`. Non-ASCII letters are NOT case-folded by SQLite `LIKE`: say so in a comment above the case and in `docs/ARCHITECTURE.md` (one line).

## Stage 2 — test
Make sure the repo's conformance run (`grep -rn "conformance" tests/*.go | head`) executes the new `ILike` clause: it comes with the pinned storage. If the tests module pins an older `storage`, bump ONLY that pin to `v0.2.3` in `tests/go.mod`. Add `TestILike_IgnoresCase` in `tests/ilike_test.go`: insert rows `Ana`, `ana maría`, `Bob` in a `name` column, `ILike("name","%AN%")` returns the first two; row in `docs/TESTS.md`.

## Delivery check
`gotest` passes on every destination; a tree that does not compile is a failed delivery.

| Stage | Files |
|---|---|
| 1 | `translate.go`, `docs/ARCHITECTURE.md` |
| 2 | `tests/ilike_test.go`, `docs/TESTS.md` (and `tests/go.mod` only if needed) |
