package daemon

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIsIgnoredFile(t *testing.T) {
	ignored := []string{
		".git",
		".nuvem",
		".goutputstream-12345",
		".bashrc",
		"file.tmp",
		"file.TMP",
		"file.swp",
		"file.swo",
		"file.bak",
		"file.crdownload",
		"file.part",
		"file~",
		"#file#",
		".~lock.documento.docx#",
		".~lock.planilha.xlsx#",
		"~$documento.docx",
	}

	for _, name := range ignored {
		if !isIgnoredFile(name) {
			t.Errorf("esperava que %q fosse ignorado, mas não foi", name)
		}
	}

	valid := []string{
		"documento.docx",
		"peticao.pdf",
		"planilha.xlsx",
		"arquivo.txt",
		"foto.jpg",
		"codigo.go",
	}

	for _, name := range valid {
		if isIgnoredFile(name) {
			t.Errorf("esperava que %q fosse aceito, mas foi ignorado", name)
		}
	}
}

func TestFolderWatcherDetectsChangesAndDebounces(t *testing.T) {
	dir := t.TempDir()

	fw, err := NewFolderWatcher(dir, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("erro ao criar watcher: %v", err)
	}
	defer fw.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go fw.Start(ctx)

	// 1. Escreve arquivo ignorado -> não deve disparar
	lockPath := filepath.Join(dir, ".~lock.test.docx#")
	if err := os.WriteFile(lockPath, []byte("lock"), 0o644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-fw.Events():
		t.Fatal("recebeu evento para arquivo de lock ignorado")
	case <-time.After(250 * time.Millisecond):
		// Sucesso: lock ignorado
	}

	// 2. Escreve arquivo válido -> deve disparar após debounce
	docPath := filepath.Join(dir, "minuta.docx")
	if err := os.WriteFile(docPath, []byte("conteudo inicial"), 0o644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-fw.Events():
		// Sucesso: detectou o arquivo válido
	case <-time.After(1 * time.Second):
		t.Fatal("tempo esgotado esperando evento de alteração em arquivo válido")
	}

	// 3. Testa supressão (durante escrita do motor de sincronização)
	fw.SetSuppressed(true)
	if err := os.WriteFile(docPath, []byte("conteudo alterado pelo sync"), 0o644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-fw.Events():
		t.Fatal("recebeu evento enquanto estava suprimido")
	case <-time.After(250 * time.Millisecond):
		// Sucesso: evento suprimido
	}

	fw.Drain()
	fw.SetSuppressed(false)

	// 4. Testa subdiretório recursivo
	subDir := filepath.Join(dir, "subpasta")
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	time.Sleep(150 * time.Millisecond) // Aguarda o watcher registrar a subpasta
	fw.Drain()

	subDoc := filepath.Join(subDir, "subdocumento.docx")
	if err := os.WriteFile(subDoc, []byte("texto da subpasta"), 0o644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-fw.Events():
		// Sucesso: detectou alteração dentro da subpasta
	case <-time.After(1 * time.Second):
		t.Fatal("não detectou alteração em subdiretório recursivo")
	}
}
