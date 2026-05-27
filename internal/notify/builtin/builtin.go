// Package builtin blank-imports all built-in notifier types so their init() registers.
package builtin

import (
	_ "github.com/sunshow/siphongear/internal/notify/serverchan"
)
