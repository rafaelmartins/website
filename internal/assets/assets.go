package assets

import "embed"

//go:embed embed/main.*
var Main embed.FS

//go:embed embed/project.*
var Project embed.FS

//go:embed embed/cdocs.*
var CDocs embed.FS

//go:embed embed/dfu-flasher.* embed/dfuse.js
var DfuFlasher embed.FS

//go:embed embed/search.*
var Search embed.FS
