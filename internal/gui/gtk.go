// Package gui provides the Linux-native GTK control surface.
package gui

/*
#cgo pkg-config: gtk4
#include <stdlib.h>
#include "gtk.h"
*/
import "C"

import (
	"sync"
	"unsafe"
)

type Controller interface {
	Activated()
	SyncNow()
	Save(local, remote string, minutes int)
	ConnectGoogleDrive()
}

type Dashboard struct {
	Local      string
	Remote     string
	Minutes    int
	Service    string
	Configured bool
}

var (
	controller Controller
	mu         sync.RWMutex
)

func Run(c Controller, portuguese bool) {
	mu.Lock()
	controller = c
	mu.Unlock()
	if portuguese {
		C.nuvem_set_portuguese(1)
	} else {
		C.nuvem_set_portuguese(0)
	}
	C.nuvem_run()
}

func SetStatus(text string, isError bool) {
	value := C.CString(text)
	defer C.free(unsafePointer(value))
	if isError {
		C.nuvem_set_status(value, 1)
	} else {
		C.nuvem_set_status(value, 0)
	}
}

func SetDashboard(value Dashboard) {
	local, remote, service := C.CString(value.Local), C.CString(value.Remote), C.CString(value.Service)
	defer C.free(unsafePointer(local))
	defer C.free(unsafePointer(remote))
	defer C.free(unsafePointer(service))
	configured := 0
	if value.Configured {
		configured = 1
	}
	C.nuvem_set_dashboard(local, remote, C.int(value.Minutes), service, C.int(configured))
}

func SetDonationURLs(bmc, livepix, binance, github string) {
	cBmc := C.CString(bmc)
	defer C.free(unsafePointer(cBmc))
	cLivepix := C.CString(livepix)
	defer C.free(unsafePointer(cLivepix))
	cBinance := C.CString(binance)
	defer C.free(unsafePointer(cBinance))
	cGithub := C.CString(github)
	defer C.free(unsafePointer(cGithub))
	C.nuvem_set_donation_urls(cBmc, cLivepix, cBinance, cGithub)
}

func unsafePointer(value *C.char) unsafe.Pointer { return unsafe.Pointer(value) }

func currentController() Controller { mu.RLock(); defer mu.RUnlock(); return controller }

//export goNuvemActivated
func goNuvemActivated() {
	if c := currentController(); c != nil {
		go c.Activated()
	}
}

//export goNuvemSyncNow
func goNuvemSyncNow() {
	if c := currentController(); c != nil {
		go c.SyncNow()
	}
}

//export goNuvemSaveConfig
func goNuvemSaveConfig(local, remote *C.char, minutes C.int) {
	if c := currentController(); c != nil {
		go c.Save(C.GoString(local), C.GoString(remote), int(minutes))
	}
}

//export goNuvemConnectGoogle
func goNuvemConnectGoogle() {
	if c := currentController(); c != nil {
		go c.ConnectGoogleDrive()
	}
}
