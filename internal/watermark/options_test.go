package watermark

import "testing"

func TestParsePosition(t *testing.T) {
	cases := map[string]Position{
		"top-left": TopLeft, "top": TopCenter, "top-center": TopCenter,
		"top-right": TopRight, "left": MiddleLeft, "middle-left": MiddleLeft,
		"center": Center, "middle": Center, "right": MiddleRight,
		"middle-right": MiddleRight, "bottom-left": BottomLeft,
		"bottom": BottomCenter, "bottom-center": BottomCenter,
		"bottom-right": BottomRight,
	}
	for in, want := range cases {
		got, err := ParsePosition(in)
		if err != nil || got != want {
			t.Errorf("ParsePosition(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
	if _, err := ParsePosition("canto"); err == nil {
		t.Error("posição inválida deveria retornar erro")
	}
}

func TestValidate(t *testing.T) {
	ok := DefaultOptions()
	if err := ok.Validate(); err != nil {
		t.Fatalf("default deveria ser válido: %v", err)
	}
	bad := []Options{
		{Position: Center, SizePercent: 0, MarginPercent: 3, Opacity: 0.7, Quality: 92, Format: FormatJPEG},
		{Position: Center, SizePercent: 101, MarginPercent: 3, Opacity: 0.7, Quality: 92, Format: FormatJPEG},
		{Position: Center, SizePercent: 20, MarginPercent: -1, Opacity: 0.7, Quality: 92, Format: FormatJPEG},
		{Position: Center, SizePercent: 20, MarginPercent: 3, Opacity: 1.5, Quality: 92, Format: FormatJPEG},
		{Position: Center, SizePercent: 20, MarginPercent: 3, Opacity: 0.7, Quality: 0, Format: FormatJPEG},
		{Position: Center, SizePercent: 20, MarginPercent: 3, Opacity: 0.7, Quality: 92, Format: "bmp"},
	}
	for i, o := range bad {
		if err := o.Validate(); err == nil {
			t.Errorf("caso %d deveria ser inválido: %+v", i, o)
		}
	}
}
