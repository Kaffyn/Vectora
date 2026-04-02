package assets

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
	"log"

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

	// 2. Traça o alvo na dimensão ideal, re-escalando o desenho do vetor pra 64x64 pixels do Systray
	w, h := int(icon.ViewBox.W), int(icon.ViewBox.H)
	icon.SetTarget(0, 0, float64(64), float64(64))

	// 3. Renderiza o SVG vetorizado nativamente numa Matrix RGBA nova em RAM
	rgba := image.NewRGBA(image.Rect(0, 0, 64, 64))
	
	// Utiliza o scanner de software do Rasterx para desenhar na nossa Matrix Go pura
	scanner := rasterx.NewScannerGV(w, h, rgba, rgba.Bounds())
	dasher := rasterx.NewDasher(w, h, scanner)
	icon.Draw(dasher, 1)

	// 4. Codifica a Image construida pra byte buffer padrão de array com PNG format type (suportado pelo Windows native shell handle)
	var buf bytes.Buffer
	err = png.Encode(&buf, rgba)
	if err != nil {
		log.Fatalf("Falha ao codificar PNG array final da logo SVG: %v", err)
	}
	
	IconData = buf.Bytes()
}
