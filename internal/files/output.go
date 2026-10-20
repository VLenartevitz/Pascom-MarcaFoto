package files

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// EnsureOutDir cria a pasta de saída e recusa usar a pasta de entrada
// como destino — os originais nunca são sobrescritos (SDD §9).
func EnsureOutDir(outDir, inDir string) error {
	absOut, err := filepath.Abs(outDir)
	if err != nil {
		return err
	}
	absIn, err := filepath.Abs(inDir)
	if err != nil {
		return err
	}
	if absOut == absIn {
		return fmt.Errorf("pasta de saída não pode ser a pasta de entrada (originais preservados)")
	}
	return os.MkdirAll(absOut, 0o755)
}

// ResolveOutput aplica a política de conflito (RF-06): sem --overwrite,
// gera nome sequencial (foto_1.jpg) em vez de sobrescrever.
func ResolveOutput(outDir, base string, overwrite bool) (string, error) {
	cand := filepath.Join(outDir, base)
	if overwrite {
		return cand, nil
	}
	if _, err := os.Stat(cand); os.IsNotExist(err) {
		return cand, nil
	}
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	for i := 1; i <= 99999; i++ {
		c := filepath.Join(outDir, fmt.Sprintf("%s_%d%s", stem, i, ext))
		if _, err := os.Stat(c); os.IsNotExist(err) {
			return c, nil
		}
	}
	return "", fmt.Errorf("sem nome disponível para %s", base)
}

// WriteAtomic grava em arquivo temporário no mesmo diretório e só então
// move para o destino definitivo — exportação segura (SDD §5 passo 7).
// Em caso de erro, o temporário incompleto é removido.
func WriteAtomic(path string, write func(w io.Writer) error) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	failed := true
	defer func() {
		if failed {
			_ = os.Remove(name)
		}
	}()
	if err := write(tmp); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(name, path); err != nil {
		// Windows não permite Rename sobre existente: remove e tenta de novo.
		if rerr := os.Remove(path); rerr == nil {
			if rerr = os.Rename(name, path); rerr == nil {
				failed = false
				return nil
			} else {
				err = rerr
			}
		}
		return err
	}
	failed = false
	return nil
}
