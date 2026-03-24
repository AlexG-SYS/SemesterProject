package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Order struct {
	ID               int64       `json:"order_id"`
	CustomerID       int64       `json:"customer_id"`
	ShippingMethodID int64       `json:"shipping_method_id,omitempty"`
	Status           string      `json:"status"`
	ShippingFee      float64     `json:"shipping_fee"`
	OrderItems       []OrderItem `json:"order_items,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
}

type OrderItem struct {
	ID             int64   `json:"order_item_id"`
	OrderID        int64   `json:"order_id"`
	VariantID      int64   `json:"variant_id"`
	Quantity       int     `json:"quantity"`
	PriceAtReserve float64 `json:"price_at_reserve"`
	CostAtReserve  float64 `json:"cost_at_reserve"`
}

type OrderModel struct {
	DB *sql.DB
}

func (m OrderModel) Insert(ctx context.Context, order *Order, locationID int64) error {
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Fetch Shipping Fee from shipping_methods table
	shippingQuery := `SELECT base_rate FROM shipping_methods WHERE method_id = $1`
	err = tx.QueryRowContext(ctx, shippingQuery, order.ShippingMethodID).Scan(&order.ShippingFee)
	if err != nil {
		return fmt.Errorf("invalid shipping method: %w", err)
	}

	// 2. Insert Order Header
	query := `
        INSERT INTO orders (customer_id, shipping_method_id, shipping_fee, status)
        VALUES ($1, $2, $3, $4)
        RETURNING order_id, created_at`

	err = tx.QueryRowContext(ctx, query, order.CustomerID, order.ShippingMethodID, order.ShippingFee, order.Status).Scan(&order.ID, &order.CreatedAt)
	if err != nil {
		return err
	}

	// 3. Process each item: Fetch Price -> Update Inventory -> Insert Order Item
	for i := range order.OrderItems {
		// A. Fetch current Price and Cost from product_variants
		priceQuery := `SELECT v.selling_price, v.cost_price 
			FROM product_variants v
			JOIN inventory i ON v.variant_id = i.variant_id
			WHERE v.variant_id = $1 AND i.location_id = $2
			FOR UPDATE OF i`
		err := tx.QueryRowContext(ctx, priceQuery, order.OrderItems[i].VariantID, locationID).Scan(
			&order.OrderItems[i].PriceAtReserve,
			&order.OrderItems[i].CostAtReserve,
		)
		if err != nil {
			return fmt.Errorf("variant %d not found: %w", order.OrderItems[i].VariantID, err)
		}

		// B. Update Inventory (Reserved Stock)
		// This query assumes you have a 'quantity' column in inventory.
		// We check if (quantity - stock_reserved) >= requested amount.
		inventoryQuery := `
			UPDATE inventory 
			SET stock_reserved = stock_reserved + $1 
			WHERE variant_id = $2 AND location_id = $3 
			AND (stock_on_hand - stock_reserved) >= $1`

		result, err := tx.ExecContext(ctx, inventoryQuery, order.OrderItems[i].Quantity, order.OrderItems[i].VariantID, locationID)
		if err != nil {
			return err
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			return fmt.Errorf("insufficient stock for variant %d", order.OrderItems[i].VariantID)
		}

		// C. Insert the Order Item
		itemQuery := `
			INSERT INTO order_items (order_id, variant_id, quantity, price_at_reserve, cost_at_reserve)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING order_item_id`

		order.OrderItems[i].OrderID = order.ID
		err = tx.QueryRowContext(ctx, itemQuery,
			order.OrderItems[i].OrderID,
			order.OrderItems[i].VariantID,
			order.OrderItems[i].Quantity,
			order.OrderItems[i].PriceAtReserve,
			order.OrderItems[i].CostAtReserve,
		).Scan(&order.OrderItems[i].ID)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (m OrderModel) GetByID(ctx context.Context, id int64) (*Order, error) {
	// 1. Fetch Order Header
	query := `
        SELECT o.order_id, o.customer_id, o.shipping_method_id, o.status, o.shipping_fee, o.created_at
        FROM orders o
        WHERE o.order_id = $1`

	var order Order
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.CustomerID,
		&order.ShippingMethodID,
		&order.Status,
		&order.ShippingFee,
		&order.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// 2. Fetch Order Items with Product Names
	// We join product_variants and products to get the name
	itemQuery := `
        SELECT oi.order_item_id, oi.variant_id, oi.quantity, oi.price_at_reserve, p.name
        FROM order_items oi
        JOIN product_variants v ON oi.variant_id = v.variant_id
        JOIN products p ON v.product_id = p.product_id
        WHERE oi.order_id = $1`

	rows, err := m.DB.QueryContext(ctx, itemQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item OrderItem
		var productName string // Temporary variable to hold the joined name

		err := rows.Scan(
			&item.ID,
			&item.VariantID,
			&item.Quantity,
			&item.PriceAtReserve,
			&productName,
		)
		if err != nil {
			return nil, err
		}
		// Note: You might want to add a 'ProductName' field to your OrderItem struct
		order.OrderItems = append(order.OrderItems, item)
	}

	return &order, nil
}

func (m OrderModel) UpdateStatus(ctx context.Context, id int64, newStatus string, locationID int64) error {
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var currentStatus string
	err = tx.QueryRowContext(ctx, "SELECT status FROM orders WHERE order_id = $1 FOR UPDATE", id).Scan(&currentStatus)
	if err != nil {
		return err
	}

	if newStatus == "Cancelled" && currentStatus != "Cancelled" {
		// --- STEP A: Read all items into memory first ---
		type item struct {
			vID int64
			qty int
		}
		var itemsToRelease []item

		rows, err := tx.QueryContext(ctx, "SELECT variant_id, quantity FROM order_items WHERE order_id = $1", id)
		if err != nil {
			return err
		}

		for rows.Next() {
			var i item
			if err := rows.Scan(&i.vID, &i.qty); err != nil {
				rows.Close()
				return err
			}
			itemsToRelease = append(itemsToRelease, i)
		}
		rows.Close() // Close it explicitly here!

		// --- STEP B: Now perform the updates ---
		for _, i := range itemsToRelease {
			releaseQuery := `
                UPDATE inventory 
                SET stock_reserved = stock_reserved - $1 
                WHERE variant_id = $2 AND location_id = $3`

			_, err = tx.ExecContext(ctx, releaseQuery, i.qty, i.vID, locationID)
			if err != nil {
				return fmt.Errorf("failed to release stock for variant %d: %w", i.vID, err)
			}
		}
	}

	// 3. Update the order status
	_, err = tx.ExecContext(ctx, "UPDATE orders SET status = $1, updated_at = NOW() WHERE order_id = $2", newStatus, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}
