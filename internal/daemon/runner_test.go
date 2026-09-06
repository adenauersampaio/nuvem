package daemon

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/adenauersampaio/nuvem/internal/config"
)

type failingEngine struct{ calls int }

func (e *failingEngine) Sync(context.Context, config.SyncConfig) error {
	e.calls++
	return errors.New("offline")
}

func TestRunnerReportsFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	engine := &failingEngine{}
	reported := make(chan error, 1)
	runner := New(engine, config.SyncConfig{}, time.Hour)

	go func() { _ = runner.Run(ctx, func(err error) { reported <- err }) }()

	select {
	case err := <-reported:
		if err == nil || engine.calls != 1 {
			t.Fatalf("report=%v calls=%d", err, engine.calls)
		}
	case <-time.After(time.Second):
		t.Fatal("a falha não foi reportada")
	}
}
