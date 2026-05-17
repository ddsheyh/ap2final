package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"payment-service/internal/domain"
	"payment-service/internal/usecase"
	pb "payment-service/pkg/pb"
)

type PaymentHandler struct {
	pb.UnimplementedPaymentServiceServer
	uc *usecase.PaymentUsecase
}

func NewPaymentHandler(uc *usecase.PaymentUsecase) *PaymentHandler {
	return &PaymentHandler{uc: uc}
}

func toProtoPayment(p *domain.Payment) *pb.Payment {
	return &pb.Payment{
		Id:            p.ID,
		OrderId:       p.OrderID,
		UserId:        p.UserID,
		Amount:        p.Amount,
		Currency:      p.Currency,
		Status:        p.Status,
		PaymentMethod: p.PaymentMethod,
		CreatedAt:     timestamppb.New(p.CreatedAt),
		UpdatedAt:     timestamppb.New(p.UpdatedAt),
	}
}

func toProtoTransaction(t *domain.Transaction) *pb.Transaction {
	return &pb.Transaction{
		Id:          t.ID,
		PaymentId:   t.PaymentID,
		Type:        t.Type,
		Amount:      t.Amount,
		Status:      t.Status,
		Description: t.Description,
		CreatedAt:   timestamppb.New(t.CreatedAt),
	}
}

func (h *PaymentHandler) CreatePayment(ctx context.Context, req *pb.CreatePaymentRequest) (*pb.CreatePaymentResponse, error) {
	if req.OrderId == 0 || req.UserId == 0 || req.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "order_id, user_id, and amount are required")
	}

	payment, err := h.uc.CreatePayment(ctx, req.OrderId, req.UserId, req.Amount, req.Currency, req.PaymentMethod)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create payment failed: %v", err)
	}

	return &pb.CreatePaymentResponse{Payment: toProtoPayment(payment)}, nil
}

func (h *PaymentHandler) GetPayment(ctx context.Context, req *pb.GetPaymentRequest) (*pb.GetPaymentResponse, error) {
	payment, err := h.uc.GetPayment(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "payment not found: %v", err)
	}
	return &pb.GetPaymentResponse{Payment: toProtoPayment(payment)}, nil
}

func (h *PaymentHandler) ListPayments(ctx context.Context, req *pb.ListPaymentsRequest) (*pb.ListPaymentsResponse, error) {
	payments, total, err := h.uc.ListPayments(ctx, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list payments failed: %v", err)
	}

	pbPayments := make([]*pb.Payment, 0, len(payments))
	for _, p := range payments {
		pbPayments = append(pbPayments, toProtoPayment(p))
	}
	return &pb.ListPaymentsResponse{Payments: pbPayments, Total: int32(total)}, nil
}

func (h *PaymentHandler) GetPaymentsByOrder(ctx context.Context, req *pb.GetPaymentsByOrderRequest) (*pb.GetPaymentsByOrderResponse, error) {
	payments, err := h.uc.GetPaymentsByOrder(ctx, req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get payments by order failed: %v", err)
	}

	pbPayments := make([]*pb.Payment, 0, len(payments))
	for _, p := range payments {
		pbPayments = append(pbPayments, toProtoPayment(p))
	}
	return &pb.GetPaymentsByOrderResponse{Payments: pbPayments}, nil
}

func (h *PaymentHandler) UpdatePaymentStatus(ctx context.Context, req *pb.UpdatePaymentStatusRequest) (*pb.UpdatePaymentStatusResponse, error) {
	payment, err := h.uc.UpdatePaymentStatus(ctx, req.PaymentId, req.Status)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update status failed: %v", err)
	}
	return &pb.UpdatePaymentStatusResponse{Payment: toProtoPayment(payment)}, nil
}

func (h *PaymentHandler) CancelPayment(ctx context.Context, req *pb.CancelPaymentRequest) (*pb.CancelPaymentResponse, error) {
	payment, err := h.uc.CancelPayment(ctx, req.PaymentId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cancel payment failed: %v", err)
	}
	return &pb.CancelPaymentResponse{Payment: toProtoPayment(payment)}, nil
}

func (h *PaymentHandler) RefundPayment(ctx context.Context, req *pb.RefundPaymentRequest) (*pb.RefundPaymentResponse, error) {
	payment, tx, err := h.uc.RefundPayment(ctx, req.PaymentId, req.Amount)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "refund failed: %v", err)
	}
	return &pb.RefundPaymentResponse{
		Payment:           toProtoPayment(payment),
		RefundTransaction: toProtoTransaction(tx),
	}, nil
}

func (h *PaymentHandler) GetPaymentsByUser(ctx context.Context, req *pb.GetPaymentsByUserRequest) (*pb.GetPaymentsByUserResponse, error) {
	payments, total, err := h.uc.GetPaymentsByUser(ctx, req.UserId, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get payments by user failed: %v", err)
	}

	pbPayments := make([]*pb.Payment, 0, len(payments))
	for _, p := range payments {
		pbPayments = append(pbPayments, toProtoPayment(p))
	}
	return &pb.GetPaymentsByUserResponse{Payments: pbPayments, Total: int32(total)}, nil
}

func (h *PaymentHandler) ListTransactions(ctx context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
	txs, total, err := h.uc.ListTransactions(ctx, req.PaymentId, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list transactions failed: %v", err)
	}

	pbTxs := make([]*pb.Transaction, 0, len(txs))
	for _, t := range txs {
		pbTxs = append(pbTxs, toProtoTransaction(t))
	}
	return &pb.ListTransactionsResponse{Transactions: pbTxs, Total: int32(total)}, nil
}

func (h *PaymentHandler) GetTransaction(ctx context.Context, req *pb.GetTransactionRequest) (*pb.GetTransactionResponse, error) {
	tx, err := h.uc.GetTransaction(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "transaction not found: %v", err)
	}
	return &pb.GetTransactionResponse{Transaction: toProtoTransaction(tx)}, nil
}

func (h *PaymentHandler) GetPaymentStats(ctx context.Context, req *pb.GetPaymentStatsRequest) (*pb.GetPaymentStatsResponse, error) {
	stats, err := h.uc.GetPaymentStats(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get stats failed: %v", err)
	}
	return &pb.GetPaymentStatsResponse{
		TotalPayments:      stats.TotalPayments,
		TotalAmount:        stats.TotalAmount,
		SuccessfulPayments: stats.SuccessfulPayments,
		FailedPayments:     stats.FailedPayments,
		RefundedPayments:   stats.RefundedPayments,
	}, nil
}

func (h *PaymentHandler) RetryPayment(ctx context.Context, req *pb.RetryPaymentRequest) (*pb.RetryPaymentResponse, error) {
	payment, err := h.uc.RetryPayment(ctx, req.PaymentId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "retry payment failed: %v", err)
	}
	return &pb.RetryPaymentResponse{Payment: toProtoPayment(payment)}, nil
}
