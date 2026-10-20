// Command watermark é o CLI da Fase 1 (scan/apply/batch/preview).
// Uso típico:
//
//	watermark scan --in ./fotos
//	watermark apply --in ./fotos/a.jpg --logo ./logo.png --out ./saida/a.jpg
//	watermark batch --in ./fotos --logo ./logo.png --out ./saida
//	watermark preview --in ./fotos/a.jpg --logo ./logo.png --out ./prev.jpg
package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"watermark-studio/internal/batch"
	"watermark-studio/internal/files"
	"watermark-studio/internal/watermark"
)

type commonFlags struct {
	pos, format      string
	size, margin     float64
	opacity          float64
	quality          int
	logo, out        string
	workers          int
	overwrite        bool
	maxWidth         int
}

func (c *commonFlags) register(fs *flag.FlagSet, needLogo, needOut, needBatch bool) {
	fs.StringVar(&c.pos, "pos", "bottom-right", "posição: top-left, top-center, top-right, middle-left, center, middle-right, bottom-left, bottom-center, bottom-right")
	fs.Float64Var(&c.size, "size", 20, "largura da logo em % da largura da foto (1-100)")
	fs.Float64Var(&c.margin, "margin", 3, "margem em % da largura da foto (0-40)")
	fs.Float64Var(&c.opacity, "opacity", 0.7, "opacidade 0-1")
	fs.IntVar(&c.quality, "quality", 92, "qualidade JPEG 1-100")
	fs.StringVar(&c.format, "format", "jpeg", "formato de saída: jpeg ou png")
	if needLogo {
		fs.StringVar(&c.logo, "logo", "", "arquivo da logo (PNG com transparência)")
	}
	if needOut {
		fs.StringVar(&c.out, "out", "", "arquivo/pasta de saída")
	}
	if needBatch {
		fs.IntVar(&c.workers, "workers", runtime.NumCPU(), "workers concorrentes")
		fs.BoolVar(&c.overwrite, "overwrite", false, "sobrescrever arquivos existentes")
	}
	fs.IntVar(&c.maxWidth, "maxwidth", 800, "largura máxima do preview (px)")
}

func (c *commonFlags) options() (watermark.Options, error) {
	pos, err := watermark.ParsePosition(c.pos)
	if err != nil {
		return watermark.Options{}, err
	}
	format, err := watermark.ParseFormat(c.format)
	if err != nil {
		return watermark.Options{}, err
	}
	opts := watermark.Options{
		Position: pos, SizePercent: c.size, MarginPercent: c.margin,
		Opacity: c.opacity, Quality: c.quality, Format: format,
	}
	return opts, opts.Validate()
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "scan":
		err = cmdScan(os.Args[2:])
	case "apply":
		err = cmdApply(os.Args[2:])
	case "batch":
		err = cmdBatch(os.Args[2:])
	case "preview":
		err = cmdPreview(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`watermark — marca d'água em lote (Fase 1)

  watermark scan --in PASTA
  watermark apply --in FOTO --logo LOGO --out SAIDA [opções]
  watermark batch --in PASTA --logo LOGO --out PASTA [opções]
  watermark preview --in FOTO --logo LOGO --out SAIDA [--maxwidth 800]

Opções visuais: --pos --size --margin --opacity --quality --format`)
}

func cmdScan(args []string) error {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	in := fs.String("in", "", "pasta de entrada")
	_ = fs.Parse(args)
	if *in == "" {
		return fmt.Errorf("informe --in PASTA")
	}
	res, err := files.ScanFolder(*in)
	if err != nil {
		return err
	}
	fmt.Printf("encontradas: %d | ignoradas: %d | inválidas: %d\n", len(res.Files), res.Ignored, res.Invalid)
	return nil
}

func loadLogoAndOpts(c *commonFlags) (image.Image, watermark.Options, error) {
	opts, err := c.options()
	if err != nil {
		return nil, opts, err
	}
	if c.logo == "" {
		return nil, opts, fmt.Errorf("informe --logo ARQUIVO")
	}
	logo, err := watermark.DecodeLogo(c.logo)
	if err != nil {
		return nil, opts, err
	}
	return logo, opts, nil
}

func cmdApply(args []string) error {
	fs := flag.NewFlagSet("apply", flag.ExitOnError)
	in := fs.String("in", "", "foto de entrada")
	var c commonFlags
	c.register(fs, true, true, false)
	_ = fs.Parse(args)
	if *in == "" || c.out == "" {
		return fmt.Errorf("informe --in FOTO, --logo LOGO e --out SAIDA")
	}
	logo, opts, err := loadLogoAndOpts(&c)
	if err != nil {
		return err
	}
	start := time.Now()
	if err := watermark.ProcessOne(*in, logo, c.out, opts); err != nil {
		return err
	}
	fmt.Printf("ok: %s -> %s (%s)\n", *in, c.out, time.Since(start).Round(time.Millisecond))
	return nil
}

func cmdPreview(args []string) error {
	fs := flag.NewFlagSet("preview", flag.ExitOnError)
	in := fs.String("in", "", "foto de entrada")
	var c commonFlags
	c.register(fs, true, true, false)
	_ = fs.Parse(args)
	if *in == "" || c.out == "" {
		return fmt.Errorf("informe --in FOTO, --logo LOGO e --out SAIDA")
	}
	logo, opts, err := loadLogoAndOpts(&c)
	if err != nil {
		return err
	}
	prev, err := watermark.Preview(*in, logo, opts, c.maxWidth)
	if err != nil {
		return err
	}
	return files.WriteAtomic(c.out, func(w io.Writer) error {
		return watermark.Encode(w, prev, opts)
	})
}

func cmdBatch(args []string) error {
	fs := flag.NewFlagSet("batch", flag.ExitOnError)
	in := fs.String("in", "", "pasta de entrada")
	var c commonFlags
	c.register(fs, true, true, true)
	_ = fs.Parse(args)
	if *in == "" || c.out == "" {
		return fmt.Errorf("informe --in PASTA, --logo LOGO e --out PASTA")
	}
	logo, opts, err := loadLogoAndOpts(&c)
	if err != nil {
		return err
	}
	res, err := files.ScanFolder(*in)
	if err != nil {
		return err
	}
	fmt.Printf("encontradas: %d | ignoradas: %d | inválidas: %d\n", len(res.Files), res.Ignored, res.Invalid)
	if len(res.Files) == 0 {
		return fmt.Errorf("nenhuma imagem compatível em %s", *in)
	}
	if len(res.Files) > batch.MaxBatch {
		return fmt.Errorf("%w (encontradas: %d)", batch.ErrTooMany, len(res.Files))
	}
	if err := files.EnsureOutDir(c.out, *in); err != nil {
		return err
	}
	absOut, _ := filepath.Abs(c.out)
	jobs := make([]batch.Job, 0, len(res.Files))
	for _, f := range res.Files {
		stem := strings.TrimSuffix(filepath.Base(f), filepath.Ext(f))
		outBase := stem + opts.Format.Ext()
		dst, rerr := files.ResolveOutput(absOut, outBase, c.overwrite)
		if rerr != nil {
			return rerr
		}
		jobs = append(jobs, batch.Job{Input: f, Output: dst})
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	fn := func(ctx context.Context, j batch.Job) batch.Result {
		start := time.Now()
		r := batch.Result{Input: j.Input, Output: j.Output}
		if ctx.Err() != nil {
			r.Err = ctx.Err()
			return r
		}
		r.Err = watermark.ProcessOne(j.Input, logo, j.Output, opts)
		r.Elapsed = time.Since(start)
		return r
	}
	rep, err := batch.Run(ctx, jobs, c.workers, fn, func(p batch.Progress) {
		fmt.Fprintf(os.Stderr, "\rprogresso: %d/%d ok, %d falhas", p.Done, p.Total, p.Failed)
	})
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return err
	}
	rep.OutDir = absOut
	printReport(rep)
	if rep.Failed > 0 {
		return fmt.Errorf("%d falha(s) — veja acima", rep.Failed)
	}
	return nil
}

func printReport(rep batch.Report) {
	fmt.Printf("concluídos: %d | falhas: %d | total: %d | tempo: %s | saída: %s\n",
		rep.Done, rep.Failed, rep.Total, rep.Elapsed.Round(time.Millisecond), rep.OutDir)
	for _, f := range rep.Failures {
		fmt.Printf("  FALHA %s: %v\n", f.Input, f.Err)
	}
}
