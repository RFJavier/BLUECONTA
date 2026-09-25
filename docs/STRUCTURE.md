# Estructura

```text
BLUECONTA/
  main.go                         composicion Wails + embed del frontend
  wails.json
  go.mod
  README.md
  LICENSE
  pb_data/                        base PocketBase local (no versionar)
  backend/
    app.go                        metodos expuestos a la UI
    license.go                    punto de extension de licencia
    bootstrap/pocketbase.go       esquema, indices y seed de primer arranque
    services/accounting_service.go  reglas de negocio
    services/ledger_xlsx.go       libro Excel semanal
  frontend/dist/
    index.html
    app.js
    styles.css
  frontend/wailsjs/               bindings generados por Wails
  shared/types.go                 DTOs
  docs/
    MANUAL.md
    STRUCTURE.md
    CHANGELOG.md
    THIRD_PARTY_LICENSES.md
  build/                          iconos y binarios locales
```

## Colecciones

- `categories`: nombre y tipo `income` o `expense`.
- `transactions`: tipo, monto, descripcion, `category_id`, `created_at`.
- `budgets`: un tope semanal activo por categoria de gasto.
- `app_settings`: `ranking_days`, `week_start_day` (1=lunes ... 7=domingo), `weekly_budget` (0 = sin tope global).
- `billing_categories`, `ai_providers`, `ai_models`, `ai_settings`, `ai_analyses`: soporte de cotizacion e IA, no visibles en el menu actual.

## Flujo

UI -> `window.go.backend.App` -> `AccountingService` -> PocketBase embebido en `pb_data`.
