// Command gen gera fotos sintéticas + logo para testes e benchmarks.
// Os arquivos gerados ficam em tests/testdata/{in,out} (gitignored);
// apenas o gerador e a logo de referência são versionados.
//
//	go run ./tests/gen -n 100
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
)

func main() {
	n := flag.Int("n", 20, "quantidade de fotos")
	root := flag.String("root", "tests/testdata", "pasta base")
	flag.Parse()

	inDir := filepath.Join(*root, "in")
	outDir := filepath.Join(*root, "out")
	if err := os.MkdirAll(inDir, 0o755); err != nil {
		fatal(err)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fatal(err)
	}
	rng := rand.New(rand.NewSource(42))

	logoPath := filepath.Join(*root, "logo.png")
	if err := writePNG(logoPath, makeLogo()); err != nil {
		fatal(err)
	}
	sizes := [][2]int{{1600, 1200}, {1200, 1600}, {2000, 1500}, {800, 600}}
	for i := 0; i < *n; i++ {
		s := sizes[i%len(sizes)]
		name := filepath.Join(inDir, fmt.Sprintf("foto_%03d.jpg", i+1))
		if err := writeJPEG(name, makePhoto(s[0], s[1], rng), 90); err != nil {
			fatal(err)
		}
	}
	fmt.Printf("geradas: %d fotos em %s + %s\n", *n, inDir, logoPath)
}

// makePhoto gera gradiente + faixas (determinístico via rng).
func makePhoto(w, h int, rng *rand.Rand) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	r0, g0, b0 := uint8(rng.Intn(200)), uint8(rng.Intn(200)), uint8(rng.Intn(200))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			f := float64(x) / float64(w)
			img.Set(x, y, color.RGBA{
				uint8(float64(r0) * (0.4 + 0.6*f)),
				uint8(float64(g0) * (0.4 + 0.6*float64(y) / float64(h))),
				uint8(float64(b0) * (0.6 + 0.4*f)),
				255,
			})
		}
	}
	// Faixa diagonal clara para dar textura à compressão.
	for i := 0; i < w; i++ {
		for k := 0; k < 8; k++ {
			y := (i*3/2 + k) % h
			img.Set(i, y, color.RGBA{240, 240, 240, 255})
		}
	}
	return img
}

// makeLogo gera logo 800x400 com transparência: retângulo + círculo opacos.
func makeLogo() image.Image {
	const w, h = 800, 400
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 60; y < h-60; y++ {
		for x := 60; x < w-60; x++ {
			img.Set(x, y, color.NRGBA{20, 60, 160, 255})
		}
	}
	cx, cy, r := w/2, h/2, 90
	for y := cy - r; y <= cy+r; y++ {
		for x := cx - r; x <= cx+r; x++ {
			dx, dy := x-cx, y-cy
			if dx*dx+dy*dy <= r*r {
				img.Set(x, y, color.NRGBA{255, 255, 255, 255})
			}
		}
	}
	return img
}

func writeJPEG(path string, img image.Image, q int) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, img, &jpeg.Options{Quality: q})
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "erro:", err)
	os.Exit(1)
}
