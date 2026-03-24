package routes

import (
	"log/slog"
	"net/http"

	"database/sql"

	"github.com/AlexG-SYS/semesterproject/internal/data"
	"github.com/AlexG-SYS/semesterproject/internal/handlers"
	"github.com/AlexG-SYS/semesterproject/internal/helpers"
	"github.com/AlexG-SYS/semesterproject/internal/middleware"
)

func SetupRoutes(db *sql.DB, logger *slog.Logger, rps float64, burst int, enabled bool, origins []string) http.Handler {
	app := &helpers.Application{Logger: logger}
	// 1. Initialize the models with the DB connection
	models := data.NewModels(db)

	h := &handlers.Handler{
		App:    app,
		Models: models,
	}

	mw := middleware.Middleware{
		App:            app,
		LimiterRPS:     rps,
		LimiterBurst:   burst,
		LimiterEnabled: enabled,
		TrustedOrigins: origins,
	}

	mux := http.NewServeMux()

	// Global / Home
	mux.HandleFunc("GET /", h.HomeHandler)
	mux.HandleFunc("POST /login", h.LoginHandler)
	mux.HandleFunc("GET /v1/metrics", h.MetricsHandler)

	// --- CATEGORIES ---
	mux.HandleFunc("POST /v1/categories", h.CreateCategoryHandler)
	mux.HandleFunc("GET /v1/categories", h.ListCategoriesHandler)
	mux.HandleFunc("PATCH /v1/categories/{id}", h.UpdateCategoryHandler)

	// --- LOCATIONS ---
	mux.HandleFunc("POST /v1/locations", h.CreateLocationHandler)
	mux.HandleFunc("GET /v1/locations", h.ListLocationsHandler)
	mux.HandleFunc("PATCH /v1/locations/{id}", h.UpdateLocationHandler)

	// --- PRODUCTS ---
	mux.HandleFunc("POST /v1/products", h.CreateProductHandler)
	mux.HandleFunc("GET /v1/products", h.ListProductsHandler)
	mux.HandleFunc("GET /v1/products/{id}", h.GetProductHandler)
	mux.HandleFunc("PATCH /v1/products/{id}", h.UpdateProductHandler) // Changed to PATCH for partial updates

	// --- VARIANTS (Product Specifics) ---
	mux.HandleFunc("POST /v1/variants", h.CreateVariantHandler)
	mux.HandleFunc("GET /v1/products/{id}/variants", h.ListVariantsHandler)
	mux.HandleFunc("PATCH /v1/variants/{id}", h.UpdateVariantHandler) // Changed to PATCH for partial updates

	// --- INVENTORY ---
	mux.HandleFunc("POST /v1/inventory", h.CreateInventoryHandler)
	mux.HandleFunc("GET /v1/variants/{id}/inventory", h.GetInventoryHandler)
	mux.HandleFunc("PATCH /v1/inventory/{id}", h.UpdateInventoryHandler)

	// --- PROFILES ---
	mux.HandleFunc("POST /v1/profiles", h.CreateProfileHandler)
	mux.HandleFunc("GET /v1/profiles/{id}", h.GetProfileHandler)
	mux.HandleFunc("PATCH /v1/profiles/{id}", h.UpdateProfileHandler)

	// --- SHIPPING ---
	mux.HandleFunc("POST /v1/shipping", h.CreateShippingHandler)
	mux.HandleFunc("GET /v1/shipping/{id}", h.GetShippingHandler)
	mux.HandleFunc("PATCH /v1/shipping/{id}", h.UpdateShippingHandler)

	// --- ORDERS ---
	mux.HandleFunc("POST /v1/orders", h.CreateOrderHandler)
	mux.HandleFunc("GET /v1/orders/{id}", h.GetOrderHandler)
	mux.HandleFunc("PATCH /v1/orders/{id}", h.UpdateOrderHandler)

	// Middleware Chain (Executed bottom to top)
	return mw.Metrics(
		mw.RateLimit( // 1. First, check if they are allowed in
			mw.Logger( // 2. Then, log the request details
				mw.Compress( // 3. Then, prepare to compress the response
					mw.EnableCORS(mux), // 4. Finally, handle CORS and Routing
				),
			),
		),
	)
}
