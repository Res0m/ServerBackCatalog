package graph

import (
	"backend-catalog/graph/model"
	"context"
	"fmt"
)

// Resolver - главный резолвер
type Resolver struct{}

// QueryResolver - определение резолвера для запроса

// Query - метод для получения Query
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

// Метод для получения списка товаров
func (r *queryResolver) Products(ctx context.Context) ([]*model.Product, error) {
	products := []*model.Product{
		{
			ID:          "1",
			Name:        "Product 1",
			Price:       10.99,
			Description: "Description of Product 1",
		},
		{
			ID:          "2",
			Name:        "Product 2",
			Price:       20.99,
			Description: "Description of Product 2",
		},
	}
	if len(products) == 0 {
		return nil, fmt.Errorf("no products found")
	}
	return products, nil
}
