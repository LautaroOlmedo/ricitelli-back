package repository

// xlsx_seeder.go
//
// Parsers that load the five xlsx files from the data/ directory and populate
// the InMemoryRepository.
//
// File → Domain mapping:
//
//  INSUMOS SECOS.xlsx        → DrySupply + DrySupplyInventory (physical stock)
//  VESTIDO Y SV.xlsx         → Product + ProductInventory (dressed / undressed stock)
//  Insumos Comprometidos.xlsx→ DrySupplyInventory (committed / reserved movements)
//  PENDIENTES.xlsx           → Customer + SaleOrder (status CONFIRMED) + SaleOrderItem
//  REMITIDOS.xlsx            → Customer + SaleOrder (status DISPATCHED) + SaleOrderItem

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	excelize "github.com/xuri/excelize/v2"

	customerDomain "ricitelli-back/internal/domain/customer"
	drysupply "ricitelli-back/internal/domain/dry-supply"
	drysupplyinventory "ricitelli-back/internal/domain/dry-supply-inventory"
	"ricitelli-back/internal/domain/product"
	productinventory "ricitelli-back/internal/domain/product-inventory"
	saleorder "ricitelli-back/internal/domain/sale-order"
	valueobject "ricitelli-back/internal/value-object"
)

// resolveDataPath returns the first existing candidate for the data directory,
// supporting both `go run cmd/main.go` (cwd = project root → "data") and
// `cd cmd && go run main.go` (cwd = cmd/ → "../data").
func resolveDataPath() string {
	for _, candidate := range []string{"data", "../data"} {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return "data" // fallback; will produce a clear error message
}

// bomProductKeys stores product matching data during seeding.
// Populated by loadVestidoYSV, consumed by linkProductsToDrySupplies.
var bomProductKeys map[string][2]string

// SeedFromXLSX is the main entry point. basePath is the directory that
// contains the xlsx files (e.g. "data").
func (r *InMemoryRepository) SeedFromXLSX(basePath string) error {
	bomProductKeys = make(map[string][2]string)

	// Phase 1: Load catalogs (dry supplies and products)
	if err := r.loadInsumosSecos(basePath + "/INSUMOS SECOS.xlsx"); err != nil {
		return fmt.Errorf("INSUMOS SECOS: %w", err)
	}
	if err := r.loadVestidoYSV(basePath + "/VESTIDO Y SV.xlsx"); err != nil {
		return fmt.Errorf("VESTIDO Y SV: %w", err)
	}

	// Phase 2: Infer BOM from product/supply name matching
	r.linkProductsToDrySupplies()
	bomProductKeys = nil // free memory

	// Phase 3: Load transactional data
	if err := r.loadInsumosComprometidos(basePath + "/Insumos Comprometidos.xlsx"); err != nil {
		return fmt.Errorf("INSUMOS COMPROMETIDOS: %w", err)
	}
	if err := r.loadPendientes(basePath + "/PENDIENTES.xlsx"); err != nil {
		return fmt.Errorf("PENDIENTES: %w", err)
	}
	if err := r.loadRemitidos(basePath + "/REMITIDOS.xlsx"); err != nil {
		return fmt.Errorf("REMITIDOS: %w", err)
	}
	return nil
}

// =============================================================================
// Internal helpers
// =============================================================================

// cell returns row[idx] trimmed, or "" if out of bounds.
func cell(row []string, idx int) string {
	if idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

// parseQty parses a float/int string and rounds to uint64.
// Handles negative values (takes absolute value).
func parseQty(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty")
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return uint64(math.Round(math.Abs(f))), nil
}

// slugify lowercases s, replaces non-alphanumeric chars with "_" and
// collapses consecutive underscores.
func slugify(s string) string {
	var sb strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
		} else {
			sb.WriteRune('_')
		}
	}
	result := sb.String()
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}
	return strings.Trim(result, "_")
}

// cleanProductName strips the ERP " SV" stage suffix from descriptions.
func cleanProductName(desc string) string {
	s := strings.TrimSpace(desc)
	if strings.HasSuffix(s, " SV") {
		s = strings.TrimSuffix(s, " SV")
	}
	return strings.TrimSpace(s)
}

// =============================================================================
// BOM inference: link products to dry supplies by name matching
// =============================================================================
//
// The xlsx files do not contain explicit product-to-supply mappings.
// This function infers the BOM by matching product line+varietal tokens
// against dry supply names. Each matching supply gets QuantityPerUnit = 1.
//
// See docs/bom-inference-strategy.md for rationale and future improvements.

func (r *InMemoryRepository) linkProductsToDrySupplies() {
	if len(bomProductKeys) == 0 {
		return
	}

	// Pre-normalize all dry supply names once
	type supplyEntry struct {
		id         string
		normalized string
	}
	supplies := make([]supplyEntry, 0, len(r.DrySupplies))
	for _, ds := range r.DrySupplies {
		supplies = append(supplies, supplyEntry{
			id:         ds.GetID(),
			normalized: normalizeForMatch(ds.GetName()),
		})
	}

	// Count how many products share each ord1 (product line).
	// Single-varietal lines match on ord1 tokens only (no varietal needed).
	ord1Count := make(map[string]int)
	for _, keys := range bomProductKeys {
		ord1Count[keys[0]]++
	}

	linked := 0
	for i, p := range r.Products {
		keys, ok := bomProductKeys[p.GetID()]
		if !ok {
			continue
		}
		ord1, ord2 := keys[0], keys[1]

		// For single-varietal product lines, match on ord1 only.
		// For multi-varietal lines, require ord1 + ord2 tokens.
		var tokens []string
		if ord1Count[ord1] == 1 {
			tokens = bomFilterTokens(ord1)
		} else {
			tokens = bomFilterTokens(ord1 + " " + ord2)
		}
		if len(tokens) == 0 {
			continue
		}

		// Find all supplies whose name contains ALL product tokens
		var bods []valueobject.BillOfDrySupply
		for _, se := range supplies {
			if allTokensPresent(se.normalized, tokens) {
				bods = append(bods, valueobject.BillOfDrySupply{
					DrySupplyID:     se.id,
					QuantityPerUnit: 1,
				})
			}
		}

		if len(bods) > 0 {
			r.Products[i] = product.ReconstitueProduct(
				p.GetID(), p.GetName(), bods, true,
			)
			linked++
		}
	}
	fmt.Printf("[BOM inference] linked %d/%d products to dry supplies\n", linked, len(r.Products))
}

// bomFilterTokens normalizes a string and returns significant tokens,
// filtering out noise words and very short tokens.
func bomFilterTokens(s string) []string {
	raw := normalizeForMatch(s)
	words := strings.Fields(raw)

	noise := map[string]bool{
		"de": true, "la": true, "del": true, "y": true,
		"the": true, "and": true, "from": true, "is": true, "not": true,
	}
	var tokens []string
	for _, w := range words {
		if len(w) < 2 || noise[w] {
			continue
		}
		tokens = append(tokens, w)
	}
	return tokens
}

// allTokensPresent returns true if every token appears in s.
func allTokensPresent(s string, tokens []string) bool {
	for _, t := range tokens {
		if !strings.Contains(s, t) {
			return false
		}
	}
	return true
}

// normalizeForMatch lowercases, strips punctuation, and collapses whitespace.
func normalizeForMatch(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			return r
		}
		return -1
	}, s)
	return strings.Join(strings.Fields(s), " ")
}

// mapDrySupplyCategory maps the ERP Ord-2 string to the domain Category.
func mapDrySupplyCategory(ord2 string) drysupply.Category {
	switch strings.TrimSpace(ord2) {
	case "Etiquetas":
		return drysupply.CategoryLabel
	case "Cajas":
		return drysupply.CategoryBox
	case "Contraetiquetas":
		return drysupply.CategoryContraetiqueta
	case "Capsulas":
		return drysupply.CategoryCapsule
	default:
		return drysupply.CategoryOther
	}
}

// parseRemitidoDate converts "DD/MM/YYYY" to RFC3339.
func parseRemitidoDate(s string) string {
	t, err := time.Parse("02/01/2006", strings.TrimSpace(s))
	if err != nil {
		return time.Now().UTC().Format(time.RFC3339)
	}
	return t.UTC().Format(time.RFC3339)
}

// --- index helpers (avoid stale pointers after slice appends) ---

func (r *InMemoryRepository) drySupplyIndexByCode(code string) int {
	for i := range r.DrySupplies {
		if r.DrySupplies[i].GetCode() == code {
			return i
		}
	}
	return -1
}

func (r *InMemoryRepository) drySupplyInventoryIndexBySupplyID(id string) int {
	for i := range r.DrySupplyInventories {
		if r.DrySupplyInventories[i].GetDrySupplyID() == id {
			return i
		}
	}
	return -1
}

func (r *InMemoryRepository) productIndexByName(name string) int {
	lower := strings.ToLower(name)
	for i := range r.Products {
		if strings.ToLower(r.Products[i].GetName()) == lower {
			return i
		}
	}
	return -1
}

func (r *InMemoryRepository) productIndexByID(id string) int {
	for i := range r.Products {
		if r.Products[i].GetID() == id {
			return i
		}
	}
	return -1
}

func (r *InMemoryRepository) productInventoryIndex(productID, sku string) int {
	for i := range r.ProductInventory {
		if r.ProductInventory[i].GetProductID() == productID &&
			r.ProductInventory[i].GetSku() == sku {
			return i
		}
	}
	return -1
}

func (r *InMemoryRepository) customerIndexByReason(reason string) int {
	lower := strings.ToLower(reason)
	for i := range r.Customers {
		if strings.ToLower(r.Customers[i].GetSocialReason()) == lower {
			return i
		}
	}
	return -1
}

// findOrCreateProduct returns the ID of the product with the given name,
// creating one if it does not exist.
func (r *InMemoryRepository) findOrCreateProduct(name string) string {
	if idx := r.productIndexByName(name); idx != -1 {
		return r.Products[idx].GetID()
	}
	id := uuid.New().String()
	p := product.ReconstitueProduct(id, name, nil, true)
	r.Products = append(r.Products, p)
	return id
}

// findOrCreateProductByERPCode returns the ID of a product whose ID equals
// the given ERP code (used for REMITIDOS), creating a stub if needed.
func (r *InMemoryRepository) findOrCreateProductByERPCode(code string) string {
	if idx := r.productIndexByID(code); idx != -1 {
		return code
	}
	p := product.ReconstitueProduct(code, code, nil, true)
	r.Products = append(r.Products, p)
	return code
}

// getOrCreateCustomer returns the ID of the customer with the given social
// reason, creating one with defaults if it does not exist.
func (r *InMemoryRepository) getOrCreateCustomer(reason, market string) string {
	if idx := r.customerIndexByReason(reason); idx != -1 {
		return r.Customers[idx].GetID()
	}
	marketType := customerDomain.MarketTypeInternal
	group := customerDomain.GroupDistributor
	if strings.EqualFold(market, "Mercado Externo") {
		marketType = customerDomain.MarketTypeExternal
		group = customerDomain.GroupExportAgent
	}
	c, err := customerDomain.NewCustomer(customerDomain.NewCustomerParams{
		SocialReason: reason,
		MarketType:   marketType,
		Group:        group,
	})
	if err != nil {
		return uuid.New().String()
	}
	r.Customers = append(r.Customers, c)
	return c.GetID()
}

// =============================================================================
// INSUMOS SECOS.xlsx
// =============================================================================
//
// Columns (0-indexed):
//  0 Depósito           (always "Insumos Secos")
//  1 Depósito Cód.      (always "IS")
//  2 Artículo Ord.1     (always "Ins Fraccionamientos")
//  3 Artículo Ord.2     category: Etiquetas | Cajas | Contraetiquetas | Capsulas
//  4 Artículo Cód. Gen. ERP code (e.g. "IF035    ")
//  5 Artículo Desc.Gen. human-readable name
//  6 Stock Cant. Real   physical stock quantity
//
// → DrySupply + DrySupplyInventory (with AddStock)

func (r *InMemoryRepository) loadInsumosSecos(path string) error {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return err
	}
	defer f.Close()

	rows, err := f.GetRows("Hoja1")
	if err != nil {
		return err
	}

	seen := make(map[string]bool)

	for i, row := range rows {
		if i == 0 {
			continue // header
		}
		code := cell(row, 4)
		name := cell(row, 5)
		if code == "" || name == "" {
			continue
		}
		if seen[code] {
			continue
		}
		seen[code] = true

		category := mapDrySupplyCategory(cell(row, 3))
		id := uuid.New().String()

		ds, err := drysupply.NewDrySupplyWithID(id, code, name, category, "UNIT")
		if err != nil {
			continue
		}
		r.DrySupplies = append(r.DrySupplies, ds)

		inv, _ := drysupplyinventory.NewDrySupplyInventory(id)
		if qty, qerr := parseQty(cell(row, 6)); qerr == nil && qty > 0 {
			_ = inv.AddStock(qty, "xlsx-initial-stock")
		}
		r.DrySupplyInventories = append(r.DrySupplyInventories, inv)
	}
	return nil
}

// =============================================================================
// VESTIDO Y SV.xlsx
// =============================================================================
//
// Columns (0-indexed):
//  0 Depósito              (always "Frigorifico")
//  1 Depósito Cód.         (always "PT" = Producto Terminado / Dressed)
//  2 Artículo Ord.1        product line  (e.g. "Hey !")
//  3 Artículo Ord.2        varietal      (e.g. "Malbec")
//  4 Artículo Desc.Gen.    full ERP description (e.g. "Hey Malbec! Malbec SV")
//  5 Artículo Elem.1       vintage year  (e.g. "2024")
//  6 Artículo Elem.2       bottle format (e.g. "Bot 750cc")
//  7 Stock Cant. Real UM1  quantity in bottles
//  8 Stock Cant. Real UM2  quantity in cases
//  9 Lot number            (e.g. "L-040825-5-21")
//
// → Product (keyed by ord1+"|"+ord2) + ProductInventory
//   All rows represent DRESSED (PT) stock:
//     AddUndressed(qty) → ConvertSVtoPT(qty, lot)

func (r *InMemoryRepository) loadVestidoYSV(path string) error {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return err
	}
	defer f.Close()

	rows, err := f.GetRows("Hoja1")
	if err != nil {
		return err
	}

	// productKey ("ord1|ord2") → productID
	productByKey := make(map[string]string)

	for i, row := range rows {
		if i == 0 {
			continue // header
		}

		ord1 := cell(row, 2)
		ord2 := cell(row, 3)
		desc := cleanProductName(cell(row, 4))
		vintage := cell(row, 5)
		format := cell(row, 6)
		lotNumber := cell(row, 9)

		if desc == "" || vintage == "" {
			continue
		}

		qty, err := parseQty(cell(row, 7))
		if err != nil || qty == 0 {
			continue
		}

		// --- Product ---
		productKey := ord1 + "|" + ord2
		productID, exists := productByKey[productKey]
		if !exists {
			productID = uuid.New().String()
			productByKey[productKey] = productID
			p := product.ReconstitueProduct(productID, desc, nil, true)
			r.Products = append(r.Products, p)
			if bomProductKeys != nil {
				bomProductKeys[productID] = [2]string{ord1, ord2}
			}
		}

		// --- SKU: slugified ord1_ord2_vintage_format ---
		sku := fmt.Sprintf("%s_%s_%s_%s",
			slugify(ord1), slugify(ord2), vintage, slugify(format))

		// --- ProductInventory ---
		invIdx := r.productInventoryIndex(productID, sku)
		if invIdx == -1 {
			newInv, err := productinventory.NewProductInventory(productID, sku)
			if err != nil {
				continue
			}
			r.ProductInventory = append(r.ProductInventory, newInv)
			invIdx = len(r.ProductInventory) - 1
		}

		// Model as: first produce undressed, then dress with lot number.
		if err := r.ProductInventory[invIdx].AddUndressed("xlsx-initial-production", qty); err != nil {
			continue
		}
		_ = r.ProductInventory[invIdx].ConvertSVtoPT("xlsx-initial-dressing", qty, lotNumber)
	}
	return nil
}

// =============================================================================
// Insumos Comprometidos.xlsx
// =============================================================================
//
// Row 0 (index 0): auxiliary header
// Row 1 (index 1): real header
// Data starts at row 2.
//
// Columns (0-indexed):
//  0 Tipo Comp.   always "OP" (Orden de Producción)
//  1 Comprobante  production order number (integer)
//  2 Artículo     ERP dry supply code (e.g. "IF1156       ")
//  3 Desc.        human-readable description
//  4 Qty committed (negative integer, e.g. -594)
//
// → DrySupplyInventory.Commit() for each OP/code pair.

func (r *InMemoryRepository) loadInsumosComprometidos(path string) error {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return err
	}
	defer f.Close()

	rows, err := f.GetRows("Hoja1")
	if err != nil {
		return err
	}

	for i, row := range rows {
		if i < 2 {
			continue // skip both header rows
		}
		if cell(row, 0) != "OP" {
			continue
		}

		comprobante := cell(row, 1)
		code := cell(row, 2)
		qtyStr := cell(row, 4)

		if code == "" || qtyStr == "" {
			continue
		}

		qty, err := parseQty(qtyStr)
		if err != nil || qty == 0 {
			continue
		}

		dsIdx := r.drySupplyIndexByCode(code)
		if dsIdx == -1 {
			continue // dry supply not found in loaded catalog
		}
		invIdx := r.drySupplyInventoryIndexBySupplyID(r.DrySupplies[dsIdx].GetID())
		if invIdx == -1 {
			continue
		}

		ref := "OP-" + comprobante
		// Errors (e.g. insufficient stock) are silently skipped to keep seeding
		// idempotent even if xlsx snapshots are from different dates.
		_ = r.DrySupplyInventories[invIdx].Commit(qty, ref)
	}
	return nil
}

// =============================================================================
// PENDIENTES.xlsx
// =============================================================================
//
// Columns (0-indexed):
//  0 Comp. Cliente Ord.1     market: "Mercado Externo" | "Mercado Interno"
//  1 Comp. Cliente Razón Social  customer name
//  2 Comp. Ppal.             order number (e.g. "NP  0005-00002267")
//  3 Artículo Ord.1          product line
//  4 Artículo Ord.2          varietal
//  5 Artículo Desc.Gen.      full product description
//  6 Artículo Elem.1         vintage
//  7 Artículo Elem.2         package size
//  8 Ítem Cant. Pend. UM1    pending quantity (bottles)
//  9 Ítem Cant. Pend. UM2    pending quantity (cases)
//
// → Customer + SaleOrder (status CONFIRMED) + SaleOrderItem

type pendienteAcc struct {
	customerReason string
	market         string
	items          []valueobject.SaleOrderItem
}

func (r *InMemoryRepository) loadPendientes(path string) error {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return err
	}
	defer f.Close()

	rows, err := f.GetRows("Hoja1")
	if err != nil {
		return err
	}

	// orderNum → accumulated data (preserving insertion order)
	orderMap := make(map[string]*pendienteAcc)
	orderSeq := make([]string, 0)

	for i, row := range rows {
		if i == 0 {
			continue // header
		}
		market := cell(row, 0)
		reason := cell(row, 1)
		orderNum := cell(row, 2)
		productDesc := cell(row, 5)

		if orderNum == "" || productDesc == "" || reason == "" {
			continue
		}

		qty, err := parseQty(cell(row, 8))
		if err != nil || qty == 0 {
			continue
		}

		productID := r.findOrCreateProduct(productDesc)

		if _, ok := orderMap[orderNum]; !ok {
			orderMap[orderNum] = &pendienteAcc{
				customerReason: reason,
				market:         market,
				items:          []valueobject.SaleOrderItem{},
			}
			orderSeq = append(orderSeq, orderNum)
		}
		orderMap[orderNum].items = append(orderMap[orderNum].items, valueobject.SaleOrderItem{
			ProductID: productID,
			Quantity:  qty,
			UnitPrice: 0,
		})
	}

	for _, orderNum := range orderSeq {
		acc := orderMap[orderNum]
		if len(acc.items) == 0 {
			continue
		}

		customerID := r.getOrCreateCustomer(acc.customerReason, acc.market)

		market := saleorder.MarketDomestic
		if strings.EqualFold(acc.market, "Mercado Externo") {
			market = saleorder.MarketExport
		}

		so, err := saleorder.NewSaleOrder(saleorder.NewSaleOrderParams{
			CustomerID: customerID,
			Items:      acc.items,
			Currency:   saleorder.CurrencyARS,
			Market:     market,
			SaleType:   saleorder.SaleTypeRegular,
		})
		if err != nil {
			continue
		}
		// Pending orders are confirmed (placed and awaiting dispatch).
		_ = so.UpdateStatus(saleorder.StatusConfirmed)
		r.SaleOrders = append(r.SaleOrders, so)
	}
	return nil
}

// =============================================================================
// REMITIDOS.xlsx
// =============================================================================
//
// Columns (0-indexed):
//  0 Tipo Comp. Cód.       "RT"
//  1 Pto. Vta. Cód.        "0005"
//  2 Ítem Tipo Cód.        "A"
//  3 Comp. Ppal.           remito number (e.g. "RT R 0005-00003225")
//  4 Comp. F. Emisión      date "DD/MM/YYYY"
//  5 Comp. Cliente Razón   customer name
//  6 Comp. Mensaje         notes / PO reference
//  7 Ítem Artículo Cód.Gen ERP product code (e.g. "HEYM     ")
//  8 Ítem Artículo Elem.1  vintage
//  9 Ítem Artículo Elem.2  package code
// 10 Ítem Cant. Rt. UM1    dispatched quantity (bottles)
// 11 Ítem Cant. Rt. UM2    dispatched quantity (cases)
//
// → Customer + SaleOrder (status DISPATCHED) + SaleOrderItem

type remitidoAcc struct {
	customerReason string
	date           string
	items          []valueobject.SaleOrderItem
}

func (r *InMemoryRepository) loadRemitidos(path string) error {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return err
	}
	defer f.Close()

	rows, err := f.GetRows("Hoja1")
	if err != nil {
		return err
	}

	orderMap := make(map[string]*remitidoAcc)
	orderSeq := make([]string, 0)

	for i, row := range rows {
		if i == 0 {
			continue // header
		}
		remito := cell(row, 3)
		dateStr := cell(row, 4)
		reason := cell(row, 5)
		productCode := cell(row, 7)

		if remito == "" || productCode == "" || reason == "" {
			continue
		}

		qty, err := parseQty(cell(row, 10))
		if err != nil || qty == 0 {
			continue
		}

		productID := r.findOrCreateProductByERPCode(productCode)

		if _, ok := orderMap[remito]; !ok {
			orderMap[remito] = &remitidoAcc{
				customerReason: reason,
				date:           parseRemitidoDate(dateStr),
				items:          []valueobject.SaleOrderItem{},
			}
			orderSeq = append(orderSeq, remito)
		}
		orderMap[remito].items = append(orderMap[remito].items, valueobject.SaleOrderItem{
			ProductID: productID,
			Quantity:  qty,
			UnitPrice: 0,
		})
	}

	for _, remito := range orderSeq {
		acc := orderMap[remito]
		if len(acc.items) == 0 {
			continue
		}

		// Infer market from customer: external customers are exported orders.
		// Default to DOMESTIC; can be refined with additional data.
		customerID := r.getOrCreateCustomer(acc.customerReason, "Mercado Interno")

		so, err := saleorder.NewSaleOrder(saleorder.NewSaleOrderParams{
			CustomerID: customerID,
			Items:      acc.items,
			Currency:   saleorder.CurrencyARS,
			Market:     saleorder.MarketDomestic,
			SaleType:   saleorder.SaleTypeRegular,
		})
		if err != nil {
			continue
		}
		// Advance through the full pipeline to DISPATCHED.
		_ = so.UpdateStatus(saleorder.StatusConfirmed)
		_ = so.UpdateStatus(saleorder.StatusInvoiced)
		_ = so.UpdateStatus(saleorder.StatusDispatched)
		r.SaleOrders = append(r.SaleOrders, so)
	}
	return nil
}
