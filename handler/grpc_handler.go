package handler

import (
	"context"
	"database/sql"
	pb "phase3-api-architecture/pb/proto/inventory"
	"phase3-api-architecture/repository"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcInventoryHandler struct {
	pb.UnimplementedInventoryServiceServer

	Repo *repository.ProductRepository
}

func (h *GrpcInventoryHandler) GetStock(ctx context.Context, req *pb.GetStockRequest) (*pb.GetStockResponse, error) {
	product, err := h.Repo.GetByID(ctx, int(req.Id))

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "produk tidak ditemukan")
		}
		return nil, status.Error(codes.Internal, "error database")
	}

	return &pb.GetStockResponse{
		Id:    int32(product.ID),
		Name:  product.Name,
		Stock: int32(product.Stock),
	}, nil
}
