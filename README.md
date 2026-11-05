# webtyp/sqlt
<img src="docs/img/badges.svg">

SQLite SQL compiler. Implements `storage.Compiler` (DML) and `ddl.Compiler` (DDL) —
translates `storage.Query`/`ddl.Stmt` into SQLite SQL. `ExportDDL` builds full-schema
SQL (ordered via `ddl.TopologicalSort`) for build-time schema generation. Pure compiler:
no connections, no execution — see [webtyp/sqlite](https://github.com/webtyp/sqlite) for the driver that wraps a real
`*sql.DB` in a `storage.Conn` and [webtyp/ddl](https://github.com/webtyp/ddl) for DDL operations.

## Why This Exists

This repository translates storage operations into SQLite queries without managing database connections.
Without it, each application would have to write SQLite SQL by hand for every schema or query change.
With `sqlt`, functions like `ddl.Sync`, `ddl.CopyColumn`, `ddl.DropColumn`, and `storage` queries automatically render through `sqlt` into SQL strings.

## Before You Start

- Requires Go 1.25 or newer.
- Requires a `*sql.DB` connection instance from a standard SQLite driver for actual execution (not provided by this module).

## Step-by-Step Guide

```go
package example

import (
	"webtyp.com/ddl"
	"webtyp.com/model"
	"webtyp.com/sqlt"
	"webtyp.com/storage"
)

func Example() {
	// Initialize compiler
	c := sqlt.NewCompiler()

	// Export Data Definition Language
	// sql, err := c.ExportDDL(models)

	// Translate DML
	// sql, args, err := sqlt.Translate(storage.Query{...}, m)

	// Translate DDL (example: copying a column)
	sql, args, err := sqlt.TranslateDDL(ddl.Stmt{
		Op:         ddl.OpCopyColumn,
		Table:      "holiday",
		OldName:    "label",
		ColumnName: "name",
	}, nil)
	// sql is: "UPDATE holiday SET name = label WHERE name IS NULL AND label IS NOT NULL"
}
```
For API implementation details, refer to [translate.go](translate.go).

## Documentation Map

| File | Purpose |
|------|---------|
| `translate.go` | DML and DDL SQLite generation logic |
| `tests/conformance_test.go` | SQL driver compliance contracts proved |
| `AGENTS.md` | Development and contribution rules |
