package assets

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"image"
	"image/png"
	"log"
	"runtime"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

//go:embed logo.svg
var logoSVG []byte

var IconData []byte

func init() {
	// 1. Decodifica o SVG do embed byte array
	icon, err := oksvg.ReadIconStream(bytes.NewReader(logoSVG))
	if err != nil {
		log.Fatalf("Falha crítica ao ler VECTORA logo SVG incorporado: %v", err)
	}

	w, h := int(icon.ViewBox.W), int(icon.ViewBox.H)
	icon.SetTarget(0, 0, float64(64), float64(64))

	// 2. Renderiza o SVG vetorizado nativamente numa Matrix RGBA
	rgba := image.NewRGBA(image.Rect(0, 0, 64, 64))
	scanner := rasterx.NewScannerGV(w, h, rgba, rgba.Bounds())
	dasher := rasterx.NewDasher(w, h, scanner)
	icon.Draw(dasher, 1)

	// 3. Codifica para PNG
	var buf bytes.Buffer
	err = png.Encode(&buf, rgba)
	if err != nil {
		log.Fatalf("Falha ao codificar PNG array final da logo SVG: %v", err)
	}
	pngBytes := buf.Bytes()

	// 4. No Windows, o systray EXIGE um cabeçalho ICO válido encapsulando o PNG
	if runtime.GOOS == "windows" {
		IconData = createIco(pngBytes)
	} else {
		IconData = pngBytes
	}
}

// createIco escreve o cabeçalho ICO do Win32 em volta dos bytes do PNG
func createIco(pngBytes []byte) []byte {
	buf := new(bytes.Buffer)
	// ICONDIR: 0 (res), 1 (ico), 1 (count)
	buf.Write([]byte{0, 0, 1, 0, 1, 0})
	
	// ICONDIRENTRY (16 bytes)
	buf.WriteByte(64) // width
	buf.WriteByte(64) // height
	buf.WriteByte(0)  // color count
	buf.WriteByte(0)  // reserved
	buf.WriteByte(1)  // planes
	buf.WriteByte(0)  
	buf.WriteByte(32) // bpp
	buf.WriteByte(0)  
	
	// Size do payload PNG
	binary.Write(buf, binary.LittleEndian, uint32(len(pngBytes)))
	// Offset (Header 6 + Entry 16 = 22)
	binary.Write(buf, binary.LittleEndian, uint32(22))
	
	// Payload da Imagem
	buf.Write(pngBytes)
	
	return buf.Bytes()
}
