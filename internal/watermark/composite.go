package watermark

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	"watermark-studio/internal/files"
)

// DecodeFile decodifica uma fotografia e aplica a orientação EXIF (JPEG).
func DecodeFile(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	if isJPEG(path) {
		if o := ExifOrientation(path); o != 1 {
			img = ApplyOrientation(img, o)
		}
	}
	return img, nil
}

// DecodeLogo decodifica a logo. O ideal é PNG com transparência (RF-02);
// outros formatos decodificáveis são aceitos, com a opacidade aplicada
// uniformemente sobre a imagem.
func DecodeLogo(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("logo: %w", err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("logo inválida (%s): %w", path, err)
	}
	b := img.Bounds()
	if b.Dx() < 1 || b.Dy() < 1 {
		return nil, fmt.Errorf("logo com dimensões inválidas: %s", path)
	}
	return img, nil
}

// Apply compõe a logo sobre a foto conforme opts.
// As dimensões da foto são preservadas (RNF-03).
func Apply(base, logo image.Image, opts Options) (image.Image, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	bb := base.Bounds()
	lb := logo.Bounds()
	tw, th := ScaledLogoSize(bb.Dx(), bb.Dy(), lb.Dx(), lb.Dy(), opts.SizePercent)
	if tw < 1 || th < 1 {
		return nil, fmt.Errorf("escala resultou em logo vazia")
	}
	scaled := imaging.Resize(logo, tw, th, imaging.Lanczos)
	rect := LogoRect(bb.Dx(), bb.Dy(), tw, th, opts.Position, opts.MarginPercent)
	return imaging.Overlay(base, scaled, rect.Min, opts.Opacity), nil
}

// Encode grava a imagem composta no formato configurado.
// JPEG usa a qualidade configurada; PNG ignora qualidade (sem perdas).
func Encode(w io.Writer, img image.Image, opts Options) error {
	if opts.Format == FormatPNG {
		return imaging.Encode(w, img, imaging.PNG)
	}
	return imaging.Encode(w, img, imaging.JPEG, imaging.JPEGQuality(opts.Quality))
}

// ProcessOne executa o pipeline completo de um arquivo:
// foto + logo (já carregada) -> saída com escrita atômica.
func ProcessOne(photoPath string, logo image.Image, outPath string, opts Options) error {
	base, err := DecodeFile(photoPath)
	if err != nil {
		return err
	}
	composed, err := Apply(base, logo, opts)
	if err != nil {
		return err
	}
	if err := files.WriteAtomic(outPath, func(w io.Writer) error {
		return Encode(w, composed, opts)
	}); err != nil {
		return fmt.Errorf("exportação: %w", err)
	}
	return nil
}

// Preview gera uma pré-visualização reduzida (largura máxima maxWidth px,
// sem upscale) com a marca aplicada — evita operações pesadas na UI (RF-04).
func Preview(photoPath string, logo image.Image, opts Options, maxWidth int) (image.Image, error) {
	base, err := DecodeFile(photoPath)
	if err != nil {
		return nil, err
	}
	if maxWidth > 0 && base.Bounds().Dx() > maxWidth {
		base = imaging.Resize(base, maxWidth, 0, imaging.Lanczos)
	}
	return Apply(base, logo, opts)
}

func isJPEG(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".jpg" || ext == ".jpeg"
}
