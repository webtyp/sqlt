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
package main

import (
	"log"

	"webtyp.com/ddl"
	"webtyp.com/model"
	"webtyp.com/sqlt"
	"webtyp.com/storage"
)

func main() {
	// DML: a storage.Query becomes SQLite SQL plus positional args.
	sql, args, err := sqlt.Translate(storage.Query{
		Action:     storage.ActionReadAll,
		Table:      "holiday",
		Columns:    []string{"name"},
		Conditions: []storage.Condition{storage.IsNotNull("name")},
	}, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(sql, args) // SELECT name FROM holiday WHERE name IS NOT NULL []

	// DDL: copy a removed column into the one that replaced it, only where it is empty.
	sql, _, err = sqlt.TranslateDDL(ddl.Stmt{
		Op:         ddl.OpCopyColumn,
		Table:      "holiday",
		OldName:    "label",
		ColumnName: "name",
	}, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(sql) // UPDATE holiday SET name = label WHERE name IS NULL AND label IS NOT NULL

	// Full schema for build-time tooling, tables ordered by their foreign keys.
	var models []model.Model // your ormc-generated models
	schema, err := sqlt.NewCompiler().ExportDDL(models)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(schema)
}
```
For API implementation details, refer to [translate.go](translate.go).

## Documentation Map

| File | Purpose |
|------|---------|
| `translate.go` | DML and DDL SQLite generation logic |
| `tests/translate_test.go` | SQL text generated for each query and DDL step |
| `AGENTS.md` | Development and contribution rules |
