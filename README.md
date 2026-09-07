# Nuvem

Nuvem é um cliente de sincronização contínua para Google Drive no Linux.
Ele foi pensado para oferecer a experiência de uma pasta local comum: você
escolhe a pasta uma vez, e o aplicativo cuida de acompanhar alterações,
retomar transferências e mostrar conflitos de forma clara.

> Estado: beta funcional. Mantenha cópias de segurança independentes enquanto o motor de sincronização evolui.

## Princípios

- Configuração em um comando, sem exigir timers ou scripts manuais.
- Uma única execução por pasta, para impedir sincronizações concorrentes.
- Estado local persistente e diagnóstico compreensível.
- Em conflitos, a versão modificada mais recentemente vence e a decisão fica registrada pelo motor nativo.
- A integração inicial com Drive usa componentes MIT do rclone incorporados ao executável; não exige o programa `rclone` instalado.

## Idiomas

A primeira versão terá inglês e português do Brasil. O idioma poderá ser
escolhido pela interface e, no terminal, com `--lang en` ou `--lang pt-BR`.
Quando não for informado, o Nuvem seguirá `NUVEM_LANG` ou `LANG`.

## Arquitetura inicial

O executável `nuvem` contém o agendador, a configuração e uma integração
embutida com os componentes MIT do rclone necessários para Google Drive e
sincronização bidirecional. Não é necessário instalar o executável `rclone`.
O controle de estado e a experiência de uso pertencem ao Nuvem.

O comando `nuvem install` registra e inicia o daemon como um serviço de usuário
automaticamente. Não há timer para criar ou administrar manualmente.

O novo motor independente já define uma interface única para armazenamentos,
uma implementação de pasta local, planejamento de cópias e um diário
append-only de decisões. Adaptadores de Google Drive, Dropbox e OneDrive
poderão reutilizar esse núcleo. Nesta fase, exclusões continuam fora desse
motor até haver uma base histórica segura para tratá-las.

## Interface gráfica

A interface inicial é GTK4 nativa para Linux, mas o núcleo continua sem
dependência gráfica para permitir clientes futuros em Windows e macOS.
Ela permite escolher a pasta local pelo seletor do sistema, configurar a
frequência, salvar a configuração e iniciar uma sincronização. Na conexão com
o Google Drive, o navegador abre o seletor oficial do Google para que a pessoa
escolha exatamente a pasta que o Nuvem poderá usar.

Além do Go 1.22, a compilação da interface requer GTK4 de desenvolvimento e
`pkg-config` (em Debian/Ubuntu: `libgtk-4-dev pkg-config`).

```bash
go run ./cmd/nuvem-desktop
```

Para instalá-lo no menu de aplicativos do Linux após compilar o executável:

```bash
nuvem-desktop --install
```

## Conexão com Google Drive

Para a versão atual, crie um cliente OAuth do tipo **Desktop** no seu projeto
Google Cloud e habilite a **Google Drive API** e a **Google Picker API**.
Informe o ID e o segredo na janela Nuvem e clique em **Conectar e escolher
pasta**. O navegador abre o consentimento do Google e o seletor oficial da
pasta. O Nuvem solicita somente o escopo `https://www.googleapis.com/auth/drive.file`:
o acesso fica limitado aos arquivos da pasta que a pessoa escolheu, sem o
escopo restrito de acesso a todo o Drive.

O token e as credenciais ficam somente em `~/.config/nuvem/config.json`, que o
Nuvem grava com permissão exclusiva do usuário. O modo nativo pode ser
verificado sem mudar o serviço atual:

```bash
nuvem run-once --engine native
```

Depois dessa conexão, o serviço passa automaticamente ao motor nativo do
Nuvem para aquela pasta escolhida.

O adaptador próprio sincroniza arquivos e pastas comuns, inclusive quando a
pasta raiz informada pertence a um Drive compartilhado. Documentos nativos do
Google (Docs, Sheets e Slides) são ignorados por enquanto, para não gerar uma
exportação com formato diferente sem uma escolha explícita do usuário.

## Desenvolvimento

Requer Go 1.22 ou superior.

```bash
go test ./...
go run ./cmd/nuvem version
go run ./cmd/nuvem init --local /caminho/para/pasta --remote GoogleDrive:NomeDaPasta
go run ./cmd/nuvem import-rclone --remote GoogleDrive
go run ./cmd/nuvem doctor
go run ./cmd/nuvem install --binary /caminho/absoluto/para/nuvem
go run ./cmd/nuvem status
go run ./cmd/nuvem-desktop
```

## Apoie o projeto / Support

Se este software está sendo útil para você, cogite deixar um café (**US$ 1** ou quanto preferir)! O seu apoio incentiva a continuidade e a evolução do Nuvem:

- ☕ **Buy Me a Coffee**: [buymeacoffee.com/adetech](https://buymeacoffee.com/adetech)
- 🇧🇷 **Pix (Livepix)**: [livepix.gg/adetech](https://livepix.gg/adetech)
- 💛 **Binance Pay (Cripto)**: [Doar via Binance](https://app.binance.com/uni-qr/request-to-pay?billOrderId=452917517181927424&billType=request_a_payment)
- 💖 **GitHub Sponsors**: [github.com/sponsors/adenauersampaio](https://github.com/sponsors/adenauersampaio)

---

*If this software is useful to you, consider buying me a coffee (**$1** or whatever you like)! Your support keeps Nuvem active and evolving:*

- ☕ **Buy Me a Coffee**: [buymeacoffee.com/adetech](https://buymeacoffee.com/adetech)
- 💖 **GitHub Sponsors**: [github.com/sponsors/adenauersampaio](https://github.com/sponsors/adenauersampaio)
- 💛 **Binance Pay (Crypto)**: [Donate via Binance](https://app.binance.com/uni-qr/request-to-pay?billOrderId=452917517181927424&billType=request_a_payment)
- 🇧🇷 **Pix (Brazil)**: [livepix.gg/adetech](https://livepix.gg/adetech)

## Licença

MIT. Consulte [LICENSE](LICENSE).
