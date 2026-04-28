package repository

import (
	"bytes"
	"context"
	"ecommerce/search-service/internal/domain"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type esRepo struct {
	client *elasticsearch.Client
}

func NewElasticsearchRepo(client *elasticsearch.Client) domain.SearchRepository {
	return &esRepo{client: client}
}

func (r *esRepo) IndexProduct(ctx context.Context, p *domain.Product) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      "products",
		DocumentID: strconv.FormatInt(p.ID, 10),
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing document: %s", res.String())
	}
	return nil
}

// ... (Keep IndexProduct as it is) ...

func (r *esRepo) Search(ctx context.Context, query string, limit, offset int32) ([]*domain.Product, int64, error) {
	var buf bytes.Buffer
	
	// 1. Build the query with Pagination ("from" and "size")
	q := map[string]interface{}{
		"from": offset,
		"size": limit,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"multi_match": map[string]interface{}{
							"query":  query,
							"fields": []string{"name^3", "category^2", "description"},
						},
					},
				},
				"filter": []interface{}{
					map[string]interface{}{
						"term": map[string]interface{}{
							"is_active": true,
						},
					},
				},
			},
		},
	}
	
	if err := json.NewEncoder(&buf).Encode(q); err != nil {
		return nil, 0, err
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("products"),
		r.client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, 0, fmt.Errorf("elasticsearch error: %s", res.Status())
	}

	var rMap map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&rMap); err != nil {
		return nil, 0, err
	}

	hitsData, ok := rMap["hits"].(map[string]interface{})
	if !ok {
		return nil, 0, nil
	}

	// 2. Extract Total Count for the frontend
	var totalCount int64
	if total, ok := hitsData["total"].(map[string]interface{}); ok {
		if val, ok := total["value"].(float64); ok {
			totalCount = int64(val)
		}
	}

	hitsList, ok := hitsData["hits"].([]interface{})
	if !ok {
		return nil, totalCount, nil
	}

	var products []*domain.Product
	for _, hit := range hitsList {
		source := hit.(map[string]interface{})["_source"]
		var p domain.Product
		b, _ := json.Marshal(source)
		json.Unmarshal(b, &p)
		products = append(products, &p)
	}
	
	// 3. Return the products AND the total count
	return products, totalCount, nil
}