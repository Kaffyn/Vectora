package assets

import _ "embed"

//go:embed logo.ico
var IconData []byte

//go:embed test_icon.ico
var TestIconData []byte

//go:embed installer_icon.ico
var InstallerIconData []byte
