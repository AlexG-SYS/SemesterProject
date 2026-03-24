Multi-Location Inventory & Sales API
A production-ready RESTful API built in Go for managing retail operations in Belize. This system handles complex order processing, atomic inventory reservations, and features a robust middleware stack.

🚀 Key Features
Atomic Inventory Reservations: Uses PostgreSQL transactions (BEGIN/COMMIT) and row-level locking (FOR UPDATE) to prevent overselling.

Location-Aware Stock: Manage unique inventory counts for different warehouses (e.g., Belmopan, Belize City).

Financial Integrity: "Price-at-Reserve" logic captures the selling price and cost at the moment of order creation.

Resilient Middleware: Includes Rate Limiting (429), Panic Recovery, CORS handling, and Gzip compression.

🛠️ Tech Stack
Language: Go 1.22+

Database: PostgreSQL 16+

Router: Standard Library http.ServeMux

Logging: Structured JSON logging with slog

🏃 Getting Started
1. Database Setup
Ensure PostgreSQL is running, then create the database and run migrations:

Bash
make db/create
make db/migrations/up
2. Run the Application
Bash
make run/api
The server will start at http://localhost:4000.

🧪 Demonstration Guide
Use these commands to verify the core requirements of the project:

1. Rate Limiting (HTTP 429)
Simulate a burst of traffic to trigger the security middleware:

Bash
# Hit the healthcheck 10 times rapidly
for i in {1..10}; do curl -i http://localhost:4000/v1/healthcheck; done
2. CORS Preflight
Verify that the API allows cross-origin requests from front-end applications:

Bash
curl -i -X OPTIONS http://localhost:4000/v1/orders \
-H "Access-Control-Request-Method: POST" \
-H "Origin: http://example.com"
3. Response Compression (Gzip)
Check that the API compresses large JSON payloads to save bandwidth:

Bash
curl -i -H "Accept-Encoding: gzip" http://localhost:4000/v1/orders/1
4. Metrics & Observability
View internal server statistics and memory usage:

Bash
curl -i http://localhost:4000/debug/vars
📂 Project Structure
cmd/api/: Application entry point and dependency injection.
internal/data/: Database models and transaction logic.
internal/handlers/: HTTP request handlers and JSON parsing.
internal/middleware/: Security, logging, and performance layers.
migrations/: SQL files defining the database schema.

📝 Design Decisions
Pessimistic Locking: We chose FOR UPDATE on inventory rows to ensure that stock levels remain accurate even under high concurrent load.

Structured Logging: We implemented JSON logging to enable easier debugging and performance monitoring via latency tracking.

Graceful Cancellations: When an order is cancelled, a dedicated database transaction releases the stock_reserved back into the available pool.
