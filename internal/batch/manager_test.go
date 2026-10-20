package batch

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func okFn(_ context.Context, j Job) Result {
	time.Sleep(time.Millisecond)
	return Result{Input: j.Input, Output: j.Output}
}

func jobs(n int) []Job {
	js := make([]Job, n)
	for i := range js {
		js[i] = Job{Input: "in", Output: "out"}
	}
	return js
}

func TestRunAllSuccess(t *testing.T) {
	rep, err := Run(context.Background(), jobs(20), 4, okFn, nil)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Done != 20 || rep.Failed != 0 || rep.Total != 20 {
		t.Errorf("relatório inesperado: %+v", rep)
	}
}

func TestRunErrorIsolation(t *testing.T) {
	fn := func(_ context.Context, j Job) Result {
		if j.Input == "ruim" {
			return Result{Input: j.Input, Err: errors.New("corrompida")}
		}
		return Result{Input: j.Input}
	}
	js := []Job{{Input: "boa1"}, {Input: "ruim"}, {Input: "boa2"}}
	var progress []Progress
	rep, err := Run(context.Background(), js, 2, fn, func(p Progress) { progress = append(progress, p) })
	if err != nil {
		t.Fatal(err)
	}
	if rep.Done != 2 || rep.Failed != 1 || len(rep.Failures) != 1 {
		t.Errorf("falha deveria ser isolada: %+v", rep)
	}
	if len(progress) != 3 || progress[2].Done+progress[2].Failed != 3 {
		t.Errorf("progresso geral incorreto: %+v", progress)
	}
}

func TestRunTooMany(t *testing.T) {
	if _, err := Run(context.Background(), jobs(MaxBatch+1), 4, okFn, nil); !errors.Is(err, ErrTooMany) {
		t.Errorf("esperado ErrTooMany, obtido %v", err)
	}
}

func TestRunCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var started atomic.Int32
	fn := func(ctx context.Context, j Job) Result {
		started.Add(1)
		select {
		case <-ctx.Done():
			return Result{Input: j.Input, Err: ctx.Err()}
		case <-time.After(50 * time.Millisecond):
			return Result{Input: j.Input}
		}
	}
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()
	rep, err := Run(ctx, jobs(50), 2, fn, nil)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Done+rep.Failed >= 50 {
		t.Errorf("cancelamento deveria interromper o lote: %+v", rep)
	}
}

func TestRunPanicIsolated(t *testing.T) {
	fn := func(_ context.Context, j Job) Result { panic("bug") }
	rep, err := Run(context.Background(), jobs(3), 2, fn, nil)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Failed != 3 {
		t.Errorf("pânico deveria virar falha isolada: %+v", rep)
	}
}
