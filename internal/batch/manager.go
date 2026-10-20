package batch

import (
	"context"
	"errors"
	"sync"
	"time"
)

// MaxBatch é o limite de imagens por execução (RF-05).
const MaxBatch = 500

// ErrTooMany é retornado quando o lote excede MaxBatch.
var ErrTooMany = errors.New("lote excede o máximo de 500 imagens por execução")

// Job é uma unidade de trabalho: um arquivo de entrada e seu destino.
type Job struct {
	Input  string
	Output string
}

// Result é o desfecho de um Job. Err == nil indica sucesso.
type Result struct {
	Input   string
	Output  string
	Err     error
	Elapsed time.Duration
}

// Progress é o estado agregado para exibição (individual + geral, RF-05).
type Progress struct {
	Total   int
	Done    int
	Failed  int
	Elapsed time.Duration
}

// Report é o relatório final da execução (RF-07).
type Report struct {
	Total    int
	Done     int
	Failed   int
	Elapsed  time.Duration
	OutDir   string
	Failures []Result
}

// Run executa os jobs com concorrência limitada (workers) e memória
// controlada: um arquivo por vez por worker, sem carregar o lote
// inteiro (RNF-02). Erros por arquivo não interrompem o lote (RNF-05).
// O ctx cancela de forma controlada: novas tarefas param e o relatório
// reflete o que foi concluído (RNF-06). Se o lote foi cancelado,
// Done+Failed pode ser menor que Total.
func Run(ctx context.Context, jobs []Job, workers int, fn func(context.Context, Job) Result, onProgress func(Progress)) (Report, error) {
	if len(jobs) > MaxBatch {
		return Report{}, ErrTooMany
	}
	if workers < 1 {
		workers = 1
	}
	start := time.Now()
	rep := Report{Total: len(jobs)}

	jobsCh := make(chan Job)
	resCh := make(chan Result)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobsCh {
				// Isola a tarefa: fn nunca deve estourar o worker;
				// pânico vira falha do arquivo, não do lote (SDD §9).
				resCh <- safeCall(ctx, j, fn)
			}
		}()
	}
	go func() {
		wg.Wait()
		close(resCh)
	}()
	go func() {
		defer close(jobsCh)
		for _, j := range jobs {
			select {
			case <-ctx.Done():
				return
			case jobsCh <- j:
			}
		}
	}()

	for r := range resCh {
		r.Elapsed = time.Since(start)
		if r.Err != nil {
			rep.Failed++
			rep.Failures = append(rep.Failures, r)
		} else {
			rep.Done++
		}
		if onProgress != nil {
			onProgress(Progress{Total: rep.Total, Done: rep.Done, Failed: rep.Failed, Elapsed: time.Since(start)})
		}
	}
	rep.Elapsed = time.Since(start)
	return rep, nil
}

func safeCall(ctx context.Context, j Job, fn func(context.Context, Job) Result) (r Result) {
	defer func() {
		if rec := recover(); rec != nil {
			r = Result{Input: j.Input, Output: j.Output, Err: errors.New("falha interna isolada no worker")}
		}
	}()
	if ctx.Err() != nil {
		return Result{Input: j.Input, Output: j.Output, Err: ctx.Err()}
	}
	return fn(ctx, j)
}
