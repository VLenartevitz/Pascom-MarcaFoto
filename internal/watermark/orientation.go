package watermark

import (
	"image"
	"io"
	"os"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
)

// ExifOrientation lê o tag Orientation (1-8) do JPEG.
// Retorna 1 (normal) quando ausente, inválido ou em caso de erro —
// a preservação de metadados não é presumida (SDD §9), apenas a orientação
// de exibição é respeitada.
func ExifOrientation(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 1
	}
	defer f.Close()
	return ExifOrientationReader(f)
}

// ExifOrientationReader é a variante testável que lê de qualquer Reader.
func ExifOrientationReader(r io.Reader) int {
	x, err := exif.Decode(r)
	if err != nil {
		return 1
	}
	tag, err := x.Get(exif.Orientation)
	if err != nil {
		return 1
	}
	v, err := tag.Int(0)
	if err != nil || v < 1 || v > 8 {
		return 1
	}
	return v
}

// ApplyOrientation rotaciona/inverte a imagem decodificada conforme o
// valor EXIF, para que a foto composta corresponda à exibição correta.
func ApplyOrientation(img image.Image, o int) image.Image {
	switch o {
	case 2:
		return imaging.FlipH(img)
	case 3:
		return imaging.Rotate180(img)
	case 4:
		return imaging.FlipV(img)
	case 5:
		return imaging.Transpose(img)
	case 6:
		return imaging.Rotate270(img)
	case 7:
		return imaging.Transverse(img)
	case 8:
		return imaging.Rotate90(img)
	default:
		return img
	}
}
