package httpserver

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type commerceHandler struct{ db *pgxpool.Pool }

func newCommerceHandler(db *pgxpool.Pool) *commerceHandler { return &commerceHandler{db: db} }

type productDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	PriceMinor  int64     `json:"price_minor"`
	Currency    string    `json:"currency"`
	IsActive    bool      `json:"is_active"`
}

func (h *commerceHandler) publicProducts(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSpace(r.PathValue("slug"))
	rows, err := h.db.Query(r.Context(), `
		SELECT p.id, p.name, p.description, p.price_minor, p.currency, p.is_active
		FROM products p JOIN organizations o ON o.id = p.organization_id
		WHERE o.slug = $1 AND p.is_active = TRUE
		ORDER BY p.created_at DESC LIMIT 100
	`, slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "catalog_failed", "Could not load catalog")
		return
	}
	defer rows.Close()
	items := make([]productDTO, 0)
	for rows.Next() {
		var p productDTO
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.PriceMinor, &p.Currency, &p.IsActive); err != nil {
			writeError(w, http.StatusInternalServerError, "catalog_failed", "Could not load catalog")
			return
		}
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "catalog_failed", "Could not load catalog")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *commerceHandler) organizationProducts(w http.ResponseWriter, r *http.Request) {
	orgID, ok := h.authorizeOrganization(w, r, r.Method != http.MethodGet)
	if !ok {
		return
	}
	if r.Method == http.MethodPatch {
		productID, err := uuid.Parse(r.PathValue("productID"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_product", "Product ID is invalid")
			return
		}
		var in struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			PriceMinor  int64  `json:"price_minor"`
			IsActive    bool   `json:"is_active"`
		}
		if !decodeJSON(w, r, &in) {
			return
		}
		in.Name = strings.TrimSpace(in.Name)
		in.Description = strings.TrimSpace(in.Description)
		if in.Name == "" || len(in.Name) > 200 || len(in.Description) > 4000 || in.PriceMinor < 0 {
			writeError(w, http.StatusBadRequest, "invalid_product", "Name, description and non-negative price are required")
			return
		}
		var p productDTO
		err = h.db.QueryRow(r.Context(), `UPDATE products SET name=$3,description=$4,price_minor=$5,is_active=$6,updated_at=now() WHERE organization_id=$1 AND id=$2 RETURNING id,name,description,price_minor,currency,is_active`, orgID, productID, in.Name, in.Description, in.PriceMinor, in.IsActive).Scan(&p.ID, &p.Name, &p.Description, &p.PriceMinor, &p.Currency, &p.IsActive)
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "product_not_found", "Product was not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "product_update_failed", "Could not update product")
			return
		}
		writeJSON(w, http.StatusOK, p)
		return
	}
	if r.Method == http.MethodPost {
		var in struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			PriceMinor  int64  `json:"price_minor"`
		}
		if !decodeJSON(w, r, &in) {
			return
		}
		in.Name = strings.TrimSpace(in.Name)
		in.Description = strings.TrimSpace(in.Description)
		if len(in.Name) == 0 || len(in.Name) > 200 || len(in.Description) > 4000 || in.PriceMinor < 0 {
			writeError(w, http.StatusBadRequest, "invalid_product", "Name and a non-negative price are required")
			return
		}
		var p productDTO
		err := h.db.QueryRow(r.Context(), `
            INSERT INTO products(organization_id, name, description, price_minor, currency)
            SELECT $1, $2, $3, $4, default_currency FROM organizations WHERE id = $1
            RETURNING id,name,description,price_minor,currency,is_active
        `, orgID, in.Name, in.Description, in.PriceMinor).Scan(&p.ID, &p.Name, &p.Description, &p.PriceMinor, &p.Currency, &p.IsActive)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "product_create_failed", "Could not create product")
			return
		}
		writeJSON(w, http.StatusCreated, p)
		return
	}
	rows, err := h.db.Query(r.Context(), `SELECT id,name,description,price_minor,currency,is_active FROM products WHERE organization_id=$1 ORDER BY created_at DESC LIMIT 200`, orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "products_failed", "Could not load products")
		return
	}
	defer rows.Close()
	items := make([]productDTO, 0)
	for rows.Next() {
		var p productDTO
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.PriceMinor, &p.Currency, &p.IsActive); err != nil {
			writeError(w, http.StatusInternalServerError, "products_failed", "Could not load products")
			return
		}
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "products_failed", "Could not load products")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *commerceHandler) createOrder(w http.ResponseWriter, r *http.Request) {
	var in struct {
		CustomerName string `json:"customer_name"`
		Phone        string `json:"phone"`
		Email        string `json:"email"`
		Items        []struct {
			ProductID string `json:"product_id"`
			Quantity  int    `json:"quantity"`
		} `json:"items"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	in.CustomerName = strings.TrimSpace(in.CustomerName)
	in.Phone = strings.TrimSpace(in.Phone)
	in.Email = strings.TrimSpace(in.Email)
	if in.CustomerName == "" || len(in.CustomerName) > 160 || in.Phone == "" || len(in.Phone) > 40 || len(in.Email) > 320 || len(in.Items) == 0 || len(in.Items) > 50 {
		writeError(w, http.StatusBadRequest, "invalid_order", "Name, phone and 1–50 order items are required")
		return
	}
	orgSlug := strings.TrimSpace(r.PathValue("slug"))
	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if len(key) > 100 {
		writeError(w, http.StatusBadRequest, "invalid_idempotency_key", "Idempotency-Key is too long")
		return
	}
	if key == "" {
		key = uuid.NewString()
	}
	tx, err := h.db.Begin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "order_failed", "Could not start order")
		return
	}
	defer tx.Rollback(r.Context())
	var orgID uuid.UUID
	var currency string
	if err := tx.QueryRow(r.Context(), `SELECT id,default_currency FROM organizations WHERE slug=$1`, orgSlug).Scan(&orgID, &currency); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "store_not_found", "Store was not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "order_failed", "Could not load store")
		return
	}
	var existing struct {
		ID       uuid.UUID
		Status   string
		Total    int64
		Currency string
	}
	existingErr := tx.QueryRow(r.Context(), `SELECT id,status,total_minor,currency FROM orders WHERE organization_id=$1 AND idempotency_key=$2`, orgID, key).Scan(&existing.ID, &existing.Status, &existing.Total, &existing.Currency)
	if existingErr == nil {
		writeJSON(w, http.StatusOK, map[string]any{"order_id": existing.ID, "status": existing.Status, "total_minor": existing.Total, "currency": existing.Currency})
		return
	}
	if !errors.Is(existingErr, pgx.ErrNoRows) {
		writeError(w, http.StatusInternalServerError, "order_failed", "Could not check order idempotency")
		return
	}
	var customerID uuid.UUID
	if err := tx.QueryRow(r.Context(), `INSERT INTO customers(organization_id,name,email,phone) VALUES($1,$2,NULLIF($3,''),$4) RETURNING id`, orgID, in.CustomerName, in.Email, in.Phone).Scan(&customerID); err != nil {
		writeError(w, http.StatusInternalServerError, "order_failed", "Could not save customer")
		return
	}
	type line struct {
		id       uuid.UUID
		name     string
		price    int64
		currency string
		qty      int
	}
	lines := make([]line, 0, len(in.Items))
	total := int64(0)
	for _, item := range in.Items {
		id, parseErr := uuid.Parse(item.ProductID)
		if parseErr != nil || item.Quantity < 1 || item.Quantity > 100 {
			writeError(w, http.StatusBadRequest, "invalid_order_item", "Each item needs a valid product ID and quantity from 1 to 100")
			return
		}
		var ln line
		ln.id = id
		ln.qty = item.Quantity
		if err := tx.QueryRow(r.Context(), `SELECT name,price_minor,currency FROM products WHERE organization_id=$1 AND id=$2 AND is_active=TRUE FOR SHARE`, orgID, id).Scan(&ln.name, &ln.price, &ln.currency); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				writeError(w, http.StatusBadRequest, "product_unavailable", "One of the selected products is unavailable")
				return
			}
			writeError(w, http.StatusInternalServerError, "order_failed", "Could not validate products")
			return
		}
		if ln.currency != currency {
			writeError(w, http.StatusBadRequest, "currency_mismatch", "Products in one order must use the store currency")
			return
		}
		if ln.price > 0 && int64(ln.qty) > (int64(^uint64(0)>>1)-total)/ln.price {
			writeError(w, http.StatusBadRequest, "order_total_too_large", "Order total is too large")
			return
		}
		total += ln.price * int64(ln.qty)
		lines = append(lines, ln)
	}
	var orderID uuid.UUID
	insertErr := tx.QueryRow(r.Context(), `INSERT INTO orders(organization_id,customer_id,status,total_minor,currency,idempotency_key) VALUES($1,$2,'pending',$3,$4,$5) ON CONFLICT (organization_id,idempotency_key) DO NOTHING RETURNING id`, orgID, customerID, total, currency, key).Scan(&orderID)
	if errors.Is(insertErr, pgx.ErrNoRows) {
		if err := tx.QueryRow(r.Context(), `SELECT id,status,total_minor,currency FROM orders WHERE organization_id=$1 AND idempotency_key=$2`, orgID, key).Scan(&existing.ID, &existing.Status, &existing.Total, &existing.Currency); err != nil {
			writeError(w, http.StatusConflict, "duplicate_order", "The order was submitted concurrently; retry with the same idempotency key")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"order_id": existing.ID, "status": existing.Status, "total_minor": existing.Total, "currency": existing.Currency})
		return
	}
	if insertErr != nil {
		writeError(w, http.StatusInternalServerError, "order_failed", "Could not create order")
		return
	}
	for _, ln := range lines {
		if _, err := tx.Exec(r.Context(), `INSERT INTO order_items(organization_id,order_id,product_id,product_name_snapshot,quantity,unit_price_minor,currency) VALUES($1,$2,$3,$4,$5,$6,$7)`, orgID, orderID, ln.id, ln.name, ln.qty, ln.price, currency); err != nil {
			writeError(w, http.StatusInternalServerError, "order_failed", "Could not save order items")
			return
		}
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "order_failed", "Could not complete order")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"order_id": orderID, "status": "pending", "total_minor": total, "currency": currency})
}

func (h *commerceHandler) organizationOrders(w http.ResponseWriter, r *http.Request) {
	orgID, ok := h.authorizeOrganization(w, r, false)
	if !ok {
		return
	}
	rows, err := h.db.Query(r.Context(), `
        SELECT o.id,o.status,o.total_minor,o.currency,o.created_at,
               COALESCE(c.name,''),COALESCE(c.phone,''),COALESCE(c.email,'')
        FROM orders o
        LEFT JOIN customers c ON c.organization_id=o.organization_id AND c.id=o.customer_id
        WHERE o.organization_id=$1
        ORDER BY o.created_at DESC LIMIT 100
    `, orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "orders_failed", "Could not load orders")
		return
	}
	defer rows.Close()
	type order struct {
		ID           uuid.UUID `json:"id"`
		Status       string    `json:"status"`
		TotalMinor   int64     `json:"total_minor"`
		Currency     string    `json:"currency"`
		CreatedAt    time.Time `json:"created_at"`
		CustomerName string    `json:"customer_name"`
		Phone        string    `json:"phone"`
		Email        string    `json:"email"`
	}
	items := make([]order, 0)
	for rows.Next() {
		var o order
		if err := rows.Scan(&o.ID, &o.Status, &o.TotalMinor, &o.Currency, &o.CreatedAt, &o.CustomerName, &o.Phone, &o.Email); err != nil {
			writeError(w, http.StatusInternalServerError, "orders_failed", "Could not load orders")
			return
		}
		items = append(items, o)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "orders_failed", "Could not load orders")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *commerceHandler) updateOrderStatus(w http.ResponseWriter, r *http.Request) {
	orgID, ok := h.authorizeOrganization(w, r, true)
	if !ok {
		return
	}
	orderID, err := uuid.Parse(r.PathValue("orderID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_order", "Order ID is invalid")
		return
	}
	var in struct {
		Status string `json:"status"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.Status != "confirmed" && in.Status != "completed" && in.Status != "cancelled" {
		writeError(w, http.StatusBadRequest, "invalid_status", "Allowed statuses: confirmed, completed, cancelled")
		return
	}
	tx, err := h.db.Begin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "order_update_failed", "Could not start order update")
		return
	}
	defer tx.Rollback(r.Context())
	var current string
	err = tx.QueryRow(r.Context(), `SELECT status FROM orders WHERE organization_id=$1 AND id=$2 FOR UPDATE`, orgID, orderID).Scan(&current)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "order_not_found", "Order was not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "order_update_failed", "Could not load order")
		return
	}
	if !validOrderTransition(current, in.Status) {
		writeError(w, http.StatusConflict, "invalid_transition", "This order status cannot be changed to the requested status")
		return
	}
	if _, err = tx.Exec(r.Context(), `UPDATE orders SET status=$3,updated_at=now() WHERE organization_id=$1 AND id=$2`, orgID, orderID, in.Status); err != nil {
		writeError(w, http.StatusInternalServerError, "order_update_failed", "Could not update order")
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "order_update_failed", "Could not save order status")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": orderID, "status": in.Status})
}

func (h *commerceHandler) authorizeOrganization(w http.ResponseWriter, r *http.Request, write bool) (uuid.UUID, bool) {
	user, ok := userFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required")
		return uuid.Nil, false
	}
	orgID, err := uuid.Parse(r.PathValue("organizationID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_organization", "Organization ID is invalid")
		return uuid.Nil, false
	}
	var role string
	if err := h.db.QueryRow(r.Context(), `SELECT role FROM organization_members WHERE organization_id=$1 AND user_id=$2`, orgID, user.ID).Scan(&role); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "organization_not_found", "Organization was not found")
			return uuid.Nil, false
		}
		writeError(w, http.StatusInternalServerError, "authorization_failed", "Could not verify organization access")
		return uuid.Nil, false
	}
	if write && role != "owner" && role != "admin" && role != "manager" {
		writeError(w, http.StatusForbidden, "insufficient_role", "Your role cannot manage products")
		return uuid.Nil, false
	}
	return orgID, true
}

func validOrderTransition(current, next string) bool {
	return (current == "pending" && (next == "confirmed" || next == "cancelled")) ||
		(current == "confirmed" && (next == "completed" || next == "cancelled"))
}
