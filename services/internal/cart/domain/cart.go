package domain

import "errors"

var (
	// ErrProductNotFound — товара нет в каталоге (например, при добавлении несуществующего).
	ErrProductNotFound = errors.New("product not found")
	// ErrInvalidQuantity — количество должно быть положительным.
	ErrInvalidQuantity = errors.New("quantity must be positive")
)

type Item struct {
	ProductUUID string `json:"product_uuid"`
	Title       string `json:"title"`
	UnitPrice   string `json:"unit_price"`
	Quantity    int32  `json:"quantity"`
	LineTotal   string `json:"line_total"`
}

type Cart struct {
	SessionID string `json:"session_id"`
	Items     []Item `json:"items"`
	Total     string `json:"total"`
}

// CatalogProduct — то, что корзине нужно знать о товаре; приходит из catalog-сервиса.
type CatalogProduct struct {
	UUID  string
	Title string
	Price string
}
