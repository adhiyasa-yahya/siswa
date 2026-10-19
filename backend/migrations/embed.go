// Package migrations embeds the SQL migration files into the binary so the API
// container can migrate itself on boot without an extra migration tool.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
