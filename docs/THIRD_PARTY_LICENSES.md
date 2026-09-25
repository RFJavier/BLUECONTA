# Licencias de terceros

Este proyecto se distribuye bajo Apache License 2.0. Ver `LICENSE` en la raiz.

Las dependencias directas y su licencia declarada:

| Componente | Uso | Licencia |
| --- | --- | --- |
| github.com/wailsapp/wails/v2 | ventana de escritorio | MIT |
| github.com/pocketbase/pocketbase | base local embebida | MIT |
| github.com/pocketbase/dbx | consultas | MIT |
| modernc.org/sqlite | motor SQLite | BSD-3-Clause |
| golang.org/x/crypto, net, sys, text, image, sync, oauth2 | biblioteca estandar extendida | BSD-3-Clause |

Hay dependencias indirectas en `go.mod`. Sus terminos viajan con el modulo correspondiente. Este archivo no reproduce el texto completo de esas licencias.

PocketBase guarda sus propios archivos dentro de `pb_data/`. Esa carpeta es datos del usuario, no codigo de terceros.
