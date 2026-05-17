# GoTicket — Event Ticketing Platform

### 12 gRPC endpoints per microservice
The project has exactly 3 microservices, and 12 gRPC endpoints are implemented for each.
- **user-service (12)**: `Register`, `Login`, `Logout`, `RefreshToken`, `GetUser`, `GetUserByEmail`, `ListUsers`, `UpdateUser`, `DeleteUser`, `ChangePassword`, `BanUser`, `UnbanUser`.
- **order-service (12)**: `CreateOrder`, `GetOrder`, `ListOrders`, `UpdateOrder`, `DeleteOrder`, `UpdateOrderStatus`, `CancelOrder`, `GetOrdersByUser`, `AddOrderItem`, `RemoveOrderItem`, `GetOrderItems`, `GetOrderTotal`.
- **payment-service (12)**: `CreatePayment`, `GetPayment`, `ListPayments`, `GetPaymentsByOrder`, `UpdatePaymentStatus`, `CancelPayment`, `RefundPayment`, `GetPaymentsByUser`, `ListTransactions`, `GetTransaction`, `GetPaymentStats`, `RetryPayment`.

All of them are described in the corresponding `proto/*.proto` files and implemented in `handler/grpc/handler.go`. HTTP API Gateway (`services/gateway`) proxies HTTP/JSON requests to these gRPC methods.

### NATS Message Queue
NATS is used for asynchronous communication between services.
- **User Events**: `user-service` publishes `user.registered` and `user.deleted`. `order-service` listens to `user.deleted` and automatically cancels all orders for the deleted user.
- **Order Events**: `order-service` publishes `order.created`, `order.cancelled`, and `order.completed`. `payment-service` listens to `order.created` (for automatic payment creation) and `order.cancelled` (for automatic refunds).
- **Payment Events**: `payment-service` publishes `payment.completed`, `payment.failed`, and `payment.refunded`.

### Databases + caches + migrations + transactions
- **PostgreSQL**: A single database instance with separate tables for each service. The structure is set up via migration scripts (`migrations/000001_init.up.sql`). The `pgx/v5` library.
- **Transactions**: In `order-service`, adding an order and its order items (Order Items) occurs in a single atomic SQL transaction. In `payment-service`, updating the payment status and creating a record in the `transactions` table are also performed atomically via `tx.Begin() / tx.Commit()`.
- **Redis**: Used to cache user profiles (`user-service`), order lists (`order-service`), and payment statuses (`payment-service`). Redis also stores `refresh` tokens and a blacklist of tokens during logout.

### Email sending via SMTP
The standard `net/smtp` package is used.
- A welcome email (`SendWelcomeEmail`) is sent to `user-service` when a new user registers.
- Emails are sent to `payment-service` when a payment is successful, when a payment is failed, and when a refund is issued (`Refund`).

The sending logic is executed asynchronously (in a goroutine) to avoid blocking the main API response.

### Unit + integration tests
Each service has unit tests written for the usecase layer, covering validation, JWT generation, password hashing, and business model verification. Stubs for integration tests have also been prepared.

## Launch and Testing Guide

### How to Run Unit Tests
The project includes 11 short unit tests that don't require a database startup. They test isolated business logic (token generation, password hashing, model validity, and default pagination).

To run the tests, open a terminal in the project root and run the following commands:

**For user-service:**
```bash
cd services/user-service
go test -short -v ./internal/usecase/
```

**For order-service:**
```bash
cd services/order-service
go test -short -v ./internal/usecase/
```

**For payment-service:**
```bash
cd services/payment-service
go test -short -v ./internal/usecase/
```

*The `-short` flag tells Go to skip heavy integration tests that require a database.*

### How to run Integration tests
The *_usecase_test.go files contain integration tests (for example, TestOrderCRUD_Integration ). They require PostgreSQL and Redis to be running.
1. First, start the infrastructure (DB and Redis) in the background:
   ```bash
   docker compose up -d postgres redis
   ```
2. Then run the tests **without** the `-short` flag:
    ```bash
   cd services/order-service
   go test -v ./internal/usecase/ -run Integration
   ```

### Full project launch
```bash
docker compose up --build
```

**How to test via API Gateway (port 8080):**

1. **User registration (HTTP -> API Gateway -> gRPC user-service):**
```bash
curl -X POST http://localhost:8080/api/auth/register \
-H "Content-Type: application/json" \
-d '{"email":"student@university.edu","password":"exam_password","name":"Student"}'
```
*(Copy `user_id` and `access_token` from the response)*

2. **Creating an order (HTTP -> API Gateway -> gRPC order-service):**
```bash
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_ACCESS_TOKEN>" \
  -d '{
        "user_id": 1,
        "items": [
          {"product_name": "Exam Ticket", "quantity": 1, "price": 5000}
        ]
      }'
```

3. **Verifying user payments (HTTP -> API Gateway -> gRPC payment-service):**
```bash
curl -X GET "http://localhost:8080/api/payments?page=1&page_size=10" \
-H "Authorization: Bearer <YOUR_ACCESS_TOKEN>"
```

**Verifying Grafana:**
Open your browser to `http://localhost:3000`.
