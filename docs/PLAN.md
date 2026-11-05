---
PLAN: "feat(sqlt): render ddl.OpCopyColumn; update to ddl v0.0.45 and storage v0.1.18 (NULL scans as zero in the test conn)"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 7716309649223077380
PR: https://github.com/webtyp/sqlt/pull/6
REVIEWED_COMMIT: 59b16ec598936b4565b2581c3d25432d2c092f16
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to
> proceed or finalize — nobody answers, and a session that waits ends without your work.

# webtyp/sqlt — copy one column into another

## Why

`webtyp.com/ddl` v0.0.45 added `ddl.OpCopyColumn`: before a column a model no longer has is dropped, its values are
copied into the column that replaced it, only where that column is still empty. `ddl/conformance` now has two new
clauses every SQL backend must pass: `copy_column_fills_only_empty_rows` and `drop_column_drops_a_column_with_data`.
Today `sqlt` answers `unknown DDL op` for the first. Updating `ddl` also pulls `webtyp.com/storage` from v0.0.7 to
v0.1.18, whose conformance requires NULL to scan as the Go zero value: the test connection in `tests/` must apply it.

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
export GOPRIVATE=webtyp.com/*,github.com/veltylabs/* GOFLAGS=-mod=mod
# in BOTH modules (repo root and tests/):
go get webtyp.com/ddl@v0.0.45 && go mod tidy
(cd tests && go get webtyp.com/ddl@v0.0.45 && go mod tidy)
```
This upgrades `webtyp.com/storage` to v0.1.18 in both. Measured before this plan: after the upgrade, `go test ./...` in
`tests/` fails exactly on `TestSqlt_DBConformance/null_scans_as_zero` («converting NULL to string is unsupported») and
`TestSqlt_DDLConformance/copy_column_fills_only_empty_rows` («unknown DDL op»). Everything else passes.

## APIs used

| Symbol | Exact signature | Tag | File |
|---|---|---|---|
| `ddl.OpCopyColumn` | `ddl.Op` constant, after `OpDropColumn` | `webtyp.com/ddl v0.0.45` | `compiler.go` |
| `ddl.Stmt` | `struct{ Op Op; Table string; Database string; Column *model.Field; OldName string; ColumnName string }` — for `OpCopyColumn`: `OldName` = source column, `ColumnName` = destination column | v0.0.45 | `compiler.go` |
| `storage.NullSafe` | `func NullSafe(dest []any) []any` — wraps scan destinations so NULL yields the zero value | `webtyp.com/storage v0.1.18` | `scan.go` |
| `storage.ErrNoRows` | `error` | v0.1.18 | |

Reference implementation of the NULL rule (another repo, restated): `webtyp/postgres` `adapter.go`
```go
type nullRows struct{ *sql.Rows }

func (r nullRows) Scan(dest ...any) error { return r.Rows.Scan(storage.NullSafe(dest)...) }
```

If a signature in the published tag differs from this table, stop and name the symbol in the PR; never adapt it by guessing.

## Stage 1 — Red test

In `tests/sqlt_test.go`, after `TestTranslateDropColumn`, add:

```go
func TestTranslateCopyColumn(t *testing.T) {
	sql, _, err := sqlt.TranslateDDL(ddl.Stmt{
		Op:         ddl.OpCopyColumn,
		Table:      "holiday",
		OldName:    "label",
		ColumnName: "name",
	}, testModel{})
	if err != nil {
		t.Fatal(err)
	}
	want := "UPDATE holiday SET name = label WHERE name IS NULL AND label IS NOT NULL"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestTranslateCopyColumn_MissingNameFails(t *testing.T) {
	_, _, err := sqlt.TranslateDDL(ddl.Stmt{Op: ddl.OpCopyColumn, Table: "holiday", OldName: "label"}, testModel{})
	if err == nil {
		t.Fatal("expected an error when the destination column is empty")
	}
}
```
Commit.

## Stage 2 — Code

`translate.go`, in `translateDDL`, add a case before `default`:
```go
	case ddl.OpCopyColumn:
		return buildCopyColumn(s)
```
and, after `buildDropColumn`:
```go
func buildCopyColumn(s ddl.Stmt) (string, []any, error) {
	if s.Table == "" || s.OldName == "" || s.ColumnName == "" {
		return "", nil, fmt.Err("table, old name and column", "required for copy column")
	}
	return fmt.Sprintf("UPDATE %s SET %s = %s WHERE %s IS NULL AND %s IS NOT NULL",
		s.Table, s.ColumnName, s.OldName, s.ColumnName, s.OldName), nil, nil
}
```
So the phrase is shared (reuse rule), change `buildRenameColumn`'s error to
`fmt.Err("table, old name and column", "required for rename")`. Grep: `grep -n '"table, old name and column required for rename"' translate.go` → empty.

`tests/conformance_test.go` — the test connection follows the NULL rule on both read paths:
```go
func (c *sqlConn) Query(q string, a ...any) (storage.Rows, error) {
	rows, err := c.DB.Query(q, a...)
	if err != nil {
		return nil, err
	}
	return nullRows{rows}, nil
}

// nullRows applies the storage contract's NULL rule to the read-all path.
type nullRows struct{ *sql.Rows }

func (r nullRows) Scan(dest ...any) error { return r.Rows.Scan(storage.NullSafe(dest)...) }
```
and in `noRows.Scan` replace `r.s.Scan(d...)` with `r.s.Scan(storage.NullSafe(d)...)`.

Anti-footgun: do not quote identifiers (house style, AGENTS.md); do not touch `ddl` or `storage` behaviour here.

`gotest` in the root and `go test ./...` in `tests/` → green, including both new `ddl` conformance clauses. Commit.

## Stage 3 — Translations

`langc sync .` creates or updates `lang.json`; fill every empty value it reports (`es`, `zh`, `hi`, `ar`, `pt`, `fr`,
`de`, `ru`), not only the new keys. Zero untranslated keys and zero reuse warnings; a reuse warning is fixed by
composing the phrase at the call site as in Stage 2. Commit.

## Stage 4 — Documentation

`README.md` to the template (≤ 150 lines), four sections in order:
1. `## Why This Exists` — two sentences (SQLite SQL from `storage.Query`/`ddl.Stmt`, no connection) + a Without/With
   table: without, each app writes SQLite SQL by hand per change; with, `ddl.Sync`/`CopyColumn`/`DropColumn` and
   `storage` queries render through `sqlt`. Inline link to [webtyp/sqlite](https://github.com/webtyp/sqlite) (the
   driver) and [webtyp/ddl](https://github.com/webtyp/ddl).
2. `## Before You Start` — Go 1.25; a `*sql.DB` from a SQLite driver for execution (not this repo).
3. `## Step-by-Step Guide` — complete snippets: `c := sqlt.NewCompiler()`; `sqlt.Translate(storage.Query{...}, m)`;
   `sqlt.TranslateDDL(ddl.Stmt{Op: ddl.OpCopyColumn, Table: "holiday", OldName: "label", ColumnName: "name"}, nil)`
   printing the SQL; `c.ExportDDL(models)`. Link [translate.go](translate.go) instead of listing errors.
4. `## Documentation Map` — table: `translate.go`, `tests/conformance_test.go` (contracts proved), `AGENTS.md`.

Commit.

## Delivery check

1. `gotest` (root) passes and `cd tests && go test ./...` passes. A PR that does not compile, or whose tests could not
   run, is a failed delivery.
2. `langc sync .` → zero untranslated keys, zero reuse warnings.
3. `git diff --stat main` lists: `go.mod`, `go.sum`, `tests/go.mod`, `tests/go.sum`, `translate.go`,
   `tests/sqlt_test.go`, `tests/conformance_test.go`, `lang.json`, `README.md`.

Delivery rules: never pause to ask or report progress — when the check passes, commit, push and open the PR; stuck 30
minutes, open the PR listing what fails. Commit after every stage; never `git reset`, `git restore`,
`git checkout -- <file>` or `git stash` on your own work.

## Stages

| Stage | What | Files |
|---|---|---|
| 0 | Upgrade `ddl` (both modules) | `go.mod`, `go.sum`, `tests/go.mod`, `tests/go.sum` |
| 1 | Red tests | `tests/sqlt_test.go` |
| 2 | Code + NULL rule in the test conn | `translate.go`, `tests/conformance_test.go` |
| 3 | Translations | `lang.json` |
| 4 | Docs | `README.md` |
