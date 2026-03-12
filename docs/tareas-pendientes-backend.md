# Tareas Pendientes - Backend (Ricitelli)

Fecha de análisis: 2026-03-12
Basado en: `riccitelli-guide.md` + análisis del código actual

---

## Estado Actual del Backend

**Implementado correctamente:**
- Dominio completo: Customer, Product, SaleOrder, ProductionOrder, DrySupply, ProductInventory, DrySupplyInventory
- Stock Tricapa (Físico, Comprometido, Disponible) para productos y dry supplies
- Flujo SV → PT con número de lote
- Orquestación atómica `CreateOrder` (venta + producción + consumo de insumos)
- Múltiples monedas (ARS, USD, CAD, EUR), múltiples mercados (DOMESTIC/EXPORT), múltiples tipos de venta
- Alertas de stock bajo (umbral hardcodeado en 500 unidades)
- Comunicación gRPC con Protobuf
- Repositorio in-memory con seeding desde XLSX

---

## TAREAS PENDIENTES

---

### TAREA BE-01: Persistencia Real con Base de Datos (PostgreSQL)

**Prioridad:** Crítica
**Módulo:** `internal/infraestructure/`

**Descripción:**
El sistema actualmente usa un repositorio en memoria (`in-memory`). Todo se pierde al reiniciar el servidor. Se debe implementar una capa de persistencia real con PostgreSQL.

**Pasos detallados:**

1. **Agregar dependencia PostgreSQL al módulo Go:**
   ```
   go get github.com/jackc/pgx/v5
   go get github.com/jackc/pgx/v5/pgxpool
   ```

2. **Crear archivo de configuración de DB en `config/config.go`:**
   - Agregar campo `DatabaseURL string` leído desde env var `DATABASE_URL`
   - Formato esperado: `postgres://user:password@host:5432/ricitelli_db`

3. **Crear directorio `internal/infraestructure/postgres/`** con los siguientes archivos:

   - `connection.go`: Función `NewPool(databaseURL string) (*pgxpool.Pool, error)` que inicializa el pool de conexiones con configuración de max_conns, min_conns y timeouts.

   - `migrations.go`: Función `RunMigrations(pool *pgxpool.Pool) error` que ejecuta las migraciones DDL en orden. Las migraciones deben crear las tablas:
     - `customers` (id UUID PK, social_reason TEXT, market_type TEXT, group TEXT, active BOOL, created_at TIMESTAMPTZ)
     - `products` (id UUID PK, name TEXT, active BOOL)
     - `product_bods` (id UUID PK, product_id UUID FK, dry_supply_id UUID, quantity_per_unit FLOAT)
     - `dry_supplies` (id UUID PK, code TEXT UNIQUE, name TEXT, category TEXT, unit TEXT)
     - `sale_orders` (id UUID PK, customer_id UUID FK, status TEXT, currency TEXT, market TEXT, destination_country TEXT, sale_type TEXT, created_at TIMESTAMPTZ, active BOOL)
     - `sale_order_items` (id UUID PK, sale_order_id UUID FK, product_id UUID, quantity INT, unit_price FLOAT)
     - `production_orders` (id UUID PK, sale_order_id UUID FK, status TEXT, created_at TIMESTAMPTZ, active BOOL)
     - `production_items` (id UUID PK, production_order_id UUID FK, product_id UUID, quantity INT)
     - `product_inventories` (id UUID PK, product_id UUID FK UNIQUE, sku TEXT)
     - `product_movements` (id UUID PK, inventory_id UUID FK, movement_type TEXT, stage TEXT, quantity INT, lot_number TEXT, reference TEXT, created_at TIMESTAMPTZ)
     - `dry_supply_inventories` (id UUID PK, dry_supply_id UUID FK UNIQUE)
     - `dry_supply_movements` (id UUID PK, inventory_id UUID FK, movement_type TEXT, quantity INT, reference TEXT, created_at TIMESTAMPTZ)

   - `repository.go`: Implementación de la interfaz del repositorio usando `pgxpool.Pool`. Debe implementar los mismos métodos que el repositorio in-memory. Usar transacciones para operaciones que modifican múltiples tablas.

4. **Actualizar `cmd/main.go`** para:
   - Leer `DATABASE_URL` desde el entorno
   - Si `DATABASE_URL` está presente, usar el repositorio PostgreSQL
   - Si no está presente, usar el repositorio in-memory (para desarrollo local sin DB)
   - Ejecutar migraciones al iniciar

5. **Agregar `db/schema.sql`**: Dump completo del schema para referencia y deployments manuales.

**Criterios de aceptación:**
- El servidor arranca con DB configurada sin errores
- Los datos persisten entre reinicios
- Las operaciones atómicas (`CreateOrder`) usan transacciones
- Si no hay `DATABASE_URL`, cae en in-memory sin error

---

### TAREA BE-02: Umbral de Stock Configurable por Insumo

**Prioridad:** Alta
**Módulo:** `internal/domain/dry-supply/`, `internal/service/inventory/`

**Descripción:**
El umbral de alerta de stock bajo está hardcodeado en `500` unidades en `internal/service/inventory/service.go`. Cada insumo debe poder tener su propio punto de pedido (reorder point).

**Pasos detallados:**

1. **Modificar `internal/domain/dry-supply/dry_supply.go`:**
   - Agregar campo `ReorderPoint int` al struct `DrySupply`
   - Agregar constructor que acepte `reorderPoint` como parámetro
   - El valor default debe ser `0` (sin alerta)

2. **Modificar el proto `dry-supply.proto`:**
   - Agregar campo `reorder_point int32` al mensaje `DrySupply`
   - Agregar campo `reorder_point` al mensaje `CreateDrySupplyRequest`
   - Regenerar los archivos Go con `make proto` o similar

3. **Modificar `internal/service/inventory/service.go`:**
   - En la función que genera alertas, reemplazar el umbral fijo `500` por `drySupply.ReorderPoint`
   - Si `ReorderPoint == 0`, no generar alerta para ese insumo
   - Agregar campo `ReorderPoint` en el output del reporte

4. **Modificar `cmd/http/server/server.go`:**
   - En `CreateDrySupply` handler: mapear `request.ReorderPoint` → `domain.ReorderPoint`
   - En respuestas de `DrySupply`: incluir `ReorderPoint` en el proto response

5. **Agregar `UpdateDrySupply` endpoint** (ver TAREA BE-04) para poder modificar el reorder point sin recrear el insumo.

**Criterios de aceptación:**
- Crear un insumo con `reorder_point: 200` y otro con `reorder_point: 0`
- `GetLowStockAlerts` solo devuelve el primero si su disponible < 200
- El segundo nunca aparece en alertas

---

### TAREA BE-03: Endpoints CRUD Faltantes (Update/Delete)

**Prioridad:** Alta
**Módulo:** `internal/service/`, `cmd/http/server/`

**Descripción:**
Faltan operaciones de actualización y eliminación para las entidades principales. El sistema solo permite crear y leer.

**Sub-tareas:**

#### BE-03a: `UpdateProduct`
1. Agregar caso `UpdateProduct` al proto `product.proto`:
   - Request: `{ id: string, name: string, bods: [...] }`
   - Response: `Product`
2. Crear `internal/service/product/update_product.go`:
   - Recuperar producto por ID, validar que existe y está activo
   - Actualizar `Name` y reemplazar `BODS` completos
   - Guardar en repositorio
3. Agregar interfaz al repositorio: `SaveProduct(product *domain.Product) error`
4. Implementar en in-memory y PostgreSQL
5. Registrar handler en `server.go`

#### BE-03b: `UpdateCustomer`
1. Agregar caso `UpdateCustomer` al proto `customer.proto`:
   - Request: `{ id: string, social_reason: string, market_type: string, group: string }`
   - Response: `Customer`
2. Crear `internal/service/customer/update_customer.go`:
   - Recuperar cliente, actualizar campos, guardar
3. Implementar en repositorios
4. Registrar en `server.go`

#### BE-03c: `UpdateDrySupply`
1. Agregar caso `UpdateDrySupply` al proto `dry-supply.proto`:
   - Request: `{ id: string, name: string, reorder_point: int32 }`
   - Response: `DrySupply`
2. Crear `internal/service/dry-supply/update_dry_supply.go`
3. Implementar en repositorios
4. Registrar en `server.go`

#### BE-03d: `CancelSaleOrder` con razón
1. Modificar `UpdateSaleOrderStatus` para aceptar un campo opcional `cancellation_reason: string`
2. Agregar campo `CancellationReason string` al dominio `SaleOrder`
3. Al cancelar, registrar la razón y revertir el stock comprometido (llamar a `ReleaseStock` en los insumos y `ReleaseReservation` en el inventario de productos)

**Criterios de aceptación:**
- Actualizar nombre de producto y su BOM sin perder historial de movimientos
- Cancelar una orden con razón libera correctamente el stock comprometido

---

### TAREA BE-04: Generación Automática de Número de Lote

**Prioridad:** Media
**Módulo:** `internal/service/inventory/`, `internal/service/application-service/`

**Descripción:**
La bodega usa números de lote con formato `L-DDMMYY-NNN-XX` (ej: `L-081124-38-11`). Actualmente el lote se ingresa manualmente en `ConvertSVtoPT`. Debe generarse automáticamente si no se provee.

**Pasos detallados:**

1. **Crear `internal/service/inventory/lot_generator.go`:**
   ```go
   // GenerateLotNumber genera un número de lote en formato L-DDMMYY-NNN-XX
   // donde NNN es el contador del día y XX es el ID del producto truncado
   func GenerateLotNumber(productID string, dailyCounter int) string
   ```
   - Formato: `L-` + fecha en `DDMMYY` + `-` + contador diario (3 dígitos con padding 0) + `-` + últimos 2 caracteres del productID
   - Ejemplo: `L-120326-001-f3`

2. **Modificar `ConvertSVtoPT` en `inventory/service.go`:**
   - Si `lot_number` en el request está vacío, llamar a `GenerateLotNumber`
   - Para el contador diario, recuperar del repositorio cuántos lotes se generaron hoy

3. **Modificar `application-service/create_order.go`:**
   - Al llamar al proceso de vestido durante `CreateOrder`, usar `GenerateLotNumber` en lugar de un lote fijo o vacío

4. **Agregar al repositorio:** `GetLotCountForDate(date time.Time) (int, error)` para recuperar la cantidad de lotes del día.

**Criterios de aceptación:**
- Llamar a `ConvertSVtoPT` sin `lot_number` devuelve un lote generado automáticamente con el formato correcto
- Dos conversiones en el mismo día tienen contadores diferentes: `001`, `002`, etc.

---

### TAREA BE-05: Filtros y Paginación en Listados

**Prioridad:** Media
**Módulo:** `internal/service/`, protos, `cmd/http/server/`

**Descripción:**
Los endpoints `GetSaleOrders`, `GetProductionOrders`, `GetCustomers` y `GetProducts` devuelven todos los registros sin filtros ni paginación. Con volumen de datos real, esto será un problema de performance.

**Pasos detallados:**

1. **Para `GetSaleOrders`**, agregar al proto request:
   - `page int32`, `page_size int32` (default 50)
   - `status string` (filtrar por estado)
   - `market string` (DOMESTIC/EXPORT)
   - `sale_type string`
   - `customer_id string`
   - `from_date string`, `to_date string` (ISO 8601)

   Retornar también: `total int32` en la response.

2. **Para `GetProductionOrders`**, agregar:
   - `page`, `page_size`
   - `status string`
   - `sale_order_id string`

3. **Para `GetCustomers`**, agregar:
   - `page`, `page_size`
   - `market_type string`
   - `group string`
   - `active bool` (default: solo activos)
   - `search string` (busca por social_reason, case-insensitive)

4. **Implementar en cada service** la lógica de filtrado:
   - En in-memory: filtrar en Go después de recuperar todos
   - En PostgreSQL: agregar WHERE clauses dinámicas

5. **El endpoint `GetSaleOrdersByDateRange`** ya existe en el servicio; integrarlo en `GetSaleOrders` como filtro adicional y deprecar el endpoint separado.

**Criterios de aceptación:**
- `GetSaleOrders` con `status: "DISPATCHED"` solo devuelve despachadas
- `GetSaleOrders` con `page: 1, page_size: 10` devuelve máximo 10 resultados y el total

---

### TAREA BE-06: Endpoint `GetProductionOrdersBySaleOrder`

**Prioridad:** Media
**Módulo:** `internal/service/production-order/`, protos

**Descripción:**
No existe una forma directa de obtener las órdenes de producción asociadas a una orden de venta específica. Solo se puede obtener por ID individual.

**Pasos detallados:**

1. **Agregar al proto `production-order.proto`:**
   - Request: `GetProductionOrdersBySaleOrderRequest { sale_order_id: string }`
   - Response: reutilizar `GetProductionOrdersResponse`

2. **Crear `internal/service/production-order/get_production_orders_by_sale_order.go`:**
   - Filtrar todas las órdenes de producción donde `SalesOrderID == saleOrderID`
   - Retornar slice de production orders

3. **Agregar al repositorio:** `GetProductionOrdersBySaleOrderID(saleOrderID string) ([]*domain.ProductionOrder, error)`

4. **Registrar en `server.go`**

**Criterios de aceptación:**
- Crear una orden que genera una producción asociada
- Llamar al nuevo endpoint con el ID de la orden de venta retorna la producción vinculada

---

### TAREA BE-07: Módulo de Reportes Comerciales

**Prioridad:** Media
**Módulo:** `internal/service/` (nuevo: `reporting/`), protos (nuevo: `reporting.proto`)

**Descripción:**
El guide requiere dashboards de: Ventas por Mercado/País, Rendimiento por Línea de Vino, y Control de Fugas (muestras/obsequios). Estos requieren agregaciones sobre las órdenes de venta.

**Pasos detallados:**

1. **Crear `internal/service/reporting/service.go`** con las siguientes funciones:

   - `GetSalesByMarket(from, to time.Time) []MarketSalesSummary`
     - Agrupa órdenes DESPACHADAS por `Market` (DOMESTIC/EXPORT)
     - Cuenta pedidos y suma cantidades totales de botellas
     - Incluye breakdown por `DestinationCountry` para EXPORT

   - `GetProductPerformance(from, to time.Time) []ProductPerformanceSummary`
     - Agrupa items de órdenes DESPACHADAS por `ProductID`
     - Suma cantidad total vendida y valor total (quantity × unit_price)
     - Ordena por cantidad descendente

   - `GetNonCommercialOutflows(from, to time.Time) []NonCommercialOutflowSummary`
     - Agrupa órdenes por `SaleType` excluyendo `SALE`
     - Tipos: `SAMPLE_CUSTOMS`, `GIFT`, `INTERNAL`, `COMMERCIAL_SAMPLE`
     - Suma botellas y valor estimado

2. **Crear `cmd/http/gen/reporting/reporting.proto`** con los mensajes necesarios y el servicio `ReportingService`.

3. **Registrar el servicio en `cmd/http/server/server.go`** y en `cmd/main.go`.

4. **Los reportes deben aceptar parámetros `from_date` y `to_date`** en formato ISO 8601. Si no se proveen, defaultear a los últimos 30 días.

**Criterios de aceptación:**
- `GetSalesByMarket` con rango de fechas devuelve totales por DOMESTIC vs EXPORT
- `GetProductPerformance` devuelve ranking de productos más vendidos
- `GetNonCommercialOutflows` diferencia muestras de aduana vs obsequios vs consumo interno

---

### TAREA BE-08: Autenticación JWT

**Prioridad:** Alta
**Módulo:** nuevo `internal/auth/`, `cmd/http/server/`

**Descripción:**
El sistema no tiene ningún mecanismo de autenticación. Cualquier cliente puede leer y modificar todos los datos.

**Pasos detallados:**

1. **Agregar dependencias:**
   ```
   go get github.com/golang-jwt/jwt/v5
   ```

2. **Crear `internal/auth/jwt.go`:**
   - `GenerateToken(userID, role string, secret string) (string, error)` — genera JWT con claims: `sub`, `role`, `exp` (24h)
   - `ValidateToken(token, secret string) (*Claims, error)` — valida y parsea el token

3. **Crear `internal/auth/interceptor.go`:**
   - gRPC UnaryInterceptor que extrae el token del metadata `authorization: Bearer <token>`
   - Valida el token con `ValidateToken`
   - Inyecta los claims en el contexto con `context.WithValue`
   - Retorna `codes.Unauthenticated` si no hay token o es inválido
   - Permite sin auth: solo el endpoint de login

4. **Crear un endpoint `Login`** (nuevo servicio `AuthService`):
   - Request: `{ username: string, password: string }`
   - Valida contra usuarios hardcodeados (primera versión) o tabla `users` en DB
   - Response: `{ token: string, expires_at: string }`

5. **Registrar el interceptor en `cmd/main.go`** al crear el servidor gRPC:
   ```go
   grpc.NewServer(grpc.UnaryInterceptor(auth.NewInterceptor(secret)))
   ```

6. **Variables de entorno nuevas:** `JWT_SECRET`, `AUTH_ADMIN_USER`, `AUTH_ADMIN_PASSWORD`

**Criterios de aceptación:**
- Llamar a cualquier endpoint sin token devuelve error `Unauthenticated`
- Después de login exitoso, el token permite llamar endpoints
- Token expirado devuelve `Unauthenticated`

---

### TAREA BE-09: Exportación de Reportes a Excel/CSV

**Prioridad:** Baja
**Módulo:** nuevo `internal/service/export/`

**Descripción:**
El equipo de Riccitelli trabaja con Excel. El sistema debe poder exportar listados de órdenes, inventario y reportes comerciales a formato XLSX.

**Pasos detallados:**

1. **Usar la dependencia ya instalada `github.com/xuri/excelize/v2`** (ya está en `go.mod`).

2. **Crear `internal/service/export/service.go`** con:
   - `ExportSaleOrders(orders []*domain.SaleOrder, customers map[string]*domain.Customer) ([]byte, error)` — genera XLSX con columnas: ID, Cliente, Estado, Mercado, Moneda, Tipo de Venta, País Destino, Cant. Items, Fecha
   - `ExportInventoryReport(report *InventoryReport) ([]byte, error)` — dos hojas: Productos (con tricapa) e Insumos Secos (con tricapa y estado de alerta)
   - `ExportProductPerformance(data []ProductPerformanceSummary) ([]byte, error)` — ranking de productos

3. **Exponer mediante HTTP REST** (no gRPC, porque gRPC no maneja bien binarios grandes):
   - Crear un servidor HTTP simple en `cmd/http/rest/server.go` que escuche en el puerto `HTTP_PORT` (default 8080)
   - `GET /export/sale-orders?from=YYYY-MM-DD&to=YYYY-MM-DD` → responde con `Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`
   - `GET /export/inventory` → exporta reporte de inventario completo
   - `GET /export/performance?from=YYYY-MM-DD&to=YYYY-MM-DD` → ranking de productos
   - Aplicar middleware de autenticación JWT también a estos endpoints

**Criterios de aceptación:**
- `GET /export/inventory` devuelve un XLSX con al menos dos hojas descargable
- El XLSX de órdenes tiene todas las columnas especificadas y es abierto correctamente por Excel/LibreOffice

---

### TAREA BE-10: Índices y Optimización del Repositorio PostgreSQL

**Prioridad:** Baja (post TAREA BE-01)
**Módulo:** `internal/infraestructure/postgres/`

**Descripción:**
Una vez implementada la DB real (BE-01), agregar índices en las columnas de búsqueda frecuente para garantizar performance en producción.

**Pasos detallados:**

1. **En `migrations.go`**, agregar los siguientes índices después de crear las tablas:
   ```sql
   CREATE INDEX idx_sale_orders_customer_id ON sale_orders(customer_id);
   CREATE INDEX idx_sale_orders_status ON sale_orders(status);
   CREATE INDEX idx_sale_orders_created_at ON sale_orders(created_at);
   CREATE INDEX idx_sale_orders_market ON sale_orders(market);
   CREATE INDEX idx_sale_order_items_order_id ON sale_order_items(sale_order_id);
   CREATE INDEX idx_production_orders_sale_order_id ON production_orders(sale_order_id);
   CREATE INDEX idx_product_movements_inventory_id ON product_movements(inventory_id);
   CREATE INDEX idx_dry_supply_movements_inventory_id ON dry_supply_movements(inventory_id);
   CREATE INDEX idx_customers_active ON customers(active);
   CREATE INDEX idx_dry_supplies_code ON dry_supplies(code);
   ```

2. **Implementar connection pooling adecuado** en `connection.go`:
   - `MaxConns: 10`
   - `MinConns: 2`
   - `MaxConnLifetime: 1h`
   - `MaxConnIdleTime: 30m`

**Criterios de aceptación:**
- Queries de filtrado por status/fecha/cliente usan índices (verificable con EXPLAIN ANALYZE)
- El pool no supera las 10 conexiones simultáneas

---

### TAREA BE-11: Historial de Precios por Producto

**Prioridad:** Baja
**Módulo:** `internal/domain/product/`, `internal/service/product/`

**Descripción:**
El sistema no guarda precios de referencia por producto. Cada item de orden lleva su precio unitario, pero no hay un catálogo de precios histórico.

**Pasos detallados:**

1. **Agregar al dominio `Product`:**
   - Slice `PriceHistory []PriceEntry`
   - Struct `PriceEntry { Currency string; Price float64; ValidFrom time.Time }`

2. **Crear `internal/service/product/set_product_price.go`:**
   - Agrega una nueva entrada al historial de precios
   - No elimina las anteriores (historial inmutable)

3. **Agregar proto `SetProductPrice`** con request `{ product_id, currency, price }`

4. **En el endpoint `GetProducts`**, incluir el precio vigente por moneda en la respuesta (el más reciente por currency).

**Criterios de aceptación:**
- Setear precio USD=15 para un producto, luego actualizarlo a USD=18
- El historial muestra ambas entradas con sus fechas
- Al listar productos, aparece el precio vigente USD=18

---

### TAREA BE-12: Tests Unitarios e Integración

**Prioridad:** Media
**Módulo:** todos los servicios

**Descripción:**
El proyecto no tiene tests. Se deben cubrir al menos los flujos críticos del negocio.

**Pasos detallados:**

1. **Tests unitarios del dominio** (`internal/domain/*/`):
   - `sale_order_test.go`: Testear transiciones de estado válidas e inválidas
   - `product_inventory_test.go`: Testear cálculo de tricapa con movimientos
   - `dry_supply_inventory_test.go`: Testear cálculo de tricapa con movimientos

2. **Tests del servicio `application-service`** (`internal/service/application-service/`):
   - `create_order_test.go`: Caso 1: stock PT suficiente → reserva directa
   - `create_order_test.go`: Caso 2: stock PT insuficiente, SV suficiente → viste + reserva
   - `create_order_test.go`: Caso 3: stock insuficiente total → error
   - Usar el repositorio in-memory como mock

3. **Tests del servicio `inventory`:**
   - `get_low_stock_alerts_test.go`: Verificar que solo aparecen insumos bajo su `ReorderPoint`
   - `convert_sv_to_pt_test.go`: Verificar que la conversión registra los movimientos correctos

4. **Ejecutar con:** `go test ./...` desde la raíz del proyecto

**Criterios de aceptación:**
- `go test ./...` pasa sin errores
- Cobertura > 70% en `internal/domain/` y `internal/service/application-service/`
