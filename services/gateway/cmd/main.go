package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	orderpb "gateway/pkg/orderpb"
	paymentpb "gateway/pkg/paymentpb"
	userpb "gateway/pkg/userpb"
)

var (
	userClient    userpb.UserServiceClient
	orderClient   orderpb.OrderServiceClient
	paymentClient paymentpb.PaymentServiceClient
	jwtSecret     []byte

	reqTotal  atomic.Int64
	reqErrors atomic.Int64
)

// helpers
func jsonResp(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func jsonErr(w http.ResponseWriter, code int, msg string) {
	reqErrors.Add(1)
	jsonResp(w, code, map[string]string{"error": msg})
}

func cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(204)
			return
		}
		reqTotal.Add(1)
		next(w, r)
	}
}

func extractUserID(r *http.Request) (int64, error) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return 0, fmt.Errorf("no token")
	}
	token, err := jwt.Parse(strings.TrimPrefix(auth, "Bearer "), func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}
	claims := token.Claims.(jwt.MapClaims)
	uid, _ := claims["user_id"].(float64)
	return int64(uid), nil
}

// monitoring

func handleHealth(w http.ResponseWriter, r *http.Request) {
	jsonResp(w, 200, map[string]interface{}{
		"status":  "ok",
		"service": "api-gateway",
		"time":    time.Now().UTC(),
	})
}

// returns Prometheus text format metrics
func handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "# HELP gateway_requests_total Total HTTP requests handled by the gateway\n")
	fmt.Fprintf(w, "# TYPE gateway_requests_total counter\n")
	fmt.Fprintf(w, "gateway_requests_total %d\n", reqTotal.Load())
	fmt.Fprintf(w, "# HELP gateway_errors_total Total HTTP errors returned by the gateway\n")
	fmt.Fprintf(w, "# TYPE gateway_errors_total counter\n")
	fmt.Fprintf(w, "gateway_errors_total %d\n", reqErrors.Load())
	fmt.Fprintf(w, "# HELP gateway_up Gateway service status\n")
	fmt.Fprintf(w, "# TYPE gateway_up gauge\n")
	fmt.Fprintf(w, "gateway_up 1\n")
}

// user routes

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonErr(w, 405, "method not allowed")
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	resp, err := userClient.Register(r.Context(), &userpb.RegisterRequest{Email: req.Email, Password: req.Password, Name: req.Name})
	if err != nil {
		jsonErr(w, 400, err.Error())
		return
	}
	jsonResp(w, 201, resp)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonErr(w, 405, "method not allowed")
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	resp, err := userClient.Login(r.Context(), &userpb.LoginRequest{Email: req.Email, Password: req.Password})
	if err != nil {
		jsonErr(w, 401, err.Error())
		return
	}
	jsonResp(w, 200, resp)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonErr(w, 405, "method not allowed")
		return
	}
	auth := r.Header.Get("Authorization")
	token := strings.TrimPrefix(auth, "Bearer ")
	resp, err := userClient.Logout(r.Context(), &userpb.LogoutRequest{AccessToken: token})
	if err != nil {
		jsonErr(w, 400, err.Error())
		return
	}
	jsonResp(w, 200, resp)
}

func handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonErr(w, 405, "method not allowed")
		return
	}
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	resp, err := userClient.RefreshToken(r.Context(), &userpb.RefreshTokenRequest{RefreshToken: req.RefreshToken})
	if err != nil {
		jsonErr(w, 401, err.Error())
		return
	}
	jsonResp(w, 200, resp)
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		size, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
		if page < 1 {
			page = 1
		}
		if size < 1 {
			size = 20
		}
		resp, err := userClient.ListUsers(r.Context(), &userpb.ListUsersRequest{Page: int32(page), PageSize: int32(size)})
		if err != nil {
			jsonErr(w, 500, err.Error())
			return
		}
		jsonResp(w, 200, resp)
	default:
		jsonErr(w, 405, "method not allowed")
	}
}

func handleUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	switch r.Method {
	case http.MethodGet:
		resp, err := userClient.GetUser(r.Context(), &userpb.GetUserRequest{Id: id})
		if err != nil {
			jsonErr(w, 404, err.Error())
			return
		}
		jsonResp(w, 200, resp)
	case http.MethodPut:
		var req struct{ Name, Email string }
		json.NewDecoder(r.Body).Decode(&req)
		resp, err := userClient.UpdateUser(r.Context(), &userpb.UpdateUserRequest{Id: id, Name: req.Name, Email: req.Email})
		if err != nil {
			jsonErr(w, 400, err.Error())
			return
		}
		jsonResp(w, 200, resp)
	case http.MethodDelete:
		resp, err := userClient.DeleteUser(r.Context(), &userpb.DeleteUserRequest{Id: id})
		if err != nil {
			jsonErr(w, 400, err.Error())
			return
		}
		jsonResp(w, 200, resp)
	default:
		jsonErr(w, 405, "method not allowed")
	}
}

// order routes

func handleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req struct {
			UserID int64 `json:"user_id"`
			Items  []struct {
				ProductName string  `json:"product_name"`
				Quantity    int32   `json:"quantity"`
				Price       float64 `json:"price"`
			} `json:"items"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		var items []*orderpb.CreateOrderItemInput
		for _, i := range req.Items {
			items = append(items, &orderpb.CreateOrderItemInput{ProductName: i.ProductName, Quantity: i.Quantity, Price: i.Price})
		}
		resp, err := orderClient.CreateOrder(r.Context(), &orderpb.CreateOrderRequest{UserId: req.UserID, Items: items})
		if err != nil {
			jsonErr(w, 400, err.Error())
			return
		}
		jsonResp(w, 201, resp)
	case http.MethodGet:
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		size, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
		resp, err := orderClient.ListOrders(r.Context(), &orderpb.ListOrdersRequest{Page: int32(page), PageSize: int32(size)})
		if err != nil {
			jsonErr(w, 500, err.Error())
			return
		}
		jsonResp(w, 200, resp)
	default:
		jsonErr(w, 405, "method not allowed")
	}
}

func handleOrderByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	switch r.Method {
	case http.MethodGet:
		resp, err := orderClient.GetOrder(r.Context(), &orderpb.GetOrderRequest{Id: id})
		if err != nil {
			jsonErr(w, 404, err.Error())
			return
		}
		jsonResp(w, 200, resp)
	case http.MethodDelete:
		resp, err := orderClient.DeleteOrder(r.Context(), &orderpb.DeleteOrderRequest{Id: id})
		if err != nil {
			jsonErr(w, 400, err.Error())
			return
		}
		jsonResp(w, 200, resp)
	default:
		jsonErr(w, 405, "method not allowed")
	}
}

// payment routes

func handlePayments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req struct {
			OrderID  int64   `json:"order_id"`
			UserID   int64   `json:"user_id"`
			Amount   float64 `json:"amount"`
			Currency string  `json:"currency"`
			Method   string  `json:"payment_method"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		resp, err := paymentClient.CreatePayment(r.Context(), &paymentpb.CreatePaymentRequest{
			OrderId: req.OrderID, UserId: req.UserID, Amount: req.Amount, Currency: req.Currency, PaymentMethod: req.Method,
		})
		if err != nil {
			jsonErr(w, 400, err.Error())
			return
		}
		jsonResp(w, 201, resp)
	case http.MethodGet:
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		size, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
		resp, err := paymentClient.ListPayments(r.Context(), &paymentpb.ListPaymentsRequest{Page: int32(page), PageSize: int32(size)})
		if err != nil {
			jsonErr(w, 500, err.Error())
			return
		}
		jsonResp(w, 200, resp)
	default:
		jsonErr(w, 405, "method not allowed")
	}
}

func handlePaymentByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/payments/")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	resp, err := paymentClient.GetPayment(r.Context(), &paymentpb.GetPaymentRequest{Id: id})
	if err != nil {
		jsonErr(w, 404, err.Error())
		return
	}
	jsonResp(w, 200, resp)
}

// front

func handleFrontend(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprint(w, indexHTML)
}

// main

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting API Gateway...")

	jwtSecret = []byte(getEnv("JWT_SECRET", "goticket-secret-key-2026"))

	userConn := mustDial(getEnv("USER_SERVICE_ADDR", "localhost:50051"))
	defer userConn.Close()
	userClient = userpb.NewUserServiceClient(userConn)

	orderConn := mustDial(getEnv("ORDER_SERVICE_ADDR", "localhost:50052"))
	defer orderConn.Close()
	orderClient = orderpb.NewOrderServiceClient(orderConn)

	paymentConn := mustDial(getEnv("PAYMENT_SERVICE_ADDR", "localhost:50053"))
	defer paymentConn.Close()
	paymentClient = paymentpb.NewPaymentServiceClient(paymentConn)

	log.Println("Connected to all gRPC services")

	mux := http.NewServeMux()

	// Monitoring
	mux.HandleFunc("/health", cors(handleHealth))
	mux.HandleFunc("/metrics", handleMetrics)

	// Auth API
	mux.HandleFunc("/api/auth/register", cors(handleRegister))
	mux.HandleFunc("/api/auth/login", cors(handleLogin))
	mux.HandleFunc("/api/auth/logout", cors(handleLogout))
	mux.HandleFunc("/api/auth/refresh", cors(handleRefreshToken))

	// Users API
	mux.HandleFunc("/api/users", cors(handleUsers))
	mux.HandleFunc("/api/users/", cors(handleUserByID))

	// Orders API
	mux.HandleFunc("/api/orders", cors(handleOrders))
	mux.HandleFunc("/api/orders/", cors(handleOrderByID))

	// Payments API
	mux.HandleFunc("/api/payments", cors(handlePayments))
	mux.HandleFunc("/api/payments/", cors(handlePaymentByID))

	// Frontend SPA — serves the web UI
	mux.HandleFunc("/", handleFrontend)

	port := getEnv("HTTP_PORT", "8080")
	server := &http.Server{Addr: ":" + port, Handler: mux}

	go func() {
		log.Printf("API Gateway listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down gateway...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

func mustDial(addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Cannot connect to %s: %v", addr, err)
	}
	return conn
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
