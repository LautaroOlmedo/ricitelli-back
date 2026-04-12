# BOM Inference Strategy: Product-to-DrySupply Mapping

## Problema

Los archivos xlsx de la bodega no contienen un mapeo explicito entre productos
(vinos) y los insumos secos (etiquetas, contraetiquetas, capsulas, cajas, etc.)
que los componen. Cada producto necesita un Bill of Materials (BOM) que defina
que insumos y en que cantidad se requieren para vestir una unidad del producto.

### Archivos xlsx disponibles

| Archivo | Contenido | Tiene BOM? |
|---------|-----------|------------|
| INSUMOS SECOS.xlsx | Catalogo de ~400 insumos con stock fisico | No |
| VESTIDO Y SV.xlsx | Catalogo de ~42 productos con stock SV/PT | No |
| Insumos Comprometidos.xlsx | Movimientos de reserva por OP | No |
| PENDIENTES.xlsx | Pedidos de venta pendientes | No |
| REMITIDOS.xlsx | Pedidos de venta despachados | No |

Ninguno de los 5 archivos contiene la relacion producto <-> insumo seco.

## Solucion actual: Inferencia por nombre (v1)

Implementada en `xlsx_seeder.go` funcion `linkProductsToDrySupplies()`.

### Algoritmo

1. Al cargar productos desde VESTIDO Y SV.xlsx, se almacena `ord1` (linea de
   producto, ej: "Hey !") y `ord2` (varietal, ej: "Malbec") por cada producto.

2. Se cuentan cuantos productos comparten el mismo `ord1`:
   - **Linea unica** (1 varietal): se matchea solo con tokens de `ord1`
   - **Linea multiple** (N varietales): se matchea con tokens de `ord1 + ord2`

3. Normalizacion de nombres:
   - Lowercase
   - Remover puntuacion (!, ., ,, ´, etc.)
   - Colapsar espacios
   - Filtrar "stop words": de, la, del, y, the, and, from, is, not
   - Filtrar tokens con < 2 caracteres

4. Para cada insumo seco, se verifica si su nombre normalizado contiene TODOS
   los tokens del producto. Si matchea, se agrega al BOM con `QuantityPerUnit = 1`.

### Resultado actual

- **29 de 42 productos** del xlsx obtienen BOM (~69% de cobertura)
- Los 13 restantes no matchean por inconsistencias en los datos:
  - Typos en los xlsx: "Ricicitelli" vs "Riccitelli"
  - Nombres abreviados: "Caber Franc" vs "Cabernet Franc"
  - Nombres diferentes: "Viñedos en Pie Franco" vs "Viejos Viñedos Pie Franco"
  - Productos sin insumos en el catalogo: Aceite de Oliva, Riccitelli Rancio

### Limitaciones

1. **QuantityPerUnit siempre es 1** — no se puede inferir la cantidad real
   desde los nombres. En la realidad algunos productos pueden requerir >1
   unidad de un insumo.

2. **Falsos positivos** — un insumo para un mercado especifico (ej: "Contraet
   Hey Malbec Japon") se asigna al producto Hey Malbec aunque solo se usa
   para ordenes de exportacion a Japon.

3. **Falsos negativos** — productos con nombres inconsistentes en los xlsx
   no matchean (ver lista arriba).

4. **Productos de PENDIENTES/REMITIDOS** — los productos creados desde ordenes
   de venta no tienen `ord1`/`ord2` y no participan en la inferencia de BOM.

## Propuesta a futuro: Solucion escalable (v2)

Cuando las bodegas implementen este software en produccion, la inferencia por
nombre debe reemplazarse por un mapeo explicito. Opciones:

### Opcion A: Archivo xlsx de BOM (recomendada)

Agregar un sexto archivo xlsx al directorio `data/`:

```
BOM.xlsx
Columnas:
  - Producto Ord.1    (linea de producto, ej: "Hey !")
  - Producto Ord.2    (varietal, ej: "Malbec")
  - Insumo Cod.       (codigo ERP del insumo, ej: "IF1068")
  - Cantidad por Und.  (cantidad requerida, ej: 1)
```

Cambios necesarios:
- Nueva funcion `loadBOM(path string) error` en `xlsx_seeder.go`
- Matchear producto por `ord1|ord2` key
- Matchear insumo por codigo ERP
- Reemplazar `linkProductsToDrySupplies()` por la carga directa

### Opcion B: Modulo de BOM en la UI

Permitir que los usuarios definan el BOM desde la interfaz web:
- La pagina de edicion de producto ya soporta agregar/quitar insumos al BOM
- Al crear un producto nuevo, se seleccionan los insumos desde un dropdown
- Los datos se persisten en la tabla `product_bill_of_dry_supply`

Esto ya funciona hoy con el backend PostgreSQL. La opcion no requiere cambios
de codigo, solo disciplina operativa.

### Opcion C: Integracion con ERP

Si la bodega usa un ERP (ej: sistema contable, Bodegas+, etc.) que ya tiene
la estructura de BOM definida, se puede:
- Exportar el BOM desde el ERP como xlsx/csv
- Crear un endpoint de importacion batch
- Sincronizar periodicamente

## Archivos relevantes

| Archivo | Descripcion |
|---------|-------------|
| `internal/infraestructure/in-memory/xlsx_seeder.go` | Seeder con inferencia BOM |
| `internal/domain/product/product.go` | Aggregate Product con campo `bods` |
| `internal/value-object/bill_of_dry_supply.go` | Value object BillOfDrySupply |
| `internal/infraestructure/postgres/repository.go` | Repo PostgreSQL con tabla `product_bill_of_dry_supply` |
| `db/init/init.sql` | Schema SQL con tabla de BOM |
