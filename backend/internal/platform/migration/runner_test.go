package migration

import (
	"testing"
	"testing/fstest"
)

func TestLoadMigrationFilesOrdersAndPairsFiles(t *testing.T) {
	files := fstest.MapFS{
		"000002_second.down.sql": {Data: []byte("SELECT 2")},
		"000001_first.up.sql":    {Data: []byte("SELECT 1")},
		"000002_second.up.sql":   {Data: []byte("SELECT 2")},
		"000001_first.down.sql":  {Data: []byte("SELECT 1")},
	}

	migrations, err := loadMigrationFiles(files)
	if err != nil {
		t.Fatalf("loadMigrationFiles() error = %v", err)
	}
	if len(migrations) != 2 || migrations[0].version != 1 || migrations[1].version != 2 {
		t.Fatalf("migration order = %#v, want versions 1 and 2", migrations)
	}
}

func TestLoadMigrationFilesRejectsMissingPair(t *testing.T) {
	files := fstest.MapFS{
		"000001_first.up.sql": {Data: []byte("SELECT 1")},
	}

	if _, err := loadMigrationFiles(files); err == nil {
		t.Fatal("loadMigrationFiles() error = nil, want missing pair error")
	}
}
