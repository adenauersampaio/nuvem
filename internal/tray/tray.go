// Package tray manages the Linux system tray icon and menu for Nuvem.
package tray

import (
	_ "embed"
	"sync"

	"fyne.io/systray"
)

//go:embed assets/icon.png
var defaultIcon []byte

type Actions struct {
	OnOpen          func()
	OnSyncNow       func()
	OnToggleService func()
	OnQuit          func()
}

type TrayManager struct {
	mu             sync.Mutex
	actions        Actions
	isPortuguese   bool
	serviceActive  bool
	mOpen          *systray.MenuItem
	mSync          *systray.MenuItem
	mToggleService *systray.MenuItem
	mQuit          *systray.MenuItem
}

var (
	manager   *TrayManager
	managerMu sync.Mutex
)

// Setup initializes the system tray lifecycle hooks using RunWithExternalLoop.
// Returns the start and end functions to be called with the GTK main loop.
func Setup(actions Actions, isPortuguese bool) (start func(), end func()) {
	managerMu.Lock()
	m := &TrayManager{
		actions:      actions,
		isPortuguese: isPortuguese,
	}
	manager = m
	managerMu.Unlock()

	return systray.RunWithExternalLoop(m.onReady, m.onExit)
}

func (m *TrayManager) onReady() {
	systray.SetIcon(defaultIcon)
	systray.SetTitle("Nuvem")
	if m.isPortuguese {
		systray.SetTooltip("Nuvem - Sincronização Contínua")
	} else {
		systray.SetTooltip("Nuvem - Continuous Synchronization")
	}

	m.mu.Lock()
	if m.isPortuguese {
		m.mOpen = systray.AddMenuItem("Abrir Nuvem", "Exibir janela principal")
		m.mSync = systray.AddMenuItem("Sincronizar agora", "Disparar sincronização imediata")
		systray.AddSeparator()
		m.mToggleService = systray.AddMenuItem("Parar serviço em segundo plano", "Pausar ou retomar sincronização")
		m.mQuit = systray.AddMenuItem("Sair do Nuvem", "Encerrar aplicativo e serviço")
	} else {
		m.mOpen = systray.AddMenuItem("Open Nuvem", "Show main window")
		m.mSync = systray.AddMenuItem("Synchronize now", "Trigger immediate synchronization")
		systray.AddSeparator()
		m.mToggleService = systray.AddMenuItem("Stop background service", "Pause or resume synchronization")
		m.mQuit = systray.AddMenuItem("Quit Nuvem", "Exit application and service")
	}
	m.updateServiceMenuLocked()
	m.mu.Unlock()

	go m.listen()
}

func (m *TrayManager) onExit() {
}

func (m *TrayManager) listen() {
	for {
		select {
		case <-m.mOpen.ClickedCh:
			if m.actions.OnOpen != nil {
				m.actions.OnOpen()
			}
		case <-m.mSync.ClickedCh:
			if m.actions.OnSyncNow != nil {
				m.actions.OnSyncNow()
			}
		case <-m.mToggleService.ClickedCh:
			if m.actions.OnToggleService != nil {
				m.actions.OnToggleService()
			}
		case <-m.mQuit.ClickedCh:
			if m.actions.OnQuit != nil {
				m.actions.OnQuit()
			}
		}
	}
}

// SetServiceActive updates the dynamic menu label based on service state.
func SetServiceActive(active bool) {
	managerMu.Lock()
	m := manager
	managerMu.Unlock()
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.serviceActive = active
	m.updateServiceMenuLocked()
}

func (m *TrayManager) updateServiceMenuLocked() {
	if m.mToggleService == nil {
		return
	}
	if m.serviceActive {
		if m.isPortuguese {
			m.mToggleService.SetTitle("Parar serviço em segundo plano")
		} else {
			m.mToggleService.SetTitle("Stop background service")
		}
	} else {
		if m.isPortuguese {
			m.mToggleService.SetTitle("Iniciar serviço em segundo plano")
		} else {
			m.mToggleService.SetTitle("Start background service")
		}
	}
}

// Quit requests systray cleanup.
func Quit() {
	systray.Quit()
}
