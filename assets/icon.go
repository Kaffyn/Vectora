package assets

import _ "embed"

// Core application icons
//
//go:embed vectora.ico
var VectoraIconData []byte

//go:embed tui.ico
var VectoraTUIIconData []byte

//go:embed desktop.ico
var VectoraDesktopIconData []byte

//go:embed setup.ico
var VectoraSetupIconData []byte

//go:embed lpm.ico
var LPMIconData []byte

//go:embed vectora.ico
var MPMIconData []byte

// Legacy aliases for compatibility
//
//go:embed vectora.ico
var IconData []byte

//go:embed test.ico
var TestIconData []byte

//go:embed setup.ico
var InstallerIconData []byte
