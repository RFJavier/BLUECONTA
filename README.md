# Contaduria Personal MVP

MVP de escritorio `single-user` y `offline-first` con Go + PocketBase embebido + Wails.

## Arquitectura

```text
MVP/
  backend/
    app.go                       # API expuesta a Wails
    license.go                   # punto de extension para licenciamiento
    bootstrap/
      pocketbase.go              # arranque PocketBase + auto schema + seed inicial
    services/
      accounting_service.go      # logica de negocio
  frontend/
    dist/
      index.html                 # UI por vistas (dashboard/contabilidad/calculadora/auditoria)
      app.js                     # estado UI + llamadas Wails->Go
      styles.css                 # estilos
  shared/
    types.go                     # DTOs y tipos compartidos
  main.go                        # composicion Wails + backend
  wails.json
  go.mod
```

## Vistas del frontend

1. Dashboard
- Balance total
- Ingresos totales
- Egresos totales
- Top 5 categorias (ingresos / egresos / comparativa)
- Crear categoria (modal, verde=ingreso / rojo=egreso)
- Categorias (click en una card para agregar transaccion)
- Desglose por categorias contables (ingresos, egresos y neto)

2. Ayuda
- Asistente IA (analisis del dashboard)
- Historial de analisis guardados (solo lectura)

3. Calculadora
- Calculadora de cobro
- Categorias de cobro

4. Auditoria
- Transacciones recientes
- Filtros por fecha/categoria/tipo
- Exportacion CSV

5. Configuraciones
- Gestion de proveedores API IA
- Gestion de modelos IA por proveedor
- Guardado de API key + modelo activo
- Periodo del top de categorias (dias, 0 = desde el 1 del mes)

## Requisitos

- Go 1.23+
- Wails CLI (opcional, para modo dev/build):

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

## Como correr

```bash
cd MVP
go mod tidy
go run .
```

PocketBase se levanta embebido automaticamente.

## API expuesta al frontend

### Contabilidad
- `CreateTransaction(data)`
- `GetTransactions()`
- `GetTransactionsFiltered(filters)`
- `ExportTransactionsCSV(filters)`
- `GetBalance()`
- `GetDashboardSummary()`
- `GetCategoryRanking(data)`
- `CreateCategory(data)`
- `GetCategories()`
- `UpdateCategory(data)`
- `DeleteCategory(id)`

### Cotizacion / factura rapida
- `CreateBillingCategory(data)`
- `GetBillingCategories()`
- `UpdateBillingCategory(data)`
- `DeleteBillingCategory(id)`
- `CalculateQuote(data)`

### IA (compatible con APIs tipo OpenAI/DeepSeek)
- `GetAIProviders()`
- `CreateAIProvider(data)`
- `GetAIModels(providerID)`
- `CreateAIModel(data)`
- `GetAIConfiguration()`
- `SaveAIConfiguration(data)`
- `AnalyzeDashboardWithAI(input)`
- `GetAIAnalyses()`
- `DeleteAIAnalysis(id)`

### Configuracion
- `GetAppSettings()`
- `SaveAppSettings(data)`

## Nota tecnica del dashboard

`GetDashboardSummary()` agrupa las transacciones por categoria contable y calcula por cada una: ingresos, egresos y neto.

`GetCategoryRanking(data)` devuelve el top de categorias (ingresos/egresos) con el total y el rango de fechas del periodo configurable en `app_settings.ranking_days` (0 = desde el 1 del mes).
