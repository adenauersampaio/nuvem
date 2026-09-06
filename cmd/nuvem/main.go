package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/adenauersampaio/nuvem/internal/config"
	"github.com/adenauersampaio/nuvem/internal/daemon"
	"github.com/adenauersampaio/nuvem/internal/i18n"
	"github.com/adenauersampaio/nuvem/internal/rcloneconfig"
	"github.com/adenauersampaio/nuvem/internal/runlock"
	"github.com/adenauersampaio/nuvem/internal/service"
	"github.com/adenauersampaio/nuvem/internal/syncer"
)

const version = "0.1.0-beta.1"

func main() {
	language, args, err := languageFromArgs(os.Args[1:])
	if err == nil {
		err = run(args, language, os.Stdout, os.Stderr)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, i18n.Text(language, i18n.KeyError), err)
		os.Exit(1)
	}
}

func languageFromArgs(args []string) (i18n.Language, []string, error) {
	language := i18n.Detect(os.Getenv("NUVEM_LANG"))
	if os.Getenv("NUVEM_LANG") == "" {
		language = i18n.Detect(os.Getenv("LANG"))
	}
	if len(args) < 2 || args[0] != "--lang" {
		return language, args, nil
	}
	selected, ok := i18n.Parse(args[1])
	if !ok {
		return language, nil, fmt.Errorf(i18n.Text(language, i18n.KeyInvalidLanguage), args[1])
	}
	return selected, args[2:], nil
}

func run(args []string, language i18n.Language, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		printHelp(language, stdout)
		return nil
	}

	switch args[0] {
	case "help", "--help", "-h":
		printHelp(language, stdout)
		return nil
	case "version", "--version":
		_, err := fmt.Fprintln(stdout, "nuvem", version)
		return err
	case "doctor":
		return doctor(language, stdout)
	case "init":
		return initialize(args[1:], language, stdout)
	case "run-once":
		return runOnce(language)
	case "daemon":
		return runDaemon(language)
	case "install":
		return install(args[1:], language, stdout)
	case "uninstall":
		return uninstall(language, stdout)
	case "status":
		return status(language, stdout)
	case "import-rclone":
		return importRclone(args[1:], language, stdout)
	default:
		return fmt.Errorf(i18n.Text(language, i18n.KeyUnknownCommand), args[0])
	}
}

func printHelp(language i18n.Language, w io.Writer) {
	fmt.Fprint(w, i18n.Text(language, i18n.KeyHelp))
}

func initialize(args []string, language i18n.Language, stdout io.Writer) error {
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	local := flags.String("local", "", "pasta local")
	remote := flags.String("remote", "", "pasta remota")
	interval := flags.Duration("interval", 1*time.Minute, "intervalo de verificação")
	if err := flags.Parse(args); err != nil {
		return err
	}
	cfg := config.Config{
		Version: config.CurrentVersion,
		Sync: config.SyncConfig{
			LocalPath: *local,
			Remote:    *remote,
			Interval:  *interval,
		},
	}
	if err := config.SaveDefault(cfg); err != nil {
		return err
	}
	path, err := config.DefaultPath()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, i18n.Text(language, i18n.KeyConfigCreated), path)
	return err
}

func doctor(language i18n.Language, w io.Writer) error {
	cfg, err := config.LoadDefault()
	if errors.Is(err, os.ErrNotExist) {
		_, err = fmt.Fprintln(w, i18n.Text(language, i18n.KeyConfigMissing))
		return err
	}
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	engine, err := embeddedEngine()
	if err != nil {
		return err
	}
	if err := engine.Check(); err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, i18n.Text(language, i18n.KeyConfigValid), cfg.Sync.LocalPath, cfg.Sync.Remote)
	return err
}

func runOnce(language i18n.Language) error {
	cfg, err := config.LoadDefault()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	engine, err := embeddedEngine()
	if err != nil {
		return err
	}
	return withSyncLock(language, func() error {
		return engine.Sync(context.Background(), cfg.Sync)
	})
}

func runDaemon(language i18n.Language) error {
	cfg, err := config.LoadDefault()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	engine, err := embeddedEngine()
	if err != nil {
		return err
	}
	err = withSyncLock(language, func() error {
		runner := daemon.New(engine, cfg.Sync, cfg.Sync.Interval)
		return runner.Run(ctx, func(err error) {
			if !errors.Is(err, context.Canceled) {
				fmt.Fprintln(os.Stderr, i18n.Text(language, i18n.KeySyncFailed), err)
			}
		})
	})
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

func withSyncLock(language i18n.Language, action func() error) error {
	lock, err := runlock.AcquireDefault()
	if errors.Is(err, runlock.ErrHeld) {
		return errors.New(i18n.Text(language, i18n.KeySyncInProgress))
	}
	if err != nil {
		return err
	}
	defer func() {
		_ = lock.Release()
	}()
	return action()
}

func install(args []string, language i18n.Language, stdout io.Writer) error {
	flags := flag.NewFlagSet("install", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	binary, err := os.Executable()
	if err != nil {
		return err
	}
	flags.StringVar(&binary, "binary", binary, "caminho do executável")
	if err := flags.Parse(args); err != nil {
		return err
	}
	manager, err := serviceManager()
	if err != nil {
		return err
	}
	if err := manager.Install(context.Background(), binary); err != nil {
		return err
	}
	_, err = fmt.Fprint(stdout, i18n.Text(language, i18n.KeyServiceInstalled))
	return err
}

func uninstall(language i18n.Language, stdout io.Writer) error {
	manager, err := serviceManager()
	if err != nil {
		return err
	}
	if err := manager.Uninstall(context.Background()); err != nil {
		return err
	}
	_, err = fmt.Fprint(stdout, i18n.Text(language, i18n.KeyServiceRemoved))
	return err
}

func status(language i18n.Language, stdout io.Writer) error {
	manager, err := serviceManager()
	if err != nil {
		return err
	}
	state, err := manager.Status(context.Background())
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, i18n.Text(language, i18n.KeyServiceStatus), state)
	return err
}

func serviceManager() (service.Manager, error) {
	path, err := service.DefaultUnitPath()
	if err != nil {
		return service.Manager{}, err
	}
	return service.New(path, service.OSRunner{}), nil
}

func embeddedEngine() (syncer.EmbeddedEngine, error) {
	path, err := config.DefaultRcloneConfigPath()
	if err != nil {
		return syncer.EmbeddedEngine{}, err
	}
	return syncer.NewEmbeddedEngine(path), nil
}

func importRclone(args []string, language i18n.Language, stdout io.Writer) error {
	flags := flag.NewFlagSet("import-rclone", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	remote := flags.String("remote", "GoogleDrive", "nome do perfil do Drive")
	userConfig, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	source := flags.String("source", filepath.Join(userConfig, "rclone", "rclone.conf"), "configuração de origem")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if strings.Contains(*remote, ":") {
		return errors.New("use apenas o nome do perfil, sem ':'")
	}
	destination, err := config.DefaultRcloneConfigPath()
	if err != nil {
		return err
	}
	if err := rcloneconfig.ImportRemote(*source, destination, *remote); err != nil {
		return err
	}
	_, err = fmt.Fprint(stdout, i18n.Text(language, i18n.KeyRemoteImported))
	return err
}
