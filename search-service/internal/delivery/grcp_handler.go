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
	results, err := h.repo.Search(ctx, req.GetQuery())
	if err != nil {
		return nil, err
	}

	var pbProducts []*pb.Product
	for _, p := range results {
		pbProducts = append(pbProducts, &pb.Product{
			Id: p.ID, 
			Name: p.Name, 
			Price: float64(p.Price),
		})
	}
	return &pb.SearchResponse{
		Products: pbProducts,
		}, nil
}