package files

import (
	"bytes"
	"encoding/binary"
)

// Favicon monta um favicon.ico 16x16 32bpp na paleta da marca: quadrado âmbar
// com borda escura. Gerado em código para o scaffold não depender de binário
// embutido no repositório.
//
// Formato ICO: ICONDIR + ICONDIRENTRY + DIB (BITMAPINFOHEADER com altura
// dobrada, pixels BGRA de baixo para cima, máscara AND zerada = tudo opaco)
// essa parte foi gerada com Inteligencia Artificial, talvez precise de ajustes manuais.
func Favicon() []byte {
	const n = 16
	var px bytes.Buffer
	for y := n - 1; y >= 0; y-- { // bottom-up
		for x := range n {
			if x < 2 || x >= n-2 || y < 2 || y >= n-2 {
				px.Write([]byte{0x10, 0x14, 0x1A, 0xFF}) // #1A1410 em BGRA
			} else {
				px.Write([]byte{0x00, 0xB3, 0xFF, 0xFF}) // #FFB300 em BGRA
			}
		}
	}
	andMask := make([]byte, n*4)

	dibSize := uint32(40 + px.Len() + len(andMask))

	var b bytes.Buffer
	le := binary.LittleEndian
	binary.Write(&b, le, [3]uint16{0, 1, 1})
	b.Write([]byte{n, n, 0, 0})
	binary.Write(&b, le, [2]uint16{1, 32})
	binary.Write(&b, le, [2]uint32{dibSize, 22})
	binary.Write(&b, le, [3]uint32{40, n, n * 2})
	binary.Write(&b, le, [2]uint16{1, 32})
	binary.Write(&b, le, [6]uint32{0, 0, 0, 0, 0, 0})
	b.Write(px.Bytes())
	b.Write(andMask)
	return b.Bytes()
}
