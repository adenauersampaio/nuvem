package daemon

import (
	"context"
	"os"
	"time"

	"github.com/adenauersampaio/nuvem/internal/config"
)

type Engine interface {
	Sync(context.Context, config.SyncConfig) error
}

// Runner coordena a sincronização contínua por intervalo de tempo e por detecção
// de eventos de modificação de arquivos em tempo real, garantindo execução única
// e evitando realimentação/eco infinito.
type Runner struct {
	engine   Engine
	sync     config.SyncConfig
	interval time.Duration
	debounce time.Duration
	settle   time.Duration
}

func New(engine Engine, sync config.SyncConfig, interval time.Duration) Runner {
	return Runner{
		engine:   engine,
		sync:     sync,
		interval: interval,
		debounce: defaultDebounceDuration,
		settle:   500 * time.Millisecond,
	}
}

// WithDebounce permite definir a janela de debounce (útil para testes).
func (r Runner) WithDebounce(d time.Duration) Runner {
	r.debounce = d
	return r
}

// WithSettle permite definir o tempo de estabilização pós-sincronização (útil para testes).
func (r Runner) WithSettle(d time.Duration) Runner {
	r.settle = d
	return r
}

func (r Runner) Run(ctx context.Context, report func(error)) error {
	var watcher *FolderWatcher
	var watcherCh <-chan struct{}

	if r.sync.LocalPath != "" {
		if fi, err := os.Stat(r.sync.LocalPath); err == nil && fi.IsDir() {
			fw, err := NewFolderWatcher(r.sync.LocalPath, r.debounce)
			if err == nil {
				watcher = fw
				watcherCh = fw.Events()
				defer watcher.Close()
				go watcher.Start(ctx)
			}
		}
	}

	// Execução inicial ao iniciar o daemon
	r.syncCycle(ctx, watcher, report)

	timer := time.NewTimer(r.interval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-timer.C:
			r.syncCycle(ctx, watcher, report)
			timer.Reset(r.interval)

		case <-watcherCh:
			r.syncCycle(ctx, watcher, report)
			// Rearma o temporizador do zero para evitar que uma sincronização por
			// tempo ocorra imediatamente após a sincronização por arquivo.
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(r.interval)
		}
	}
}

func (r Runner) syncCycle(ctx context.Context, watcher *FolderWatcher, report func(error)) {
	if watcher != nil {
		watcher.SetSuppressed(true)
	}

	if err := r.engine.Sync(ctx, r.sync); err != nil {
		report(err)
	}

	if watcher != nil {
		// Intervalo de estabilização para descartar eventos gerados pelo próprio
		// download/escrita de arquivos da sincronização
		settle := r.settle
		if settle <= 0 {
			settle = 500 * time.Millisecond
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(settle):
		}
		watcher.Drain()
		watcher.SetSuppressed(false)
	}
}
