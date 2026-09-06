// Package i18n supplies the first user-facing languages supported by Nuvem.
package i18n

import "strings"

type Language string

const (
	English             Language = "en"
	PortugueseBrazil    Language = "pt-BR"
	KeyError            string   = "error"
	KeyHelp             string   = "help"
	KeyUnknownCommand   string   = "unknown_command"
	KeyConfigMissing    string   = "config_missing"
	KeyConfigCreated    string   = "config_created"
	KeyConfigValid      string   = "config_valid"
	KeySyncFailed       string   = "sync_failed"
	KeyInvalidLanguage  string   = "invalid_language"
	KeyServiceInstalled string   = "service_installed"
	KeyServiceRemoved   string   = "service_removed"
	KeyServiceStatus    string   = "service_status"
	KeySyncInProgress   string   = "sync_in_progress"
	KeyRemoteImported   string   = "remote_imported"
)

var messages = map[Language]map[string]string{
	English: {
		KeyError:            "error:",
		KeyUnknownCommand:   "unknown command %q; use 'nuvem help'",
		KeyConfigMissing:    "No configuration exists yet. Run 'nuvem init --local <folder> --remote <remote>'.",
		KeyConfigCreated:    "Configuration created at %s. Run 'nuvem doctor' to verify it.\n",
		KeyConfigValid:      "Valid configuration for %s → %s.\n",
		KeySyncFailed:       "sync failed:",
		KeyInvalidLanguage:  "unsupported language %q; use en or pt-BR",
		KeyServiceInstalled: "Nuvem service installed and started.\n",
		KeyServiceRemoved:   "Nuvem service stopped and removed.\n",
		KeyServiceStatus:    "Service status: %s\n",
		KeySyncInProgress:   "a Nuvem synchronization is already running",
		KeyRemoteImported:   "Drive profile imported into Nuvem's private configuration.\n",
		KeyHelp: `Nuvem — continuous Google Drive sync for Linux

Usage:
  nuvem [--lang en|pt-BR] version       show the version
  nuvem [--lang en|pt-BR] init          create sync configuration
  nuvem [--lang en|pt-BR] doctor        validate installation and configuration
  nuvem [--lang en|pt-BR] run-once      run one configured sync
  nuvem [--lang en|pt-BR] daemon        keep sync running
  nuvem [--lang en|pt-BR] install       install and start the background service
  nuvem [--lang en|pt-BR] uninstall     remove the background service
  nuvem [--lang en|pt-BR] status        show the service state
  nuvem [--lang en|pt-BR] import-rclone import an existing Drive profile once

The language follows NUVEM_LANG or LANG when --lang is omitted.
`,
	},
	PortugueseBrazil: {
		KeyError:            "erro:",
		KeyUnknownCommand:   "comando desconhecido %q; use 'nuvem help'",
		KeyConfigMissing:    "Configuração ainda não criada. Execute 'nuvem init --local <pasta> --remote <remoto>'.",
		KeyConfigCreated:    "Configuração criada em %s. Execute 'nuvem doctor' para conferir.\n",
		KeyConfigValid:      "Configuração válida para %s → %s.\n",
		KeySyncFailed:       "sincronização falhou:",
		KeyInvalidLanguage:  "idioma não suportado %q; use en ou pt-BR",
		KeyServiceInstalled: "Serviço do Nuvem instalado e iniciado.\n",
		KeyServiceRemoved:   "Serviço do Nuvem interrompido e removido.\n",
		KeyServiceStatus:    "Estado do serviço: %s\n",
		KeySyncInProgress:   "uma sincronização do Nuvem já está em andamento",
		KeyRemoteImported:   "Perfil do Drive importado para a configuração privada do Nuvem.\n",
		KeyHelp: `Nuvem — sincronização contínua com Google Drive no Linux

Uso:
  nuvem [--lang en|pt-BR] version       mostra a versão
  nuvem [--lang en|pt-BR] init          cria a configuração da sincronização
  nuvem [--lang en|pt-BR] doctor        valida a instalação e a configuração
  nuvem [--lang en|pt-BR] run-once      executa uma sincronização configurada
  nuvem [--lang en|pt-BR] daemon        mantém a sincronização em execução
  nuvem [--lang en|pt-BR] install       instala e inicia o serviço em segundo plano
  nuvem [--lang en|pt-BR] uninstall     remove o serviço em segundo plano
  nuvem [--lang en|pt-BR] status        mostra o estado do serviço
  nuvem [--lang en|pt-BR] import-rclone importa uma vez um perfil Drive existente

O idioma segue NUVEM_LANG ou LANG quando --lang não é informado.
`,
	},
}

func Detect(preferred string) Language {
	normalized := strings.ToLower(strings.ReplaceAll(preferred, "_", "-"))
	if strings.HasPrefix(normalized, "pt") {
		return PortugueseBrazil
	}
	return English
}

func Parse(value string) (Language, bool) {
	switch strings.ToLower(strings.ReplaceAll(value, "_", "-")) {
	case "en", "en-us", "en-gb":
		return English, true
	case "pt", "pt-br":
		return PortugueseBrazil, true
	default:
		return English, false
	}
}

func Text(language Language, key string) string {
	if text, ok := messages[language][key]; ok {
		return text
	}
	return messages[English][key]
}
