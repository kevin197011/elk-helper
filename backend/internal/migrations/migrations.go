// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

// Package migrations provides embedded database migration files
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
