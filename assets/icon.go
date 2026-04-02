package assets

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

var IconData []byte

func init() {
	// Cria uma imagem em memória de 16x16 pixels para injetar no tray
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	blue := color.RGBA{R: 0, G: 122, B: 255, A: 255}
	
	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			// Desenha um círculo azul (raio = 8)
			dx := x - 8
			dy := y - 8
			if dx*dx+dy*dy <= 64 {
				img.Set(x, y, blue)
			}
		}
	}
	
	var buf bytes.Buffer
	png.Encode(&buf, img)
	IconData = buf.Bytes()
}
