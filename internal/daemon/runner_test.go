package daemon

import (
	"context"
	"errors"
	"os"
	"path/filepath"
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

type countingEngine struct {
	syncCount int
	notifyCh  chan int
}

func (e *countingEngine) Sync(context.Context, config.SyncConfig) error {
	e.syncCount++
	if e.notifyCh != nil {
		e.notifyCh <- e.syncCount
	}
	return nil
}

func TestRunnerTriggersOnFileChange(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dir := t.TempDir()
	engine := &countingEngine{notifyCh: make(chan int, 10)}
	runner := New(engine, config.SyncConfig{LocalPath: dir}, time.Hour).
		WithDebounce(50 * time.Millisecond).
		WithSettle(10 * time.Millisecond)

	go func() { _ = runner.Run(ctx, func(error) {}) }()

	// Espera a sincronização inicial
	select {
	case count := <-engine.notifyCh:
		if count != 1 {
			t.Fatalf("esperava 1 chamada na inicialização, obteve %d", count)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("tempo esgotado aguardando sincronização inicial")
	}

	// Aguarda estabilização pós-sincronização inicial
	time.Sleep(50 * time.Millisecond)

	// Altera um arquivo na pasta monitorada
	filePath := dir + "/arquivo_teste.docx"
	if err := os.WriteFile(filePath, []byte("conteudo alterado"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Deve disparar a segunda sincronização via watcher
	select {
	case count := <-engine.notifyCh:
		if count != 2 {
			t.Fatalf("esperava 2 chamadas após alteração, obteve %d", count)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("tempo esgotado aguardando sincronização disparada pelo arquivo")
	}
}

func TestRunnerTimerResetsOnFileChange(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dir := t.TempDir()
	engine := &countingEngine{notifyCh: make(chan int, 10)}
	// Intervalo de 300ms, debounce de 40ms, settle de 10ms
	runner := New(engine, config.SyncConfig{LocalPath: dir}, 300*time.Millisecond).
		WithDebounce(40 * time.Millisecond).
		WithSettle(10 * time.Millisecond)

	go func() { _ = runner.Run(ctx, func(error) {}) }()

	// 1. Sincronização inicial
	select {
	case count := <-engine.notifyCh:
		if count != 1 {
			t.Fatalf("esperava 1 chamada na inicialização, obteve %d", count)
		}
	case <-time.After(time.Second):
		t.Fatal("tempo esgotado aguardando inicialização")
	}

	// Aguarda 180ms (mais da metade do intervalo de 300ms)
	time.Sleep(180 * time.Millisecond)

	// Modifica arquivo
	filePath := filepath.Join(dir, "minuta.docx")
	if err := os.WriteFile(filePath, []byte("versão 1"), 0o644); err != nil {
		t.Fatal(err)
	}

	// 2. Deve disparar sincronização por arquivo (count == 2)
	select {
	case count := <-engine.notifyCh:
		if count != 2 {
			t.Fatalf("esperava sincronização por arquivo (2), obteve %d", count)
		}
	case <-time.After(time.Second):
		t.Fatal("tempo esgotado aguardando sincronização por arquivo")
	}

	// Se o timer não tivesse sido reiniciado, ele dispararia aos 300ms (em ~70ms).
	// Como o timer FOI reiniciado para novos 300ms, em 150ms NENHUM novo disparo deve ocorrer.
	select {
	case count := <-engine.notifyCh:
		t.Fatalf("o timer disparou prematuramente após a alteração do arquivo: chamada %d", count)
	case <-time.After(150 * time.Millisecond):
		// Sucesso: timer foi rearmado e não disparou prematuramente!
	}
}
