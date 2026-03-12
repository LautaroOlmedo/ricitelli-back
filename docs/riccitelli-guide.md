1. Propuesta de Valor (VPS) sugerida
    "Trazabilidad 360° en Tiempo Real": Conectar el inventario de insumos secos y vinos "Sin Vestir" con las ventas y despachos, eliminando el desfasaje entre lo que hay en depósito, lo que está comprometido en producción y lo que falta enviar.
    "Gestión Comercial Global": Centralizar la complejidad de un mercado dual (Mercado Interno y Mercado Externo), automatizando el seguimiento de pedidos en múltiples monedas, destinos y formatos de empaque.

2. Necesidades y Puntos de Dolor Detectados en los Documentos

    Fragmentación del Stock de Insumos: La bodega tiene una cantidad masiva de insumos secos (cajas, etiquetas, contraetiquetas, cápsulas) altamente específicos por mercado y cliente (ej. etiquetas exclusivas para Japón, "Skurnik", o "TWC"). Actualmente, deben cruzar manualmente el archivo de "Insumos Secos" con el de "Insumos Comprometidos" (Órdenes de Producción o OP) para saber el stock real vs. el stock disponible.
    Complejidad del Producto "Sin Vestir" (SV): El vino embotellado no siempre está listo para la venta. La bodega almacena botellas "Sin Vestir" (SV) y luego las transforma en "Producto Terminado" (PT) según el destino. Gestionar esta transición es crítico.
    Trazabilidad de Lotes: El archivo de vinos vestidos y SV muestra números de lote específicos (ej. "L-081124-38-11"). Llevar este control en planillas es propenso a errores ante un "recall" o control de calidad.
    Cuellos de botella en Pedidos Pendientes: Existe un volumen importante de pedidos pendientes de entrega tanto para el mercado interno como externo. Se necesita visualizar rápidamente si un pedido no se despacha por falta de vino, o por falta de un insumo específico (ej. una contraetiqueta para Brasil).
    Salidas No Comerciales: Los remitos no solo reflejan ventas. Muestran una gran cantidad de salidas por "Muestras de Aduana", "Obsequio Bodega", "Consumidor Final - Empleados" y "Muestras Comerciales" para prensa o degustaciones. Esto requiere un módulo especial para no distorsionar los ingresos por ventas.

3. Módulos y Funcionalidades Clave para el MVP
En base a estos documentos, tu plataforma debería contar con los siguientes módulos:
A. Módulo de Inventario y Producción Inteligente

    Árbol de Producto (Bill of Materials - BOM): Vincular automáticamente un vino terminado con sus insumos. Si se programa despachar 100 cajas de Hey Malbec!, el sistema debe descontar o "comprometer" automáticamente las botellas SV, las etiquetas, cápsulas y cajas de cartón correspondientes.
    Stock Tricapa: Mostrar siempre tres métricas de inventario: Stock Físico Real, Stock Comprometido (en Órdenes de Producción), y Stock Disponible.
    Gestión SV a PT: Un flujo de trabajo donde el operario convierta el inventario "Sin Vestir" a Producto Terminado, asignando automáticamente el número de lote.

B. Módulo Comercial y Pipeline de Ventas

    Multimoneda y Multimercado: Soporte nativo para operar en Pesos Argentinos (),Doˊlares(us) y Dólares Canadienses (CAD), y diferenciar Mercado Interno de Mercado Externo.
    Pipeline de Pedidos: Un tablero tipo Kanban que trace el ciclo de vida de la venta: Nota de Pedido (NP) -> Pendiente -> Factura (FCE) -> Remito (RT) despachado.

4. Sugerencia de Dashboards y Métricas (Tiempo Real)
Para la gerencia comercial y operativa de Riccitelli, te sugiero incluir los siguientes tableros visuales:
Dashboard de Operaciones y Stock:

    Alerta de Quiebre de Stock (Punto de Pedido): Gráficos de alerta cuando insumos críticos (como cajas o cápsulas "Hey") están por debajo del nivel de los pedidos pendientes.
    Ratio SV vs PT: Un gráfico de torta que muestre cuánto vino está etiquetado y listo para vender, versus cuánto está "Sin Vestir".

Dashboard Comercial y de Rentabilidad:

    Ventas por Mercado / País: Mapa de calor o gráficos de barras mostrando el volumen facturado. En sus documentos, se ve una gran distribución global (Reino Unido, Brasil, Suiza, Grecia, Canadá, Australia, Singapur, etc.).
    Rendimiento por Línea de Vino: Ranking de los productos con mayor rotación y volumen de ventas (destacan fuertemente líneas como Hey Malbec!, The Party, Kung Fu y Old Vines From Patagonia).
    Control de Fugas / Salidas Internas: Un gráfico que cuantifique el costo de las botellas destinadas a Muestras Comerciales, Aduana o Consumo Interno para llevar un control del presupuesto de marketing.

Información adicional valiosa:
Al diseñar la arquitectura, asegúrate de utilizar SKUs o Códigos de Artículo estandarizados y relacionales. En los documentos provistos, la bodega usa códigos como "IF1156" para insumos y códigos como "HEYM", "KUNGM" o "TDCL" para productos terminados. Tu plataforma debe conectar estos mundos (Insumo -> SV -> Producto Terminado -> Venta) de manera fluida. Si logras digitalizar y automatizar esta relación, el software ahorrará cientos de horas mensuales de carga administrativa y prevendrá errores logísticos internacionales.