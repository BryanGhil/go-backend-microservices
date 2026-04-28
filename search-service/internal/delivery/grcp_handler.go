package delivery

import (
	"context"
	"ecommerce/pb"
	"ecommerce/search-service/internal/domain"
)

type SearchGrpcHandler struct {
	pb.UnimplementedSearchServiceServer
	repo domain.SearchRepository
}

func NewSearchGrpcHandler(repo domain.SearchRepository) *SearchGrpcHandler {
	return &SearchGrpcHandler{repo: repo}
}

func (h *SearchGrpcHandler) SearchProducts(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	limit := req.GetLimit()
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	offset := req.GetOffset()
	if offset < 0 {
		offset = 0
	}

	results, totalCount, err := h.repo.Search(ctx, req.GetQuery(), limit, offset)
	if err != nil {
		return nil, err
	}

	var pbProducts []*pb.Product
	for _, p := range results {
		pbProducts = append(pbProducts, &pb.Product{
			Id:          p.ID,
			SellerId:    p.SellerID,
			Name:        p.Name,
			Description: p.Description,
			Category:    p.Category,
			Price:       p.Price, // Float64 now matches the proto double
			ImageUrl:    p.ImageURL,
			IsActive:    p.IsActive,
		})
	}

	// Assuming you added total_count to the SearchResponse in product.proto
	return &pb.SearchResponse{
		Products:   pbProducts,
		TotalCount: totalCount,
	}, nil
}
