package assets

import _ "embed"

//go:embed logo.ico
var IconData []byte

//go:embed test_icon.svg
var TestIconData []byte

//go:embed installer_icon.svg
var InstallerIconData []byte
