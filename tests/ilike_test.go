package tests

import (
	"testing"
	"webtyp.com/storage"
	"webtyp.com/sqlt"
)

func TestILike_IgnoresCase(t *testing.T) {
	sql, args, err := sqlt.Translate(storage.Query{
		Action:     storage.ActionReadAll,
		Table:      "users",
		Columns:    []string{"name"},
		Conditions: []storage.Condition{storage.ILike("name", "%AN%")},
	}, testModel{})

	if err != nil {
		t.Fatal(err)
	}

	wantSQL := "SELECT name FROM users WHERE name LIKE ?"
	if sql != wantSQL {
		t.Errorf("got %q, want %q", sql, wantSQL)
	}

	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}

	if args[0] != "%AN%" {
		t.Errorf("got arg %v, want %%AN%%", args[0])
	}
}
