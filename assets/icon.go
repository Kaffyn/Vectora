package assets

import _ "embed"

// Core application icons
//go:embed logo.ico
var VectoraIconData []byte

//go:embed cli_icon.ico
var VectoraTUIIconData []byte

//go:embed app_icon.ico
var VectoraDesktopIconData []byte

//go:embed installer_icon.ico
var VectoraSetupIconData []byte

//go:embed llama_logo.ico
var LPMIconData []byte

//go:embed app_icon.ico
var MPMIconData []byte

// Legacy aliases for compatibility
//go:embed logo.ico
var IconData []byte

//go:embed test_icon.ico
var TestIconData []byte

//go:embed installer_icon.ico
var InstallerIconData []byte
