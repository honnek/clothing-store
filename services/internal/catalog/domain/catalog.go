package domain

import "errors"

// ErrProductNotFound возвращается, когда товар не найден или скрыт (is_deleted).
var ErrProductNotFound = errors.New("product not found")

// Product — товар в том виде, в котором его отдаёт каталог наружу.
type Product struct {
	ID          int32  `json:"id"`
	UUID        string `json:"uuid"`
	Title       string `json:"title"`
	Price       string `json:"price"`
	Quality     int32  `json:"quality"`
	Description string `json:"description,omitempty"`
	Slug        string `json:"slug,omitempty"`
	CategoryID  *int32 `json:"category_id,omitempty"`
	IsPublished bool   `json:"is_published"`
}

type Category struct {
	ID    int32  `json:"id"`
	Title string `json:"title"`
	Slug  string `json:"slug"`
}

// ProductFilter — необязательные фильтры списка. nil-поле означает «не фильтровать».
type ProductFilter struct {
	CategoryID *int32
	Published  *bool
	Search     *string
}

// Page — окно пагинации.
type Page struct {
	Limit  int32
	Offset int32
}

// ProductList — страница товаров вместе с полным количеством под текущий фильтр.
type ProductList struct {
	Items []Product `json:"items"`
	Total int64     `json:"total"`
}
