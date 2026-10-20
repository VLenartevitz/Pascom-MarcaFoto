package watermark

import "testing"

func TestScaledLogoSize(t *testing.T) {
	w, h := ScaledLogoSize(1000, 800, 400, 200, 20)
	if w != 200 || h != 100 {
		t.Errorf("esperado 200x100, obtido %dx%d", w, h)
	}
	// Logo alta limitada pela altura da foto.
	w, h = ScaledLogoSize(1000, 100, 100, 400, 50)
	if h != 100 || w != 25 {
		t.Errorf("limite de altura: esperado 25x100, obtido %dx%d", w, h)
	}
	if w, h := ScaledLogoSize(0, 100, 100, 100, 20); w != 0 || h != 0 {
		t.Errorf("dimensões zeradas deveriam retornar 0x0")
	}
}

func TestLogoRectAllPositions(t *testing.T) {
	const pw, ph, lw, lh = 1000, 800, 200, 100
	const margin = 3 // 30px
	type want struct {
		pos       Position
		x, y      int
	}
	cases := []want{
		{TopLeft, 30, 30}, {TopCenter, 400, 30}, {TopRight, 770, 30},
		{MiddleLeft, 30, 350}, {Center, 400, 350}, {MiddleRight, 770, 350},
		{BottomLeft, 30, 670}, {BottomCenter, 400, 670}, {BottomRight, 770, 670},
	}
	for _, c := range cases {
		r := LogoRect(pw, ph, lw, lh, c.pos, margin)
		if r.Min.X != c.x || r.Min.Y != c.y || r.Dx() != lw || r.Dy() != lh {
			t.Errorf("pos %d: obtido %v, esperado min=(%d,%d) tam=%dx%d", c.pos, r, c.x, c.y, lw, lh)
		}
	}
}

func TestLogoRectContained(t *testing.T) {
	// Logo maior que a foto: deve ser contida, nunca vazar.
	r := LogoRect(100, 100, 500, 500, BottomRight, 3)
	if r.Min.X < 0 || r.Min.Y < 0 || r.Max.X > 100 || r.Max.Y > 100 {
		t.Errorf("retângulo vazou da foto: %v", r)
	}
}
