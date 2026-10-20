package watermark

import "image"

// ScaledLogoSize calcula o tamanho da logo preservando a proporção:
// a largura-alvo é SizePercent% da largura da foto. Se a altura resultante
// exceder a foto, o cálculo é refeito limitado pela altura.
func ScaledLogoSize(photoW, photoH, logoW, logoH int, sizePercent float64) (int, int) {
	if photoW <= 0 || photoH <= 0 || logoW <= 0 || logoH <= 0 {
		return 0, 0
	}
	targetW := int(float64(photoW) * sizePercent / 100.0)
	if targetW < 1 {
		targetW = 1
	}
	if targetW > photoW {
		targetW = photoW
	}
	targetH := int(float64(targetW) * float64(logoH) / float64(logoW))
	if targetH < 1 {
		targetH = 1
	}
	if targetH > photoH {
		targetH = photoH
		targetW = int(float64(targetH) * float64(logoW) / float64(logoH))
		if targetW < 1 {
			targetW = 1
		}
	}
	return targetW, targetH
}

// LogoRect posiciona o retângulo da logo (já redimensionada) sobre a foto,
// aplicando a margem em % da largura da foto. O retângulo nunca excede a foto:
// se a logo for maior que a área útil, ela é contida mantendo a âncora.
func LogoRect(photoW, photoH, logoW, logoH int, pos Position, marginPercent float64) image.Rectangle {
	margin := int(float64(photoW) * marginPercent / 100.0)
	if margin < 0 {
		margin = 0
	}
	w, h := logoW, logoH
	if maxW := photoW - 2*margin; maxW > 0 && w > maxW {
		w = maxW
	}
	if maxH := photoH - 2*margin; maxH > 0 && h > maxH {
		h = maxH
	}
	var x, y int
	switch pos {
	case TopLeft:
		x, y = margin, margin
	case TopCenter:
		x, y = (photoW-w)/2, margin
	case TopRight:
		x, y = photoW-w-margin, margin
	case MiddleLeft:
		x, y = margin, (photoH-h)/2
	case Center:
		x, y = (photoW-w)/2, (photoH-h)/2
	case MiddleRight:
		x, y = photoW-w-margin, (photoH-h)/2
	case BottomLeft:
		x, y = margin, photoH-h-margin
	case BottomCenter:
		x, y = (photoW-w)/2, photoH-h-margin
	case BottomRight:
		x, y = photoW-w-margin, photoH-h-margin
	}
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	return image.Rect(x, y, x+w, y+h)
}
