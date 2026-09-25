# Manual de usuario

Contaduria Personal es un programa de escritorio para un estudiante. No necesita internet para registrar movimientos. Sirve para ver cuanto entra, cuanto sale y si el gasto de la semana cabe en el presupuesto.

## Instalar y abrir

1. has click en el ejecutable detro de una carpeta este generara lo encesario para su uso

Al abrir por primera vez, el programa crea la carpeta `pb_data` con tus datos y, dentro de ella, una carpeta `docs` con: `MANUAL.pdf`, `LICENSE.txt` y `LICENCIAS_TERCEROS.md`. Puedes abrir esos archivos en cualquier momento.

## Llevar los datos a otra PC

1. Cierra el programa en ambas computadoras.
2. Copia la carpeta `pb_data` completa (incluye `data.db` y archivos auxiliares).
3. Pegala junto al programa en la otra PC, reemplazando la carpeta vacia si ya existia.
4. Abre el programa. Veras las mismas cuentas, movimientos y presupuestos.

No copies solo `data.db` si hay archivos `-wal` o `-shm`: copia toda la carpeta con el programa cerrado.

## Pantallas

El menu (icono de tres lineas) tiene cuatro opciones visibles.

### Dashboard

- Balance, ingresos y gastos del periodo del ranking.
- Top 5 por ingresos, gastos o comparativa.
- Donut de ingresos contra gastos.
- Bloque "Esta semana": porcentaje usado, gastado, presupuesto, restante, diferencia contra la semana anterior y si tus fondos cubren el plan.
- Cuentas. En un gasto, el numero grande es lo gastado y el numero gris de la derecha es el presupuesto semanal.
- `+ Ingresar` registra una entrada. `+ Gasto` registra una salida. `Presupuesto` define o cambia el tope semanal de esa cuenta.
- `Crear categoria` agrega una cuenta de ingreso o de gasto.

### Historial semanal

Revisa semanas pasadas o la actual.

- Flechas para cambiar de semana. No se puede avanzar mas alla de la semana en curso.
- El mismo donut e indicadores del Dashboard, pero del periodo elegido.
- Tabla de movimientos de esa semana: fecha, cuenta, tipo, concepto, debe y haber.
- `Exportar CSV`: solo los movimientos de esa semana.
- `Exportar Excel`: libro con dos hojas, tambien solo de esa semana.

### Auditoria

Lista general con filtros de fecha, cuenta y tipo. `Esta semana` rellena el rango actual. `Exportar CSV` usa los filtros que tengas aplicados, no solo la semana.

### Configuraciones

- Dias del top de categorias. `0` toma desde el dia 1 del mes.
- Inicio de la semana: lunes a domingo. El valor por defecto es lunes.
- Tope global semanal. `0` lo apaga. Si es mayor que cero, el resumen semanal usa ese tope en lugar de la suma de presupuestos por cuenta.

## Como se usa en la practica

1. Registra el dinero que entra en `Fondos personales` u otra cuenta de ingreso (`+ Ingresar`).
2. Cuando gastes, abre la cuenta de gasto y pulsa `+ Gasto`.
3. En cada cuenta de gasto pulsa `Presupuesto` y escribe el monto semanal. Se guarda una sola vez y se repite cada semana hasta que lo edites.
4. Mira "Esta semana". Verde: dentro del tope. Ambar: mas del 80 %. Rojo: te pasaste. El aviso de fondos dice si el dinero que tienes alcanza para el plan.
5. Al terminar la semana, abre Historial semanal, revisa el donut y, si un asesor lo pide, exporta el Excel.

El presupuesto no es dinero. Es un limite. Por eso puedes definirlo aunque todavia no hayas registrado ingresos. La cobertura te avisa si ese plan no esta respaldado.

## Cuentas iniciales

Solo en una base nueva:

- Ingreso: Fondos personales.
- Gasto: Alimentacion, Transporte, Otros.

Puedes crear mas cuentas. No borres ni renombres una cuenta que ya tenga movimientos: el programa lo impide para no perder el historial.

## Exportaciones

CSV semanal: `Fecha, Cuenta, Tipo, Concepto, Debe, Haber`. Debe es gasto. Haber es ingreso.

Excel (`libro_semanal_....xlsx`):

- Hoja `Libro diario`: encabezado del periodo, moneda USD, movimientos en orden cronologico, debe, haber y saldo del periodo. El saldo no arrastra el saldo inicial.
- Hoja `Presupuesto`: cuenta, presupuesto, ejecutado, variacion, porcentaje y una observacion (dentro del tope, cerca del tope o excedido), mas un resumen para que alguien con formacion contable pueda orientar al estudiante.

No es un estado financiero formal. Es un libro de caja semanal con ejecucion presupuestaria.

## Que no aparece en el menu

Asistencia con IA, calculadora de cobro y la gestion de proveedores o modelos de IA siguen en el codigo, pero estan ocultas. No hacen falta para el uso diario.

## Respaldo

Cierra el programa y copia `pb_data`. Guardala en otra carpeta o en una USB. Para restaurar, cierra el programa y vuelve a colocar esa carpeta.
