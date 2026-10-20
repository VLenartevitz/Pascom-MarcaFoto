package files

import (
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// SupportedExt são as extensões aceitas na triagem (RF-01).
// A confirmação final é pelo conteúdo (magic bytes).
var SupportedExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true,
	".tif": true, ".tiff": true, ".bmp": true,
	".gif": true, ".webp": true,
}

// ScanResult descreve o que foi encontrado na pasta de entrada.
type ScanResult struct {
	Files   []string // caminhos absolutos das imagens compatíveis
	Ignored int      // extensões não-imagem (ex.: .txt, .db)
	Invalid int      // extensão de imagem mas conteúdo ilegível/corrompido
}

// ScanFolder identifica imagens compatíveis recursivamente (RF-01).
func ScanFolder(root string) (ScanResult, error) {
	var res ScanResult
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			res.Invalid++
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !SupportedExt[strings.ToLower(filepath.Ext(p))] {
			res.Ignored++
			return nil
		}
		ok, serr := sniffImage(p)
		if serr != nil || !ok {
			res.Invalid++
			return nil
		}
		abs, aerr := filepath.Abs(p)
		if aerr != nil {
			res.Invalid++
			return nil
		}
		res.Files = append(res.Files, abs)
		return nil
	})
	return res, err
}

func sniffImage(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	head := make([]byte, 512)
	n, _ := io.ReadFull(f, head)
	if n == 0 {
		return false, nil
	}
	switch http.DetectContentType(head[:n]) {
	case "image/jpeg", "image/png", "image/gif", "image/bmp", "image/tiff", "image/webp":
		return true, nil
	default:
		return false, nil
	}
}
