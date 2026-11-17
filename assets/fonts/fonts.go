package fonts

import (
	_ "embed"
)

//go:embed Aptos.ttf
var AptosRegular []byte

//go:embed Aptos-Bold.ttf
var AptosBold []byte

//go:embed Aptos-Italic.ttf
var AptosItalic []byte

//go:embed Helvetica.ttf
var HelveticaRegular []byte

//go:embed Helvetica-Bold.ttf
var HelveticaBold []byte

//go:embed Helvetica-Oblique.ttf
var HelveticaItalic []byte

//go:embed Helvetica-BoldOblique.ttf
var HelveticaBoldItalic []byte
