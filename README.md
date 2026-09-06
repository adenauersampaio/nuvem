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
- Conflitos preservam as duas versões até que a pessoa decida.
- Integração com o Google Drive pela API oficial.

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

```text
pasta local <-> Nuvem daemon <-> adaptador Drive <-> Google Drive
                     |
                  SQLite (planejado)
```

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
```

## Licença

MIT. Consulte [LICENSE](LICENSE).
