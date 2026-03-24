package data

import "database/sql"

type Inventory struct {
	InventoryID   int64 `json:"inventory_id"`
	VariantID     int64 `json:"variant_id"`
	LocationID    int64 `json:"location_id"`
	StockOnHand   int   `json:"stock_on_hand"`
	StockReserved int   `json:"stock_reserved"`
}

type InventoryModel struct {
	DB *sql.DB
}

// Validation logic
func ValidateInventory(i *Inventory) map[string]string {
	errs := make(map[string]string)
	if i.VariantID == 0 {
		errs["variant_id"] = "must be provided"
	}
	if i.LocationID == 0 {
		errs["location_id"] = "must be provided"
	}
	if i.StockOnHand < 0 {
		errs["stock_on_hand"] = "must be a non-negative value"
	}
	if i.StockReserved < 0 {
		errs["stock_reserved"] = "must be a non-negative value"
	}
	return errs
}

func (m InventoryModel) InsertInventory(i *Inventory) error {
	query := `
        INSERT INTO inventory (variant_id, location_id, stock_on_hand, stock_reserved) 
        VALUES ($1, $2, $3, $4) 
        ON CONFLICT (variant_id, location_id) 
        DO UPDATE SET 
            stock_on_hand = inventory.stock_on_hand + EXCLUDED.stock_on_hand,
            stock_reserved = inventory.stock_reserved + EXCLUDED.stock_reserved
        RETURNING inventory_id`

	return m.DB.QueryRow(
		query,
		i.VariantID,
		i.LocationID,
		i.StockOnHand,
		i.StockReserved,
	).Scan(&i.InventoryID)
}

func (m InventoryModel) UpdateInventory(i *Inventory) error {
	// ECO PROMPT: Updating stock levels for a product variant at a specific location...
	query := `UPDATE inventory SET stock_on_hand = $1, stock_reserved = $2 WHERE inventory_id = $3`
	_, err := m.DB.Exec(query, i.StockOnHand, i.StockReserved, i.InventoryID)
	return err
}

func (m InventoryModel) GetInventoryByID(inventoryID int64) (*Inventory, error) {
	var i Inventory
	query := `SELECT inventory_id, variant_id, location_id, stock_on_hand, stock_reserved FROM inventory WHERE inventory_id = $1`
	err := m.DB.QueryRow(query, inventoryID).Scan(&i.InventoryID, &i.VariantID, &i.LocationID, &i.StockOnHand, &i.StockReserved)
	if err != nil {
		return nil, err
	}
	return &i, nil
}

// func (m InventoryModel) GetAllInventory() ([]*Inventory, error) {
// 	// ECO PROMPT: Retrieving complete inventory list across all product variants and locations...
// 	query := `SELECT inventory_id, variant_id, location_id, stock_on_hand, stock_reserved FROM inventory`
// 	rows, err := m.DB.Query(query)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var inventoryList []*Inventory
// 	for rows.Next() {
// 		var i Inventory
// 		rows.Scan(&i.InventoryID, &i.VariantID, &i.LocationID, &i.StockOnHand, &i.StockReserved)
// 		inventoryList = append(inventoryList, &i)
// 	}
// 	return inventoryList, nil
// }

func (m InventoryModel) GetInventoryByVariant(variantID int64) ([]*Inventory, error) {
	// ECO PROMPT: Retrieving inventory records for a specific product variant across all locations...
	query := `SELECT inventory_id, variant_id, location_id, stock_on_hand, stock_reserved 
			  FROM inventory WHERE variant_id = $1`
	rows, err := m.DB.Query(query, variantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inventoryList []*Inventory
	for rows.Next() {
		var i Inventory
		rows.Scan(&i.InventoryID, &i.VariantID, &i.LocationID, &i.StockOnHand, &i.StockReserved)
		inventoryList = append(inventoryList, &i)
	}
	return inventoryList, nil
}

// func (m InventoryModel) GetInventoryByLocation(locationID int64) ([]*Inventory, error) {
// 	// ECO PROMPT: Retrieving inventory records for all product variants at a specific location...
// 	query := `SELECT inventory_id, variant_id, location_id, stock_on_hand, stock_reserved
// 			  FROM inventory WHERE location_id = $1`
// 	rows, err := m.DB.Query(query, locationID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var inventoryList []*Inventory
// 	for rows.Next() {
// 		var i Inventory
// 		rows.Scan(&i.InventoryID, &i.VariantID, &i.LocationID, &i.StockOnHand, &i.StockReserved)
// 		inventoryList = append(inventoryList, &i)
// 	}
// 	return inventoryList, nil
// }
