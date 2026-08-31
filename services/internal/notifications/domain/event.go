// Package domain описывает событие заказа так, как его читает воркер уведомлений.
// Структура своя, а не общая с order-service: потребитель берёт из payload только
// то, что ему нужно для письма, и переживает появление новых полей у продюсера.
package domain

import "time"

type OrderItem struct {
	Title     string `json:"title"`
	UnitPrice string `json:"unit_price"`
	Quantity  int32  `json:"quantity"`
	LineTotal string `json:"line_total"`
}

type OrderCreated struct {
	OrderID   int32       `json:"order_id"`
	OwnerID   *int32      `json:"owner_id"`
	Total     string      `json:"total"`
	Items     []OrderItem `json:"items"`
	CreatedAt time.Time   `json:"created_at"`
}
