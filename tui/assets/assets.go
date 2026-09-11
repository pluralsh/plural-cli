package assets

import _ "embed"

// Logo is the terminal-cell interpretation of plural-logo.png. It is embedded
// so installed binaries do not depend on their working directory.
//
//go:embed logo.txt
var Logo string
