# Contaduria Personal

Aplicacion de escritorio, un solo usuario y offline, para que un estudiante controle ingresos, gastos y presupuestos semanales.

Stack: Go + PocketBase embebido + Wails. La interfaz vive en `frontend/dist`.

## Documentacion

- [Manual de usuario](docs/MANUAL.md)
- [Estructura del proyecto](docs/STRUCTURE.md)
- [Changelog](docs/CHANGELOG.md)
- [Licencias de terceros](docs/THIRD_PARTY_LICENSES.md)

## Como ejecutar

Requisitos: Go 1.23+ y, para empaquetar, Wails CLI.

```bash
go run .
```

Los datos quedan en `pb_data/`, junto al ejecutable o a la carpeta del proyecto. Para usar la misma contabilidad en otra PC, copia esa carpeta completa y abre el programa alli.

## Vistas visibles

1. Dashboard: balance, top de categorias, donut ingresos/gastos, resumen de esta semana y cuentas.
2. Historial semanal: semanas anteriores, donut del periodo, movimientos y exportacion CSV/Excel.
3. Auditoria: filtros, filtro rapido de esta semana y CSV.
4. Configuraciones: dias del ranking, inicio de semana y tope global.

Asistencia IA, calculadora y el catalogo de proveedores/modelos siguen en el codigo, pero estan ocultos en la interfaz.

## API principal expuesta al frontend

- Contabilidad: categorias, transacciones, balance, dashboard, ranking, filtros y CSV.
- Semana: `GetWeeklySummary`, `GetWeeklyTransactions`, `GetBudgets`, `SaveBudget`, `DeleteBudget`, `ExportWeeklyCSV`, `ExportWeeklyLedgerXLSX`.
- Ajustes: `GetAppSettings`, `SaveAppSettings` (`ranking_days`, `week_start_day`, `weekly_budget`).
