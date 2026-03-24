include .envrc

# Default variables
DB_DSN ?= ${ECOMMERCE_DB_DSN}


## run: run the cmd/api application
.PHONY: run
run:
	@echo  'Running application…'
	@go run ./cmd/api \
		-port=${PORT} \
		-env=${ENV} \
		-db-dsn="${DB_DSN}" \
		-limiter-rps=${LIMITER_RPS} \
		-limiter-burst=${LIMITER_BURST} \
		-limiter-enabled=${LIMITER_ENABLED} \
		-cors-trusted-origins="${CORS_TRUSTED_ORIGINS}"


## db/psql: connect to the database using psql (terminal)
.PHONY: db/psql
db/psql:
	psql ${DB_DSN}

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up:
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${ECOMMERCE_DB_DSN} up

## db/migrations/down: apply all down database migrations
.PHONY: db/migrations/down
db/migrations/down:
	@echo 'Running down migrations...'
	migrate -path ./migrations -database ${ECOMMERCE_DB_DSN} down

## db/migrations/fix: force the database to a specific version (use v=NUMBER)
.PHONY: db/migrations/fix
db/migrations/fix:
	migrate -path ./migrations -database ${ECOMMERCE_DB_DSN} force ${v}

# ==================================================================================== #
# API REQUETS (FOR DEMO)
# ==================================================================================== #

## demo/setup: Populate categories and locations first
.PHONY: demo/setup
demo/setup:
	@echo "Creating Category..."
	curl -i -X POST http://localhost:4000/v1/categories -d '{"name": "Electronics", "description": "Gadgets and devices"}'
	@echo "\nCreating Location..."
	curl -i -X POST http://localhost:4000/v1/locations -d '{"name": "Belmopan Warehouse", "address": "123 Orchard Garden"}'

## demo/category/list: Get all categories
.PHONY: demo/category/list
demo/category/list:
	curl -i http://localhost:4000/v1/categories	

## demo/category/update id=$1: Update a category's name and description
.PHONY: demo/category/update
demo/category/update:
	curl -i -X PATCH http://localhost:4000/v1/categories/${id} \
	-d '{"name": "Updated Electronics", "description": "Updated Gadgets and devices"}'

## demo/location/update id=$1: Update a location's name and address
.PHONY: demo/location/update
demo/location/update:
	curl -i -X PATCH http://localhost:4000/v1/locations/${id} \
	-d '{"name": "Updated Belmopan Warehouse", "address": "456 New Orchard Garden"}'

## demo/location/list: Get all locations
.PHONY: demo/location/list
demo/location/list:
	curl -i http://localhost:4000/v1/locations

## demo/product/create: Add a new product (Assumes Category ID 1 exists)
.PHONY: demo/product/create
demo/product/create:
	curl -i -X POST http://localhost:4000/v1/products \
	-d '{"category_id": 1, "name": "iPhone 15", "description": "Latest Apple Smartphone", "is_gst_eligible": true}'

## demo/product/list: Get all products with filters, sorting, and pagination
.PHONY: demo/product/list
demo/product/list:
	curl -i "http://localhost:4000/v1/products?name=iphone&category_id=1&sort=-product_id&page=1&page_size=5"

## demo/product/get id=$1: Get a specific product by ID
.PHONY: demo/product/get
demo/product/get:
	curl -i http://localhost:4000/v1/products/${id}

## demo/product/update id=$1: Partial update of a product
.PHONY: demo/product/update
demo/product/update:
	curl -i -X PATCH http://localhost:4000/v1/products/${id} \
	-d '{"name": "iPhone 15 Pro Max", "is_gst_eligible": false}'

## demo/variant/create: Add specifics to Product ID 1
.PHONY: demo/variant/create
demo/variant/create:
	curl -i -X POST http://localhost:4000/v1/variants \
	-d '{"product_id": 1, "sku": "IPH15-BLU-130", "size_attr": "128GB", "color_attr": "Blue", "cost_price": 700.00, "selling_price": 999.00}'

## demo/variant/list id=$1: See all specifics for a product
.PHONY: demo/variant/list
demo/variant/list:
	curl -i http://localhost:4000/v1/products/${id}/variants

## demo/variant/update id=$1: Partial update of a variant
.PHONY: demo/variant/update
demo/variant/update:
	curl -i -X PATCH http://localhost:4000/v1/variants/${id} \
	-d '{"selling_price": 1099.00}'
	
## demo/inventory/create: Add initial stock for Variant ID 1 at Location ID 1
.PHONY: demo/inventory/create
demo/inventory/create:
	curl -i -X POST http://localhost:4000/v1/inventory \
	-d '{"variant_id": 4, "location_id": 1, "stock_on_hand": 100}'

## demo/inventory/list: Get all inventory records
.PHONY: demo/inventory/list
demo/inventory/list:
	curl -i http://localhost:4000/v1/variants/${id}/inventory

## demo/inventory/update: Add stock to a variant at a location
.PHONY: demo/inventory/update
demo/inventory/update:
	curl -i -X PATCH http://localhost:4000/v1/inventory/${id} \
	-d '{"stock_on_hand": 150, "stock_reserved": 25}'

## demo/profile/create: Create a new user account and profile record
.PHONY: demo/profile/create
demo/profile/create:
	curl -i -X POST http://localhost:4000/v1/profiles \
	-d '{"email": "alexxxx@example.com", "password": "pa55word", "full_name": "Alex G", "phone": "555-0199", "address": "123 Go Lane", "district": "Cayo", "town_village": "San Ignacio"}'

## demo/profile/get: Fetch a combined user and profile object by ID
.PHONY: demo/profile/get
demo/profile/get:
	curl -i -X GET http://localhost:4000/v1/profiles/1

## demo/profile/update: Partially update profile details (PATCH)
.PHONY: demo/profile/update
demo/profile/update:
	curl -i -X PATCH http://localhost:4000/v1/profiles/${id} \
	-d '{"phone": "999-0000", "district": "Cayo", "town_village": "San Ignacio"}'


# demo/shipping/create: Add a new shipping provider (e.g., BPMS or DHL)
.PHONY: demo/shipping/create
demo/shipping/create:
	@echo "Creating shipping method..."
	curl -i -X POST http://localhost:4000/v1/shipping \
	-d '{"provider_name": "BPMS", "service_type": "Next Day", "base_rate": 15.00, "contact_phone": "223-4567"}'

## demo/shipping/get: Fetch details of shipping method 1
.PHONY: demo/shipping/get
demo/shipping/get:
	@echo "Fetching shipping method ID 1..."
	curl -i -X GET http://localhost:4000/v1/shipping/${id} 

## demo/shipping/update: Partially update a shipping provider (PATCH)
.PHONY: demo/shipping/update
demo/shipping/update:
	@echo "Updating base rate for shipping method 1..."
	curl -i -X PATCH http://localhost:4000/v1/shipping/${id} \
	-d '{"base_rate": 12.50, "contact_phone": "223-9999"}'

## demo/order/create: Place a real order (fetches price & reserves stock)
.PHONY: demo/order/create
demo/order/create:
	curl -i -X POST http://localhost:4000/v1/orders \
	-d '{"customer_id": 2, "location_id": 1, "shipping_method_id": 1, "items": [{"variant_id": 4, "quantity": 1}]}'


## demo/limiter: Test the rate limiter by hammering the healthcheck endpoint
.PHONY: demo/limiter
demo/limiter:
	@echo "Sending 10 rapid requests to test rate limiting..."
	seq 10 | xargs -I % -P 10 curl -i http://localhost:4000/v1/metrics

## demo/order/get: Fetch details for order 1
.PHONY: demo/order/get
demo/order/get:
	@echo "Retrieving Order..."
	curl -i -X GET http://localhost:4000/v1/orders/${id}

## demo/order/cancel: Cancel order 1 and release inventory holds
.PHONY: demo/order/cancel
demo/order/cancel:
	@echo "Cancelling Order..."
	curl -i -X PATCH http://localhost:4000/v1/orders/${id} \
	-d '{"status": "Cancelled", "location_id": 1}'

## demo/order/pay: Mark order 1 as Paid
.PHONY: demo/order/pay
demo/order/pay:
	curl -i -X PATCH http://localhost:4000/v1/orders/${id} \
	-d '{"status": "Paid", "location_id": 1}'

## demo/cors: Test if CORS headers are being returned
.PHONY: demo/cors
demo/cors:
	curl -i -X OPTIONS http://localhost:4000/v1/healthcheck \
	-H "Access-Control-Request-Method: PATCH" \
	-H "Origin: http://localhost:3000"

## demo/gzip: Test if the response is compressed
.PHONY: demo/gzip
demo/gzip:
	@echo "Requesting with gzip..."
	curl -i -H "Accept-Encoding: gzip" http://localhost:4000/v1/products | grep "Content-Encoding"
	@echo "\nComparing sizes (Raw vs Gzip):"
	@curl -s -o /dev/null -w "Raw Size: %{size_download} bytes\n" http://localhost:4000/v1/products
	@curl -s -o /dev/null -w "Gzip Size: %{size_download} bytes\n" -H "Accept-Encoding: gzip" http://localhost:4000/v1/products

## demo/metrics: Generate traffic and then view the metrics report
.PHONY: demo/metrics
demo/metrics:
	@echo "Generating traffic..."
	@curl -s http://localhost:4000/v1/healthcheck > /dev/null
	@curl -s http://localhost:4000/v1/products > /dev/null
	@curl -s http://localhost:4000/v1/invalid-route > /dev/null
	@echo "\n--- API METRICS REPORT ---"
	curl -s http://localhost:4000/v1/metrics | jq .