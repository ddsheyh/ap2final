package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"order-service/internal/domain"
	"order-service/internal/usecase"
	pb "order-service/pkg/pb"
)

type OrderHandler struct {
	pb.UnimplementedOrderServiceServer
	uc *usecase.OrderUsecase
}

func NewOrderHandler(uc *usecase.OrderUsecase) *OrderHandler {
	return &OrderHandler{uc: uc}
}

func toProtoOrder(o *domain.Order) *pb.Order {
	po := &pb.Order{
		Id:         o.ID,
		UserId:     o.UserID,
		Status:     o.Status,
		TotalPrice: o.TotalPrice,
		CreatedAt:  timestamppb.New(o.CreatedAt),
		UpdatedAt:  timestamppb.New(o.UpdatedAt),
	}
	for _, item := range o.Items {
		po.Items = append(po.Items, toProtoItem(&item))
	}
	return po
}

func toProtoItem(i *domain.OrderItem) *pb.OrderItem {
	return &pb.OrderItem{
		Id:          i.ID,
		OrderId:     i.OrderID,
		ProductName: i.ProductName,
		Quantity:    i.Quantity,
		Price:       i.Price,
		CreatedAt:   timestamppb.New(i.CreatedAt),
	}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	if req.UserId == 0 || len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and items are required")
	}

	var items []domain.OrderItem
	for _, i := range req.Items {
		items = append(items, domain.OrderItem{
			ProductName: i.ProductName,
			Quantity:    i.Quantity,
			Price:       i.Price,
		})
	}

	order, err := h.uc.CreateOrder(ctx, req.UserId, items)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create order failed: %v", err)
	}

	return &pb.CreateOrderResponse{Order: toProtoOrder(order)}, nil
}

func (h *OrderHandler) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	order, err := h.uc.GetOrder(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "order not found: %v", err)
	}
	return &pb.GetOrderResponse{Order: toProtoOrder(order)}, nil
}

func (h *OrderHandler) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	orders, total, err := h.uc.ListOrders(ctx, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list orders failed: %v", err)
	}

	pbOrders := make([]*pb.Order, 0, len(orders))
	for _, o := range orders {
		pbOrders = append(pbOrders, toProtoOrder(o))
	}

	return &pb.ListOrdersResponse{Orders: pbOrders, Total: int32(total)}, nil
}

func (h *OrderHandler) UpdateOrder(ctx context.Context, req *pb.UpdateOrderRequest) (*pb.UpdateOrderResponse, error) {
	order, err := h.uc.UpdateOrder(ctx, req.Id, req.Status)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update order failed: %v", err)
	}
	return &pb.UpdateOrderResponse{Order: toProtoOrder(order)}, nil
}

func (h *OrderHandler) DeleteOrder(ctx context.Context, req *pb.DeleteOrderRequest) (*pb.DeleteOrderResponse, error) {
	if err := h.uc.DeleteOrder(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "delete order failed: %v", err)
	}
	return &pb.DeleteOrderResponse{Success: true}, nil
}

func (h *OrderHandler) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.UpdateOrderStatusResponse, error) {
	order, err := h.uc.UpdateOrderStatus(ctx, req.OrderId, req.Status)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update status failed: %v", err)
	}
	return &pb.UpdateOrderStatusResponse{Order: toProtoOrder(order)}, nil
}

func (h *OrderHandler) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	order, err := h.uc.CancelOrder(ctx, req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cancel order failed: %v", err)
	}
	return &pb.CancelOrderResponse{Order: toProtoOrder(order)}, nil
}

func (h *OrderHandler) GetOrdersByUser(ctx context.Context, req *pb.GetOrdersByUserRequest) (*pb.GetOrdersByUserResponse, error) {
	orders, total, err := h.uc.GetOrdersByUser(ctx, req.UserId, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get user orders failed: %v", err)
	}

	pbOrders := make([]*pb.Order, 0, len(orders))
	for _, o := range orders {
		pbOrders = append(pbOrders, toProtoOrder(o))
	}

	return &pb.GetOrdersByUserResponse{Orders: pbOrders, Total: int32(total)}, nil
}

func (h *OrderHandler) AddOrderItem(ctx context.Context, req *pb.AddOrderItemRequest) (*pb.AddOrderItemResponse, error) {
	item, err := h.uc.AddOrderItem(ctx, req.OrderId, req.ProductName, req.Quantity, req.Price)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "add item failed: %v", err)
	}
	return &pb.AddOrderItemResponse{Item: toProtoItem(item)}, nil
}

func (h *OrderHandler) RemoveOrderItem(ctx context.Context, req *pb.RemoveOrderItemRequest) (*pb.RemoveOrderItemResponse, error) {
	if err := h.uc.RemoveOrderItem(ctx, req.ItemId); err != nil {
		return nil, status.Errorf(codes.Internal, "remove item failed: %v", err)
	}
	return &pb.RemoveOrderItemResponse{Success: true}, nil
}

func (h *OrderHandler) GetOrderItems(ctx context.Context, req *pb.GetOrderItemsRequest) (*pb.GetOrderItemsResponse, error) {
	items, err := h.uc.GetOrderItems(ctx, req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get items failed: %v", err)
	}

	pbItems := make([]*pb.OrderItem, 0, len(items))
	for _, i := range items {
		pbItems = append(pbItems, toProtoItem(&i))
	}

	return &pb.GetOrderItemsResponse{Items: pbItems}, nil
}

func (h *OrderHandler) GetOrderTotal(ctx context.Context, req *pb.GetOrderTotalRequest) (*pb.GetOrderTotalResponse, error) {
	total, count, err := h.uc.GetOrderTotal(ctx, req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get total failed: %v", err)
	}
	return &pb.GetOrderTotalResponse{Total: total, ItemCount: count}, nil
}
