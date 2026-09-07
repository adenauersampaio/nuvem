package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/adenauersampaio/nuvem/internal/app"
	"github.com/adenauersampaio/nuvem/internal/desktop"
	"github.com/adenauersampaio/nuvem/internal/gui"
	"github.com/adenauersampaio/nuvem/internal/i18n"
)

type desktopController struct{ language i18n.Language }

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--install" {
		executable, err := os.Executable()
		if err == nil {
			err = desktop.Install(executable)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if len(os.Args) == 2 && os.Args[1] == "--uninstall" {
		if err := desktop.Uninstall(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	language := i18n.Detect(os.Getenv("NUVEM_LANG"))
	if os.Getenv("NUVEM_LANG") == "" {
		language = i18n.Detect(os.Getenv("LANG"))
	}
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--lang" && i+1 < len(os.Args) {
			if l, ok := i18n.Parse(os.Args[i+1]); ok {
				language = l
			}
		}
	}
	gui.SetDonationURLs(
		os.Getenv("NUVEM_BMC_URL"),
		os.Getenv("NUVEM_LIVEPIX_URL"),
		os.Getenv("NUVEM_BINANCE_URL"),
		os.Getenv("NUVEM_GITHUB_SPONSORS_URL"),
	)
	gui.Run(desktopController{language: language}, language == i18n.PortugueseBrazil)
}

func (c desktopController) Activated() { c.refresh() }

func (c desktopController) SyncNow() {
	gui.SetStatus(c.text("Synchronizing…", "Sincronizando…"), false)
	if err := app.RunOnce(context.Background()); err != nil {
		gui.SetStatus(fmt.Sprintf(c.text("Synchronization needs attention: %v", "A sincronização precisa de atenção: %v"), err), true)
		c.refresh()
		return
	}
	gui.SetStatus(c.text("Everything is up to date.", "Tudo está atualizado."), false)
	c.refresh()
}

func (c desktopController) Save(local, remote string, minutes int, clientID, clientSecret string) {
	if err := app.Save(strings.TrimSpace(local), strings.TrimSpace(remote), time.Duration(minutes)*time.Minute, strings.TrimSpace(clientID), strings.TrimSpace(clientSecret)); err != nil {
		gui.SetStatus(fmt.Sprintf(c.text("Could not save the setup: %v", "Não foi possível salvar a configuração: %v"), err), true)
		return
	}
	gui.SetStatus(c.text("Configuration saved. The background service will use it on its next check.", "Configuração salva. O serviço em segundo plano a usará na próxima verificação."), false)
	c.refresh()
}

func (c desktopController) ConnectGoogleDrive(clientID, clientSecret string) {
	gui.SetStatus(c.text("A browser will open so you can connect Google Drive and choose one folder.", "O navegador será aberto para conectar o Google Drive e escolher uma pasta."), false)
	if err := app.ConnectGoogleDrive(context.Background(), strings.TrimSpace(clientID), strings.TrimSpace(clientSecret)); err != nil {
		gui.SetStatus(fmt.Sprintf(c.text("Could not connect Google Drive: %v", "Não foi possível conectar o Google Drive: %v"), err), true)
		c.refresh()
		return
	}
	if err := app.RestartService(context.Background()); err != nil {
		gui.SetStatus(fmt.Sprintf(c.text("Google Drive is connected, but the background service needs a restart: %v", "O Google Drive está conectado, mas o serviço em segundo plano precisa ser reiniciado: %v"), err), true)
		c.refresh()
		return
	}
	gui.SetStatus(c.text("Google Drive is connected to the folder you selected.", "O Google Drive está conectado à pasta que você escolheu."), false)
	c.refresh()
}

func (c desktopController) refresh() {
	state, err := app.DashboardState(context.Background())
	if err != nil {
		gui.SetStatus(fmt.Sprintf(c.text("Setup needs attention: %v", "A configuração precisa de atenção: %v"), err), true)
		gui.SetDashboard(gui.Dashboard{Service: c.text("Status unavailable", "Estado indisponível")})
		return
	}
	if !state.Configured {
		gui.SetStatus(c.text("Set up a folder to begin continuous synchronization.", "Configure uma pasta para iniciar a sincronização contínua."), false)
		gui.SetDashboard(gui.Dashboard{Service: c.text("Not configured", "Ainda não configurado")})
		return
	}
	service := c.serviceText(state.Service)
	gui.SetDashboard(gui.Dashboard{Local: state.LocalPath, Remote: state.Remote, Minutes: int(state.Interval.Minutes()), Service: service, Configured: true, GoogleClientID: state.GoogleClientID, GoogleClientSecret: state.GoogleClientSecret})
	if state.Service == "active" {
		gui.SetStatus(c.text("Continuous synchronization is active.", "A sincronização contínua está ativa."), false)
	} else {
		gui.SetStatus(c.text("The background service is not active. You can still synchronize now.", "O serviço em segundo plano não está ativo. Ainda é possível sincronizar agora."), true)
	}
}

func (c desktopController) serviceText(state string) string {
	if c.language == i18n.PortugueseBrazil {
		switch state {
		case "active":
			return "Serviço em segundo plano: ativo"
		case "inactive":
			return "Serviço em segundo plano: inativo"
		default:
			return "Serviço em segundo plano: estado indisponível"
		}
	}
	switch state {
	case "active":
		return "Background service: active"
	case "inactive":
		return "Background service: inactive"
	default:
		return "Background service: status unavailable"
	}
}

func (c desktopController) text(english, portuguese string) string {
	if c.language == i18n.PortugueseBrazil {
		return portuguese
	}
	return english
}
