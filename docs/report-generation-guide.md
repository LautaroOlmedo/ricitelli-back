# Guía de Generación de Informes PDF — Ricitelli

> **Estado**: guía de diseño e implementación. Primer entregable antes de codificar.
> **Autor**: equipo backend.
> **Referencia externa**: análisis del proyecto `pdf-server` (Python/FastMCP + FPDF + Matplotlib).

Este documento define la arquitectura, decisiones y pasos para implementar un sistema de generación de informes PDF en el backend Go de Ricitelli. Cubre los informes de **Ventas**, **Producción**, **General** y extras de valor. El sistema debe funcionar inicialmente con el `InMemoryRepository` existente y escalar a PostgreSQL sin cambios en la capa de reporting.

---

## 1. Propósito y alcance

### Tipos de informe
| Informe | Alcance | Fuente de datos |
|---|---|---|
| **Ventas** | Métricas comerciales en un rango de fechas | `SaleOrderService.GetSaleOrdersByDateRange` + Customer + Product |
| **Producción** | Agregado de SV, conversiones SV→PT, movimientos de insumos, órdenes de producción | `InventoryService.GetMovements` + `ProductionOrderService.GetProductionOrders` |
| **General** | Dashboard ejecutivo combinando ventas + producción + salud de inventario | Compone los dos anteriores + `InventoryService.GetInventoryReport` |
| **Stock Bajo** (extra) | Insumos bajo `reorder_point` con recomendación de reposición | `InventoryService.GetLowStockAlerts` + histórico de consumo |
| **Trazabilidad de Lote** (extra) | Cadena completa de un `lot_number` (SV→PT→insumos→orden→cliente) | `QueryMovements` filtrado por lot |
| **Cliente** (extra) | Resumen comercial por cliente: órdenes, montos, productos, frecuencia | `CustomerService` + `SaleOrderService` |

### Filtros estándar
- **Rango de fechas** `from_date` / `to_date` en formato **RFC3339** (mismo patrón que `/audit`).
- El frontend envía `datetime-local` y convierte con `new Date(v).toISOString()` (función `toRFC3339` ya existente en `audit/index.astro`).
- Filtros secundarios opcionales: `customer_id`, `product_id`, `market`, `currency`, `lot_number`.

### Canales de consumo
1. **UI Astro** — nueva página `/reports`.
2. **Agente IA** — tools adicionales en el chat existente.

---

## 2. Arquitectura

```
┌─────────────┐         ┌────────────────┐
│ Astro UI    │◄───────►│ API /reports   │
│ /reports    │         │ (Astro route)  │
└─────────────┘         └───────┬────────┘
                                │ gRPC
┌─────────────┐         ┌───────▼────────┐
│ Chat Agent  │◄───────►│ Reporting gRPC │
│ /api/chat   │  tools  │   Service      │
└─────────────┘         └───────┬────────┘
                                │
                    ┌───────────┼───────────┐
                    ▼           ▼           ▼
             ┌──────────┐ ┌──────────┐ ┌──────────┐
             │ Data agg │ │ Charts   │ │ PDF      │
             │ (reuse   │ │ (go-     │ │ builder  │
             │ services)│ │ chart)   │ │ (gofpdf) │
             └──────────┘ └──────────┘ └──────────┘
                                                │
                                                ▼
                                ┌─────────────────────┐
                                │ output/reports/     │
                                │ YYYY-MM-DD/*.pdf    │
                                │ servido vía HTTP    │
                                └─────────────────────┘
```

### Principios de diseño
- **Use-case pattern**: un orquestador por tipo de informe (data → charts → pdf).
- **Generación asíncrona**: cada RPC dispara goroutines para cálculos + generación. El caller recibe `ReportResponse` cuando el PDF está en disco.
- **Sin acoplamiento al repositorio**: la capa de reporting sólo depende de los services ya existentes. Cambio a Postgres → transparente.
- **PNG embebido**: los gráficos se generan en memoria como PNG y se embeben con `gofpdf.RegisterImageReader`. Evita dependencias HTML/CSS.

---

## 3. Stack y dependencias

### Nuevas dependencias Go
| Librería | Propósito | Rationale |
|---|---|---|
| `github.com/jung-kurt/gofpdf` | Generación de PDF | Pure Go, API procedural similar a FPDF. Suficiente para reportes tabulares + imágenes. |
| `github.com/wcharczuk/go-chart/v2` | Gráficos PNG | Simple, sin dependencias nativas, renderiza directo a `io.Writer`. |

### Dependencias que ya existen
- `github.com/google/uuid` — IDs de reportes.
- `github.com/xuri/excelize/v2` — ya disponible; no se usa para reports pero confirma cómo hace el proyecto con assets binarios.
- `github.com/jackc/pgx/v5` — cuando se migre a Postgres, se usa el pool del Repository.

### HTTP para descarga
El backend actual es puro gRPC. Para servir PDFs a navegadores se monta **un handler HTTP adicional** en el mismo proceso (puerto configurable, por defecto `8080`). No se usa grpc-gateway para evitar complejidad. Ruta: `GET /reports/files/{id}` con validación JWT.

---

## 4. Estructura de carpetas

```
internal/service/reporting/
├── service.go                 # Service struct; GenerateSalesReport, GenerateProductionReport, etc.
├── filter.go                  # ReportFilter struct + Validate()
├── data/
│   ├── sales.go               # Aggregator de ventas (reusa SaleOrderService)
│   ├── production.go          # Aggregator de producción (reusa InventoryService + ProductionOrderService)
│   ├── general.go             # Compone sales + production + inventory health
│   ├── lowstock.go            # Stock bajo + sugerencia de reposición
│   ├── lot.go                 # Trazabilidad por lote
│   └── customer.go            # Resumen por cliente
├── charts/
│   ├── chart.go               # Config común (colores, fuentes, tamaños) + RenderPNG(cfg) []byte
│   ├── bar.go                 # Horizontal y vertical bar charts
│   ├── pie.go                 # Pie / donut
│   └── line.go                # Series temporales
├── pdf/
│   ├── builder.go             # PDFReport struct; AddCover, AddFooter, AddImage, AddTable
│   ├── kpi.go                 # AddKPIBoxes(values []KPI) — grilla de métricas
│   └── layouts/
│       ├── sales.go           # Layout del informe de ventas
│       ├── production.go      # Layout del informe de producción
│       └── general.go         # Layout del informe general
└── storage/
    ├── filesystem.go          # Save(pdfBytes, type, from, to) (id, path, url, error)
    ├── metadata.go            # ReportMetadata + index.json con historial
    └── cleanup.go             # Job diario: elimina >REPORTS_RETENTION_DAYS

cmd/http/reporting/
└── reporting.proto            # Definición del nuevo servicio gRPC

cmd/http/gen/reporting/        # Generado por protoc

cmd/http/server/
└── reporting_server.go        # Handler gRPC + handler HTTP /reports/files/{id}
```

---

## 5. Contrato gRPC

Archivo: `cmd/http/reporting/reporting.proto`

```proto
syntax = "proto3";
package reporting;
option go_package = "ricitelli-back/cmd/http/gen/reporting;reportingpb";

import "google/protobuf/empty.proto";

service ReportingService {
  rpc GenerateSalesReport        (ReportFilter)      returns (ReportResponse);
  rpc GenerateProductionReport   (ReportFilter)      returns (ReportResponse);
  rpc GenerateGeneralReport      (ReportFilter)      returns (ReportResponse);
  rpc GenerateLowStockReport     (google.protobuf.Empty) returns (ReportResponse);
  rpc GenerateLotTraceabilityReport (LotRequest)     returns (ReportResponse);
  rpc GenerateCustomerReport     (CustomerReportRequest) returns (ReportResponse);
  rpc ListReports                (ListReportsRequest) returns (ListReportsResponse);
}

message ReportFilter {
  string from_date     = 1;   // RFC3339
  string to_date       = 2;   // RFC3339
  string customer_id   = 3;
  string market        = 4;   // "" | DOMESTIC | EXPORT
  string currency      = 5;
  string product_id    = 6;
}

message LotRequest {
  string lot_number = 1;
}

message CustomerReportRequest {
  string customer_id = 1;
  string from_date   = 2;
  string to_date     = 3;
}

message ReportResponse {
  string id           = 1;    // UUID
  string type         = 2;    // SALES | PRODUCTION | GENERAL | LOW_STOCK | LOT | CUSTOMER
  string filename     = 3;
  string download_url = 4;
  string generated_at = 5;    // RFC3339
  int64  file_size    = 6;
  string from_date    = 7;
  string to_date      = 8;
}

message ListReportsRequest {
  string type      = 1;
  int32  page      = 2;
  int32  page_size = 3;
}

message ListReportsResponse {
  repeated ReportResponse reports = 1;
  int32 total_count = 2;
}
```

---

## 6. Reportes y gráficos

### 6.1 Informe de Ventas

**Fuentes:** `SaleOrderService.GetSaleOrdersByDateRange(from, to)` + `CustomerService.GetCustomerByID` + `ProductService.GetProductByID`.

**KPIs (primer bloque del PDF):**
1. Ingresos totales (ARS equivalente; las monedas extranjeras se reportan por separado).
2. Cantidad de órdenes.
3. Unidades despachadas.
4. Ticket promedio (revenue / órdenes).
5. % Tasa de cumplimiento = `DISPATCHED / total`.

**Gráficos:**
| # | Tipo | Descripción |
|---|---|---|
| 1 | Líneas | Evolución de ingresos por día. |
| 2 | Dona | DOMESTIC vs EXPORT (por cantidad de órdenes). |
| 3 | Barras horizontales | Top 10 productos por ingresos. |
| 4 | Barras verticales | Ingresos agrupados por moneda (ARS, USD, EUR, CAD). |
| 5 | Barras horizontales | Ventas por grupo de cliente (DISTRIBUTOR, WINE_SHOP, RESTAURANT, HOTEL, RETAIL, PRIVATE, EXPORT_AGENT). |
| 6 | Barras horizontales ordenadas | Destinos de exportación por país (ISO alpha-2). |
| 7 | Barras apiladas | Distribución por estado (NEW/CONFIRMED/READY/INVOICED/DISPATCHED/CANCELLED). |

**Tabla final:** Top 20 órdenes ordenadas por monto descendente. Columnas: fecha, cliente, producto principal, monto, moneda, estado.

### 6.2 Informe de Producción

**Fuentes:** `InventoryService.GetMovements(filter)` + `ProductionOrderService.GetProductionOrders()` filtrado por `created_at ∈ [from, to]`.

"Producción" aquí significa: **agregado a SV** (`PRODUCT_PRODUCED`), **conversiones SV→PT** (`PRODUCT_STAGE_IN` dressed), **ingresos de insumos secos** (`DRY_SUPPLY_IN`), **consumos** (`DRY_SUPPLY_CONSUMED`) y **estado de órdenes de producción**.

**KPIs:**
1. Botellas SV producidas.
2. Botellas convertidas SV→PT.
3. Botellas despachadas.
4. Insumos ingresados (unidades).
5. Insumos consumidos (unidades).
6. Órdenes de producción completadas / en curso / canceladas.

**Gráficos:**
| # | Tipo | Descripción |
|---|---|---|
| 1 | Barras agrupadas horizontales | SV vs PT por producto. |
| 2 | Líneas | Conversiones SV→PT por día. |
| 3 | Dos líneas | Ingresos vs consumos de insumos por día. |
| 4 | Barras horizontales | Top 10 insumos más consumidos. |
| 5 | Dona | Consumo de insumos por categoría (LABEL/CORK/CAPSULE/BOTTLE/BOX/CONTRAETIQUETA/OTHER). |
| 6 | Barras verticales | Lotes generados por día (conteo de `lot_number` únicos). |
| 7 | Dona | Estado de órdenes de producción. |

**Tabla final:** últimos 30 movimientos relevantes con fecha, usuario, tipo, ítem, cantidad, lote.

### 6.3 Informe General

Combina KPIs críticos de los dos anteriores más salud de inventario.

**Layout:**
- **Página 1**: Portada con logo, título "Informe General", rango de fechas, timestamp.
- **Página 2**: Dashboard ejecutivo — 12 KPIs en grilla 4×3.
- **Página 3**: Dos gráficos grandes: ingresos por día (líneas) + conversiones SV→PT por día (líneas).
- **Página 4**: Salud de inventario — barras apiladas (físico / comprometido / disponible) por producto + lista de alertas de stock bajo.
- **Páginas 5–6**: Resumen ventas (mercado, top productos, estados).
- **Páginas 7–8**: Resumen producción (SV vs PT, top insumos, categorías).

### 6.4 Extras

#### Informe de Stock Bajo
- Lista de insumos con `available < reorder_point`.
- Para cada insumo: consumo promedio últimos 30 días, **cantidad sugerida de reposición** (`consumo_promedio × factor_stock_seguridad`, default factor = 2).
- Un gráfico: top 15 insumos por nivel de criticidad (available / reorder_point).
- Reusa `GetLowStockAlerts()` y `QueryMovements` para el consumo promedio.

#### Informe de Trazabilidad de Lote
- Input: `lot_number` (p.ej. `L-081124-38-11`).
- Output:
  1. Encabezado con lote, fecha de generación, producto asociado.
  2. Cadena de movimientos: cuándo se convirtió SV→PT (quién, cuántas unidades).
  3. Insumos consumidos en la orden de producción asociada.
  4. Órdenes de venta que recibieron ese lote.
  5. Despachos / clientes destino.
- **Requiere extender `MovementFilter`** para filtrar por `lot_number` en `QueryMovements`.

#### Informe de Cliente
- Input: `customer_id` + rango opcional.
- KPIs: órdenes totales, monto total, ticket promedio, primera/última compra.
- Gráficos: ingresos por mes (barras), productos preferidos (barras horizontales), distribución de monedas (dona).
- Tabla de órdenes en el rango.

---

## 7. Flujo de generación end-to-end

```
1. Cliente (UI o agente IA) invoca ReportingService.GenerateSalesReport(from, to, ...)
   ├─ UI: POST /api/reports/generate → Astro route → gRPC client
   └─ Agente: tool call → executor.ts → gRPC client

2. ReportingService
   ├─ Valida el filtro (from ≤ to, RFC3339 parseable)
   └─ Delega a data/sales.go

3. data/sales.go
   ├─ SaleOrderService.GetSaleOrdersByDateRange(ctx, from, to)
   ├─ Por cada orden: resuelve Customer y Product (name, sku)
   └─ Calcula agregados: por día, mercado, moneda, estado, grupo de cliente, país destino

4. charts/*.go
   └─ Genera cada gráfico como PNG en memoria ([]byte)

5. pdf/layouts/sales.go (usando pdf/builder.go)
   ├─ AddCover("Informe de Ventas", rango)
   ├─ AddKPIBoxes([ingresos, órdenes, ticket, ...])
   ├─ AddImage(chart1_png), AddImage(chart2_png), ...
   ├─ AddTable(top20)
   └─ AddFooter()

6. storage/filesystem.go
   ├─ Genera UUID y path: output/reports/{YYYY-MM-DD}/report-{uuid}.pdf
   ├─ Escribe el PDF a disco
   ├─ Escribe entrada en index.json (metadata + ruta)
   └─ Retorna path + download_url

7. ReportingService retorna ReportResponse{id, filename, download_url, ...}
```

---

## 8. Servido de PDFs por HTTP

- Handler: `GET /reports/files/{id}` — monta un `http.FileServer` pero con lookup por UUID en el índice, no por nombre de archivo directo.
- Auth: reutiliza el secreto JWT del backend. El handler exige header `Authorization: Bearer <token>` o query param `?token=` (para simplificar la descarga desde UI).
- Content-Type: `application/pdf`; Content-Disposition: `attachment; filename="<filename>"`.
- Errores: `404` si el ID no existe, `401` si el token falta/inválido, `410` si fue eliminado por el cleanup.

---

## 9. Integración con la UI Astro

### Página nueva: `src/pages/reports/index.astro`

Estructura inspirada en `/audit`:

1. **Header**: título "Informes" + descripción.
2. **Filtros**:
   - `Tipo de informe` (select): Ventas / Producción / General / Stock bajo / Trazabilidad de lote / Cliente.
   - `Desde` y `Hasta`: `datetime-local`, `min="2006-01-01T00:00"`.
   - **Filtros secundarios contextuales** (visibilidad dinámica según tipo):
     - Ventas: market (dropdown), currency (dropdown), product_id (autocomplete).
     - Cliente: customer_id (autocomplete obligatorio).
     - Trazabilidad: lot_number (input, obligatorio; sin fechas).
3. **Botones**: `Generar informe` + `Limpiar`.
4. **Historial**: tabla con columnas Fecha | Tipo | Rango | Tamaño | Descargar. Paginación estilo `/audit` (ventana ±3 con elipsis).

### API routes nuevas

- `src/pages/api/reports/generate.ts` — POST `{type, from, to, extras}`; llama al cliente gRPC y retorna `ReportResponse`.
- `src/pages/api/reports/list.ts` — GET con `?type=&page=&page_size=`; llama a `ListReports`.

### Cliente gRPC

- `src/lib/grpc/reportingClient.ts` — análogo a `inventoryClient.ts`, expone `generateSalesReport`, `generateProductionReport`, `generateGeneralReport`, `generateLowStockReport`, `generateLotTraceabilityReport`, `generateCustomerReport`, `listReports`.

### Proto en frontend

- `src/proto/reporting.proto` — copia literal del backend (el proyecto ya sigue este patrón para los otros servicios).

### Navegación

- `src/layouts/AppLayout.astro` — agregar entrada "Informes" en el array `nav` (junto a Trazabilidad).

---

## 10. Integración con el agente IA

**Importante:** esta implementación **no usa MCP**. Reutiliza el mecanismo ya implementado del chat:

- Endpoint streaming `/api/chat` (SSE con eventos `token` / `tool_start` / `tool_end` / `done`).
- LLM local (Qwen) invocado con la API OpenAI-compatible (tool/function calling estándar).
- `src/lib/agent/executor.ts` despacha tools a los clientes gRPC del frontend.

### Tools a agregar (`src/lib/agent/tools.ts`)

Siguiendo el mismo schema JSON que las 21 tools existentes:

```ts
{
  name: "generate_sales_report",
  description: "Genera un informe PDF de ventas para un rango de fechas. Opcionalmente filtra por mercado o moneda.",
  parameters: {
    type: "object",
    properties: {
      from: { type: "string", description: "Fecha desde en RFC3339" },
      to:   { type: "string", description: "Fecha hasta en RFC3339" },
      market:   { type: "string", enum: ["DOMESTIC", "EXPORT"] },
      currency: { type: "string", enum: ["ARS", "USD", "EUR", "CAD"] },
    },
    required: ["from", "to"],
  },
}
```

Cinco tools: `generate_sales_report`, `generate_production_report`, `generate_general_report`, `generate_low_stock_report`, `generate_lot_traceability_report` (y opcionalmente `generate_customer_report`).

### Cases en `executor.ts`

Cada tool llama al `reportingClient` correspondiente y retorna `{download_url, filename}`. El LLM recibe este resultado y arma la respuesta final con un link markdown `[Descargar informe de ventas](URL)` que el UI del chat ya sabe renderizar.

---

## 11. Fase 1 — Implementación con `InMemoryRepository`

La primera fase usa los services existentes, que a su vez leen del `InMemoryRepository` (sembrado desde XLSX o fallback hardcoded). Esto garantiza:

- Los reportes funcionan inmediatamente con los datos de ejemplo.
- La misma lógica funcionará cuando el repo sea Postgres, porque **la capa de reporting nunca toca `*postgres.Repository` ni `*InMemoryRepository` directamente** — sólo invoca los services ya definidos en `internal/service/**`.

Datos disponibles hoy en memoria (relevantes para reports):
- `SaleOrderService.GetSaleOrdersByDateRange`, `GetSaleOrders`, `GetSaleOrderByID`.
- `ProductionOrderService.GetProductionOrders`, `GetProductionOrderByID`.
- `InventoryService.GetMovements`, `GetInventoryReport`, `GetLowStockAlerts`.
- `DrySupplyService.GetDrySupplies`, `GetDrySupplyByID`.
- `ProductService.GetProducts`, `GetProductByID`.
- `CustomerService.GetCustomers`, `GetCustomerByID`.

---

## 12. Fase 2 — Escalado a PostgreSQL

### Nada cambia en la capa de reporting
La capa `internal/service/reporting/` depende sólo de interfaces de servicios, no de repositorios. Cuando el backend arranque con `DATABASE_URL` poblado y el `postgres.Repository` esté en uso, los reports funcionan sin modificaciones.

### Optimizaciones recomendadas para Postgres

#### Índices adicionales
```sql
CREATE INDEX IF NOT EXISTS idx_so_status_created  ON sale_orders(status, created_at);
CREATE INDEX IF NOT EXISTS idx_so_market_created  ON sale_orders(market, created_at);
CREATE INDEX IF NOT EXISTS idx_soi_product        ON sale_order_items(product_id);
CREATE INDEX IF NOT EXISTS idx_pim_stage_type_ts  ON product_inventory_movements(stage, movement_type, created_at);
CREATE INDEX IF NOT EXISTS idx_pim_lot            ON product_inventory_movements(lot_number)
  WHERE lot_number IS NOT NULL AND lot_number != '';
CREATE INDEX IF NOT EXISTS idx_dsim_type_ts       ON dry_supply_inventory_movements(movement_type, created_at);
```

#### Agregaciones en SQL
Cuando el volumen justifique, agregar funciones SQL (en `migrations.go`):

```sql
-- Ejemplo: ingresos diarios en un rango
CREATE OR REPLACE FUNCTION sales_daily_revenue(from_ts text, to_ts text)
RETURNS TABLE(day date, revenue numeric, order_count int) AS $$
  SELECT DATE(so.created_at) AS day,
         SUM(soi.quantity * soi.unit_price) AS revenue,
         COUNT(DISTINCT so.id)::int AS order_count
  FROM sale_orders so
  JOIN sale_order_items soi ON soi.sale_order_id = so.id
  WHERE so.created_at >= from_ts AND so.created_at < to_ts
  GROUP BY DATE(so.created_at)
  ORDER BY day;
$$ LANGUAGE SQL STABLE;
```

#### Cache de reportes
Tabla `report_cache`:
```sql
CREATE TABLE report_cache (
  filter_hash TEXT PRIMARY KEY,
  report_type TEXT NOT NULL,
  file_path   TEXT NOT NULL,
  generated_at TIMESTAMPTZ DEFAULT NOW(),
  expires_at  TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_report_cache_expires ON report_cache(expires_at);
```
- TTL por tipo: reportes históricos (rango no incluye hoy) → 24h. Rango que incluye hoy → 5min.
- Hash determinístico del filter (`sha256(type|from|to|extras)`).

#### Materialized views
Para KPIs consultados muchas veces (p.ej. ingresos mensuales por mercado):
```sql
CREATE MATERIALIZED VIEW mv_monthly_revenue_by_market AS
SELECT DATE_TRUNC('month', so.created_at::timestamp) AS month,
       so.market,
       SUM(soi.quantity * soi.unit_price) AS revenue
FROM sale_orders so
JOIN sale_order_items soi ON soi.sale_order_id = so.id
GROUP BY month, so.market;

-- Refresh programado (cron job Go o pg_cron)
REFRESH MATERIALIZED VIEW CONCURRENTLY mv_monthly_revenue_by_market;
```

#### Streaming con cursores
Para rangos muy amplios que podrían cargar millones de filas, usar cursores pgx:
```go
rows, err := pool.Query(ctx, `DECLARE c CURSOR FOR SELECT ...`)
// iterar y agregar en batches, nunca cargar todo en memoria
```

---

## 13. Configuración

Agregar a `config/config.go`:

| Var | Default | Descripción |
|---|---|---|
| `REPORTS_OUTPUT_DIR` | `./output/reports` | Carpeta raíz de PDFs generados. |
| `REPORTS_RETENTION_DAYS` | `30` | Días antes de eliminar PDFs por el cleanup diario. |
| `REPORTS_PUBLIC_URL_BASE` | `http://localhost:8080/reports/files` | Base usada para construir `download_url`. |
| `REPORTS_HTTP_PORT` | `8080` | Puerto del handler HTTP de descarga. |

---

## 14. Testing

- `internal/service/reporting/data/sales_test.go` — mockea `SaleOrderService` y verifica agregaciones (revenue por día, por mercado, etc.) con datasets controlados.
- `internal/service/reporting/charts/bar_test.go` — verifica que `RenderPNG` devuelve bytes > 0 y que la cabecera mágica del PNG es correcta.
- `internal/service/reporting/pdf/builder_test.go` — genera un PDF mínimo y verifica que tiene `%PDF-` al inicio y tamaño > 0.
- **Test de integración** (`reporting_integration_test.go`): arranca el service con `InMemoryRepository` sembrado, genera informe general, verifica:
  - PDF válido (primeros 4 bytes `%PDF`).
  - Más de 3 páginas (parsear con una lib mínima o contar ocurrencias de `/Page`).
  - Tamaño razonable (> 50 KB).

---

## 15. Observabilidad

- **Logs estructurados** por cada request: `report.type`, `filter.from`, `filter.to`, `duration_ms`, `output_size_bytes`, `user_id`, `error` (si aplica). Usar `log/slog` (stdlib) o mantener consistencia con lo que use el resto del backend.
- **Métricas prometheus** (opcional, si el proyecto adopta métricas más adelante):
  - `ricitelli_report_generation_duration_seconds` (histogram, label `type`).
  - `ricitelli_report_generation_errors_total` (counter, label `type`).
  - `ricitelli_reports_total` (counter, label `type`).

---

## 16. Checklist de implementación (orden sugerido)

1. ~~Crear este documento `docs/report-generation-guide.md`~~ ✅ (este archivo).
2. Agregar deps Go: `gofpdf`, `go-chart/v2`.
3. Crear proto `cmd/http/reporting/reporting.proto` + regenerar con `make generate` (agregar entrada al Makefile).
4. Esqueleto: `internal/service/reporting/filter.go` + `service.go` con stubs.
5. `data/sales.go` con agregaciones básicas + test unitario.
6. `charts/bar.go`, `charts/line.go`, `charts/pie.go` + tests.
7. `pdf/builder.go` con portada, KPIs, footer; test básico.
8. Completar `GenerateSalesReport` end-to-end + test de integración.
9. Repetir para `production` y `general`.
10. Agregar extras: low-stock, lot-traceability, customer.
11. Montar handler HTTP `/reports/files/{id}` + cleanup job diario.
12. Registrar el gRPC service en `cmd/main.go` y `cmd/http/server/server.go`.
13. Frontend: `/reports/index.astro`, API routes, cliente gRPC, proto copy.
14. Tools del agente en `tools.ts` + cases en `executor.ts`.
15. Agregar link "Informes" al nav en `AppLayout.astro`.
16. Verificación end-to-end (sección 17).

---

## 17. Verificación end-to-end

### Backend
- `go build ./...` sin errores.
- `go test ./internal/service/reporting/...` todos en verde.
- Backend arranca y loguea `ReportingService registered`.

### gRPC directo
```bash
grpcurl -plaintext \
  -H "authorization: Bearer <token>" \
  -d '{"from_date":"2024-01-01T00:00:00Z","to_date":"2024-12-31T23:59:59Z"}' \
  localhost:50051 reporting.ReportingService/GenerateSalesReport
```
Debe retornar `{"id":"...", "download_url":"http://localhost:8080/reports/files/...", ...}`.

### Descarga HTTP
- `curl -H "Authorization: Bearer <token>" <download_url> -o sales.pdf`
- Abrir con un viewer PDF y verificar: portada, KPIs, 7 gráficos, tabla de top 20.

### Frontend
- Navegar a `/reports` logueado.
- Elegir "Ventas", rango válido, click "Generar informe". Se abre el PDF en nueva pestaña.
- El historial lista el reporte recién generado con paginación correcta si hay más de 1 página.

### Agente IA
- En el chat: _"Dame el informe de ventas del último mes"_. El agente invoca `generate_sales_report` con `from/to` calculados, recibe el `download_url`, responde con un link markdown funcional.
- Probar también: _"Genera el informe de producción de enero"_, _"Dame el estado de stock bajo"_, _"Trazá el lote L-081124-38-11"_.

---

## 18. Resumen ejecutivo

El sistema reutiliza al 100% los servicios existentes del backend (ventas, producción, inventario, movimientos) y sólo agrega una capa nueva de agregación + renderizado. La generación es eventual (no en tiempo real) pero rápida gracias a goroutines. La UI y el agente IA consumen los mismos endpoints gRPC — el frontend descarga vía HTTP tradicional.

La fase 1 funciona contra `InMemoryRepository` y es suficiente para validar UX y contenido. La fase 2 (Postgres) no requiere cambios en la capa de reporting; las optimizaciones (índices, cache, materialized views, cursores) se aplican progresivamente según el volumen real.
