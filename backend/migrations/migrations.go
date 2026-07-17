package migrations

import "embed"

// Files contains every versioned SQL migration compiled into the migration binary.
//
//go:embed *.sql
var Files embed.FS
