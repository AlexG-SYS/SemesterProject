package data

import "database/sql"

type Models struct {
	Products   ProductModel
	Categories CategoryModel
	Locations  LocationModel
	Inventory  InventoryModel
	Profile    ProfileModel
	Users      UserModel
	Shipping   ShippingModel
	Orders     OrderModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Products:   ProductModel{DB: db},
		Categories: CategoryModel{DB: db},
		Locations:  LocationModel{DB: db},
		Inventory:  InventoryModel{DB: db},
		Profile:    ProfileModel{DB: db},
		Users:      UserModel{DB: db},
		Shipping:   ShippingModel{DB: db},
		Orders:     OrderModel{DB: db},
	}
}
