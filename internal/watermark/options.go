// Package watermark aplica marca d'água de imagem em fotografias.
// Fase 1: implementação Pure-Go (stdlib + imaging), sem CGO/libvips,
// para que o cross-compile GOOS=windows funcione dentro de container Linux.
package watermark

import (
	"fmt"
	"strings"
)

// Position é um dos 9 pontos de ancoragem da logo sobre a foto.
type Position int

const (
	TopLeft Position = iota
	TopCenter
	TopRight
	MiddleLeft
	Center
	MiddleRight
	BottomLeft
	BottomCenter
	BottomRight
)

// ParsePosition converte nomes como "bottom-right" / "center".
func ParsePosition(s string) (Position, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "top-left":
		return TopLeft, nil
	case "top-center", "top":
		return TopCenter, nil
	case "top-right":
		return TopRight, nil
	case "middle-left", "left":
		return MiddleLeft, nil
	case "center", "middle-center", "middle":
		return Center, nil
	case "middle-right", "right":
		return MiddleRight, nil
	case "bottom-left":
		return BottomLeft, nil
	case "bottom-center", "bottom":
		return BottomCenter, nil
	case "bottom-right":
		return BottomRight, nil
	default:
		return Center, fmt.Errorf("posição inválida: %q (use ex.: bottom-right, center, top-left)", s)
	}
}

// OutputFormat é o formato de exportação.
type OutputFormat string

const (
	FormatJPEG OutputFormat = "jpeg"
	FormatPNG  OutputFormat = "png"
)

// ParseFormat converte "jpeg"/"jpg"/"png".
func ParseFormat(s string) (OutputFormat, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "jpeg", "jpg":
		return FormatJPEG, nil
	case "png":
		return FormatPNG, nil
	default:
		return FormatJPEG, fmt.Errorf("formato inválido: %q (use jpeg ou png)", s)
	}
}

// Ext retorna a extensão de arquivo do formato (com ponto).
func (f OutputFormat) Ext() string {
	if f == FormatPNG {
		return ".png"
	}
	return ".jpg"
}

// Options descreve como a marca d'água é aplicada, de forma
// independente das dimensões de cada fotografia (RF-03).
type Options struct {
	Position      Position
	SizePercent   float64 // largura da logo em % da largura da foto (1-100)
	MarginPercent float64 // margem em % da largura da foto (>=0)
	Opacity       float64 // 0 (invisível) a 1 (opaca)
	Quality       int     // qualidade JPEG 1-100
	Format        OutputFormat
}

// DefaultOptions retorna a configuração inicial da interface.
func DefaultOptions() Options {
	return Options{
		Position:      BottomRight,
		SizePercent:   20,
		MarginPercent: 3,
		Opacity:       0.7,
		Quality:       92,
		Format:        FormatJPEG,
	}
}

// Validate verifica ranges antes de qualquer processamento.
func (o Options) Validate() error {
	if o.SizePercent <= 0 || o.SizePercent > 100 {
		return fmt.Errorf("tamanho deve estar entre 1 e 100 (%% da largura): %v", o.SizePercent)
	}
	if o.MarginPercent < 0 || o.MarginPercent > 40 {
		return fmt.Errorf("margem deve estar entre 0 e 40 (%% da largura): %v", o.MarginPercent)
	}
	if o.Opacity < 0 || o.Opacity > 1 {
		return fmt.Errorf("opacidade deve estar entre 0 e 1: %v", o.Opacity)
	}
	if o.Quality < 1 || o.Quality > 100 {
		return fmt.Errorf("qualidade JPEG deve estar entre 1 e 100: %d", o.Quality)
	}
	if o.Format != FormatJPEG && o.Format != FormatPNG {
		return fmt.Errorf("formato inválido: %q", o.Format)
	}
	return nil
}
