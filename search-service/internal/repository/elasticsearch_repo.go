package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"ecommerce/search-service/internal/domain"
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
	if err != nil { return err }

	// Prepare the request to save the document in the "products" index
	req := esapi.IndexRequest{
		Index:      "products",
		DocumentID: strconv.FormatInt(p.ID, 10), // Use Postgres ID as Elastic ID
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, r.client)
	if err != nil { return err }
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing document: %s", res.String())
	}
	return nil
}

func (r *esRepo) Search(ctx context.Context, query string) ([]*domain.Product, error) {
	var buf bytes.Buffer
	q := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"Name": query,
			},
		},
	}
	if err := json.NewEncoder(&buf).Encode(q); err != nil { return nil, err }

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("products"),
		r.client.Search.WithBody(&buf),
	)
	if err != nil { return nil, err }
	defer res.Body.Close()

	// ---> FIX 1: CHECK FOR ELASTICSEARCH ERRORS <---
	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch error: %s", res.Status())
	}

	var rMap map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&rMap); err != nil { return nil, err }

	// ---> FIX 2: SAFELY READ HITS <---
	hitsData, ok := rMap["hits"].(map[string]interface{})
	if !ok {
		return nil, nil // No results
	}
	
	hitsList, ok := hitsData["hits"].([]interface{})
	if !ok {
		return nil, nil // No results
	}

	var products []*domain.Product
	for _, hit := range hitsList {
		source := hit.(map[string]interface{})["_source"]
		var p domain.Product
		b, _ := json.Marshal(source)
		json.Unmarshal(b, &p)
		products = append(products, &p)
	}
	return products, nil
}