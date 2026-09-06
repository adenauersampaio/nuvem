package daemon

import (
	"context"
	"time"

	"github.com/adenauersampaio/nuvem/internal/config"
)

type Engine interface {
	Sync(context.Context, config.SyncConfig) error
}

// Runner executa uma sincronização por vez. Se uma execução demorar mais que o
// intervalo, o próximo ciclo espera a anterior terminar em vez de sobrepor jobs.
type Runner struct {
	engine   Engine
	sync     config.SyncConfig
	interval time.Duration
}

func New(engine Engine, sync config.SyncConfig, interval time.Duration) Runner {
	return Runner{engine: engine, sync: sync, interval: interval}
}

func (r Runner) Run(ctx context.Context, report func(error)) error {
	for {
		if err := r.engine.Sync(ctx, r.sync); err != nil {
			report(err)
		}

		timer := time.NewTimer(r.interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}
