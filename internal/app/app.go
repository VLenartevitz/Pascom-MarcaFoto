package app

import (
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"watermark-studio/internal/batch"
	"watermark-studio/internal/files"
	"watermark-studio/internal/watermark"
)

// App struct
type App struct {
	ctx          context.Context
	cancelBatch  context.CancelFunc
	logoPath     string
	logoImage    image.Image
	outputDir    string
	processing   bool
}

// New creates a new App application struct
func New() *App {
	return &App{}
}

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// DomReady is called after the frontend is ready
func (a *App) DomReady(ctx context.Context) {
	// Inicializa com valores padrão
	runtime.EventsEmit(ctx, "defaults", DefaultConfig())
}

// BeforeClose is called when the user clicks the close button
func (a *App) BeforeClose(ctx context.Context) (prevent bool) {
	if a.processing {
		return true // Impede fechar durante processamento
	}
	return false
}

// Shutdown is called when the app is shutting down
func (a *App) Shutdown(ctx context.Context) {
	if a.cancelBatch != nil {
		a.cancelBatch()
	}
}

// ===== Configuração =====

// Config representa a configuração visual da marca d'água
type Config struct {
	Position      string  `json:"position"`
	SizePercent   float64 `json:"sizePercent"`
	MarginPercent float64 `json:"marginPercent"`
	Opacity       float64 `json:"opacity"`
	Quality       int     `json:"quality"`
	Format        string  `json:"format"`
	Workers       int     `json:"workers"`
	Overwrite     bool    `json:"overwrite"`
}

// DefaultConfig retorna a configuração padrão
func DefaultConfig() Config {
	return Config{
		Position:      "bottom-right",
		SizePercent:   20,
		MarginPercent: 3,
		Opacity:       0.7,
		Quality:       92,
		Format:        "jpeg",
		Workers:       runtime.NumCPU(),
		Overwrite:     false,
	}
}

// ValidateConfig valida a configuração
func (a *App) ValidateConfig(cfg Config) error {
	pos, err := watermark.ParsePosition(cfg.Position)
	if err != nil {
		return err
	}
	format, err := watermark.ParseFormat(cfg.Format)
	if err != nil {
		return err
	}
	opts := watermark.Options{
		Position:      pos,
		SizePercent:   cfg.SizePercent,
		MarginPercent: cfg.MarginPercent,
		Opacity:       cfg.Opacity,
		Quality:       cfg.Quality,
		Format:        format,
	}
	return opts.Validate()
}

// ===== Seleção de arquivos =====

// ScanResult representa o resultado do escaneamento
type ScanResult struct {
	Files   []string `json:"files"`
	Count   int      `json:"count"`
	Ignored int      `json:"ignored"`
	Invalid int      `json:"invalid"`
	Error   string   `json:"error,omitempty"`
}

// ScanFolder escaneia uma pasta em busca de imagens compatíveis
func (a *App) ScanFolder(path string) ScanResult {
	res, err := files.ScanFolder(path)
	if err != nil {
		return ScanResult{Error: err.Error()}
	}
	return ScanResult{
		Files:   res.Files,
		Count:   len(res.Files),
		Ignored: res.Ignored,
		Invalid: res.Invalid,
	}
}

// SelectInputFolder abre diálogo para selecionar pasta de entrada
func (a *App) SelectInputFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Selecionar pasta de fotos",
	})
}

// SelectLogoFile abre diálogo para selecionar arquivo da logo
func (a *App) SelectLogoFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Selecionar logo (PNG com transparência)",
		Filters: []runtime.FileFilter{
			{DisplayName: "Imagens", Pattern: "*.png;*.jpg;*.jpeg;*.webp"},
			{DisplayName: "Todos os arquivos", Pattern: "*.*"},
		},
	})
}

// SelectOutputFolder abre diálogo para selecionar pasta de saída
func (a *App) SelectOutputFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Selecionar pasta de saída",
	})
}

// SetLogo carrega a logo do caminho especificado
func (a *App) SetLogo(path string) error {
	if path == "" {
		a.logoPath = ""
		a.logoImage = nil
		return nil
	}
	img, err := watermark.DecodeLogo(path)
	if err != nil {
		return err
	}
	a.logoPath = path
	a.logoImage = img
	return nil
}

// GetLogoPath retorna o caminho da logo atual
func (a *App) GetLogoPath() string {
	return a.logoPath
}

// SetOutputDir define a pasta de saída
func (a *App) SetOutputDir(path string) error {
	if path == "" {
		a.outputDir = ""
		return nil
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	a.outputDir = abs
	return nil
}

// GetOutputDir retorna a pasta de saída atual
func (a *App) GetOutputDir() string {
	return a.outputDir
}

// ===== Preview =====

// PreviewResult representa o resultado da pré-visualização
type PreviewResult struct {
	Base64 string `json:"base64"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Error  string `json:"error,omitempty"`
}

// GeneratePreview gera uma pré-visualização da marca d'água
func (a *App) GeneratePreview(photoPath string, cfg Config, maxWidth int) PreviewResult {
	if a.logoImage == nil {
		return PreviewResult{Error: "Nenhuma logo carregada"}
	}
	if photoPath == "" {
		return PreviewResult{Error: "Nenhuma foto selecionada"}
	}

	if err := a.ValidateConfig(cfg); err != nil {
		return PreviewResult{Error: err.Error()}
	}

	pos, _ := watermark.ParsePosition(cfg.Position)
	format, _ := watermark.ParseFormat(cfg.Format)
	opts := watermark.Options{
		Position:      pos,
		SizePercent:   cfg.SizePercent,
		MarginPercent: cfg.MarginPercent,
		Opacity:       cfg.Opacity,
		Quality:       cfg.Quality,
		Format:        format,
	}

	previewImg, err := watermark.Preview(photoPath, a.logoImage, opts, maxWidth)
	if err != nil {
		return PreviewResult{Error: err.Error()}
	}

	// Converter para base64
	var buf strings.Builder
	writer := base64.NewEncoder(base64.StdEncoding, &buf)
	err = watermark.Encode(writer, previewImg, opts)
	writer.Close()
	if err != nil {
		return PreviewResult{Error: "Erro ao codificar preview: " + err.Error()}
	}

	bounds := previewImg.Bounds()
	return PreviewResult{
		Base64: buf.String(),
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
	}
}

// ===== Processamento em Lote =====

// BatchProgress representa o progresso do lote
type BatchProgress struct {
	Total   int     `json:"total"`
	Done    int     `json:"done"`
	Failed  int     `json:"failed"`
	Elapsed float64 `json:"elapsed"` // segundos
	Current string  `json:"current"` // arquivo atual
}

// BatchReport representa o relatório final
type BatchReport struct {
	Total    int           `json:"total"`
	Done     int           `json:"done"`
	Failed   int           `json:"failed"`
	Elapsed  float64       `json:"elapsed"`
	OutDir   string        `json:"outDir"`
	Failures []BatchFailure `json:"failures"`
}

type BatchFailure struct {
	Input string `json:"input"`
	Error string `json:"error"`
}

// StartBatch inicia o processamento em lote
func (a *App) StartBatch(inputDir string, cfg Config) (BatchReport, error) {
	if a.processing {
		return BatchReport{}, fmt.Errorf("processamento já em andamento")
	}
	if a.logoImage == nil {
		return BatchReport{}, fmt.Errorf("nenhuma logo carregada")
	}
	if a.outputDir == "" {
		return BatchReport{}, fmt.Errorf("pasta de saída não definida")
	}
	if inputDir == "" {
		return BatchReport{}, fmt.Errorf("pasta de entrada não definida")
	}

	if err := a.ValidateConfig(cfg); err != nil {
		return BatchReport{}, err
	}

	// Escanear arquivos
	res, err := files.ScanFolder(inputDir)
	if err != nil {
		return BatchReport{}, err
	}
	if len(res.Files) == 0 {
		return BatchReport{}, fmt.Errorf("nenhuma imagem compatível encontrada")
	}
	if len(res.Files) > batch.MaxBatch {
		return BatchReport{}, fmt.Errorf("lote excede o máximo de %d imagens (encontradas: %d)", batch.MaxBatch, len(res.Files))
	}

	// Validar pasta de saída
	if err := files.EnsureOutDir(a.outputDir, inputDir); err != nil {
		return BatchReport{}, err
	}
	absOut, _ := filepath.Abs(a.outputDir)

	// Preparar opções
	pos, _ := watermark.ParsePosition(cfg.Position)
	format, _ := watermark.ParseFormat(cfg.Format)
	opts := watermark.Options{
		Position:      pos,
		SizePercent:   cfg.SizePercent,
		MarginPercent: cfg.MarginPercent,
		Opacity:       cfg.Opacity,
		Quality:       cfg.Quality,
		Format:        format,
	}

	// Criar jobs
	jobs := make([]batch.Job, 0, len(res.Files))
	for _, f := range res.Files {
		stem := strings.TrimSuffix(filepath.Base(f), filepath.Ext(f))
		outBase := stem + opts.Format.Ext()
		dst, rerr := files.ResolveOutput(absOut, outBase, cfg.Overwrite)
		if rerr != nil {
			return BatchReport{}, rerr
		}
		jobs = append(jobs, batch.Job{Input: f, Output: dst})
	}

	// Context com cancelamento
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelBatch = cancel
	a.processing = true

	// Função de processamento por job
	fn := func(ctx context.Context, j batch.Job) batch.Result {
		start := time.Now()
		r := batch.Result{Input: j.Input, Output: j.Output}
		if ctx.Err() != nil {
			r.Err = ctx.Err()
			return r
		}
		// Emitir arquivo atual
		runtime.EventsEmit(a.ctx, "batch:current", filepath.Base(j.Input))
		r.Err = watermark.ProcessOne(j.Input, a.logoImage, j.Output, opts)
		r.Elapsed = time.Since(start)
		return r
	}

	// Callback de progresso
	progressCallback := func(p batch.Progress) {
		runtime.EventsEmit(a.ctx, "batch:progress", BatchProgress{
			Total:   p.Total,
			Done:    p.Done,
			Failed:  p.Failed,
			Elapsed: time.Since(time.Now().Add(-p.Elapsed)).Seconds(),
		})
	}

	// Executar em goroutine para não bloquear
	go func() {
		defer func() {
			a.processing = false
			a.cancelBatch = nil
		}()

		rep, err := batch.Run(ctx, jobs, cfg.Workers, fn, progressCallback)
		
		// Converter relatório
		failures := make([]BatchFailure, len(rep.Failures))
		for i, f := range rep.Failures {
			failures[i] = BatchFailure{Input: f.Input, Error: f.Err.Error()}
		}
		
		report := BatchReport{
			Total:    rep.Total,
			Done:     rep.Done,
			Failed:   rep.Failed,
			Elapsed:  rep.Elapsed.Seconds(),
			OutDir:   rep.OutDir,
			Failures: failures,
		}
		
		if err != nil {
			runtime.EventsEmit(a.ctx, "batch:error", err.Error())
		}
		runtime.EventsEmit(a.ctx, "batch:complete", report)
	}()

	return BatchReport{}, nil // Retorno imediato, resultado vem por evento
}

// CancelBatch cancela o processamento em andamento
func (a *App) CancelBatch() {
	if a.cancelBatch != nil {
		a.cancelBatch()
	}
}

// IsProcessing retorna se há processamento em andamento
func (a *App) IsProcessing() bool {
	return a.processing
}

// ===== Utilitários =====

// OpenFolder abre a pasta no explorador de arquivos
func (a *App) OpenFolder(path string) error {
	if path == "" {
		return fmt.Errorf("caminho vazio")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	// No Windows, usar explorer
	return runtime.BrowserOpenURL(a.ctx, "file:///"+abs)
}

// GetVersion retorna a versão do aplicativo
func (a *App) GetVersion() string {
	return "1.0.0 (Fase 2)"
}

// ReadFileAsBase64 lê um arquivo e retorna base64 (para preview de logo)
func (a *App) ReadFileAsBase64(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	
	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}