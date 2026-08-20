package storage

import _ "embed"

//go:embed seed/roles.sql
var seedRoles string

//go:embed seed/workflow.sql
var seedWorkflow string
