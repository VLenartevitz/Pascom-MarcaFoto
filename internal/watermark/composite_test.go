package watermark

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// syntheticBase cria uma foto sintética 800x600 cinza.
func syntheticBase() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 800, 600))
	gray := color.RGBA{128, 128, 128, 255}
	for y := 0; y < 600; y++ {
		for x := 0; x < 800; x++ {
			img.Set(x, y, gray)
		}
	}
	return img
}

// syntheticLogo cria uma logo 200x100: metade opaca vermelha, metade transparente.
func syntheticLogo() image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, 200, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 200; x++ {
			if x < 100 {
				img.Set(x, y, color.NRGBA{255, 0, 0, 255})
			} else {
				img.Set(x, y, color.NRGBA{0, 0, 0, 0})
			}
		}
	}
	return img
}

func TestApplyPreservesSize(t *testing.T) {
	opts := DefaultOptions()
	out, err := Apply(syntheticBase(), syntheticLogo(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if b := out.Bounds(); b.Dx() != 800 || b.Dy() != 600 {
		t.Errorf("dimensões não preservadas: %v", b)
	}
}

func TestApplyMarksExpectedArea(t *testing.T) {
	opts := DefaultOptions()
	opts.Opacity = 1
	out, err := Apply(syntheticBase(), syntheticLogo(), opts)
	if err != nil {
		t.Fatal(err)
	}
	// Logo 20% de 800 = 160px de largura, 80 de altura, canto inferior direito, margem 3% = 24px.
	// Amostra no centro da metade opaca: x = 800-24-160+40, y = 600-24-80+40.
	r, g, b, _ := out.At(800-24-160+40, 600-24-80+40).RGBA()
	if r < 0x8000 || g > 0x7FFF || b > 0x7FFF {
		t.Errorf("pixel da marca deveria ser avermelhado, obtido (%x,%x,%x)", r, g, b)
	}
	// Canto oposto deve permanecer cinza.
	r, g, b, _ = out.At(10, 10).RGBA()
	if r < 0x7F00 || r > 0x8100 || g < 0x7F00 || g > 0x8100 {
		t.Errorf("canto oposto foi alterado: (%x,%x,%x)", r, g, b)
	}
}

func TestOpacityZeroKeepsBase(t *testing.T) {
	opts := DefaultOptions()
	opts.Opacity = 0
	out, err := Apply(syntheticBase(), syntheticLogo(), opts)
	if err != nil {
		t.Fatal(err)
	}
	r, g, b, _ := out.At(700, 550).RGBA()
	if r < 0x7F00 || r > 0x8100 || g < 0x7F00 || g > 0x8100 || b < 0x7F00 || b > 0x8100 {
		t.Errorf("opacidade 0 deveria manter a base: (%x,%x,%x)", r, g, b)
	}
}

func TestProcessOneRoundTrip(t *testing.T) {
	dir := t.TempDir()
	photo := filepath.Join(dir, "foto.jpg")
	logoP := filepath.Join(dir, "logo.png")
	out := filepath.Join(dir, "saida.jpg")

	f, _ := os.Create(photo)
	_ = jpeg.Encode(f, syntheticBase(), &jpeg.Options{Quality: 95})
	_ = f.Close()
	g, _ := os.Create(logoP)
	_ = png.Encode(g, syntheticLogo())
	_ = g.Close()

	logo, err := DecodeLogo(logoP)
	if err != nil {
		t.Fatal(err)
	}
	opts := DefaultOptions()
	if err := ProcessOne(photo, logo, out, opts); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(out)
	if err != nil || st.Size() == 0 {
		t.Fatalf("saída ausente ou vazia: %v", err)
	}
	// Original intacto: ainda decodifica com as mesmas dimensões.
	back, err := DecodeFile(photo)
	if err != nil {
		t.Fatal(err)
	}
	if b := back.Bounds(); b.Dx() != 800 || b.Dy() != 600 {
		t.Errorf("original alterado: %v", b)
	}
}

func TestExifOrientationReader(t *testing.T) {
	// Sem EXIF: deve retornar 1 sem erro.
	if o := ExifOrientationReader(bytes.NewReader([]byte("sem exif aqui"))); o != 1 {
		t.Errorf("esperado 1, obtido %d", o)
	}
}
