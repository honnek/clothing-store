package domain

import (
	"errors"
	"fmt"
	"time"
)

var (
	// ErrOrderNotFound — заказа с таким id нет либо он помечен удалённым.
	ErrOrderNotFound = errors.New("order not found")
	// ErrProductNotFound — в корзине оказался товар, которого нет в каталоге.
	ErrProductNotFound = errors.New("product not found")
	// ErrEmptyCart — оформлять нечего.
	ErrEmptyCart = errors.New("cart is empty")
	// ErrIdempotencyKeyRequired — checkout без ключа идемпотентности не принимаем.
	ErrIdempotencyKeyRequired = errors.New("idempotency key is required")
	// ErrInsufficientStock — остатка не хватает на запрошенное количество.
	ErrInsufficientStock = errors.New("insufficient stock")
	// ErrInvalidStatusTransition — переход запрещён статус-машиной.
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)

// Status повторяет целочисленные статусы PHP-монолита (App\Entity\StaticStorage\OrderStatus):
// колонка "order".status общая, значения расходиться не должны.
type Status int32

const (
	StatusCreated Status = iota
	StatusProcessed
	StatusComplected
	StatusDelivered
	StatusDenied
)

var statusNames = map[Status]string{
	StatusCreated:    "created",
	StatusProcessed:  "processed",
	StatusComplected: "complected",
	StatusDelivered:  "delivered",
	StatusDenied:     "denied",
}

// Отменить можно с любого нефинального шага, вперёд — только по одному.
var statusTransitions = map[Status][]Status{
	StatusCreated:    {StatusProcessed, StatusDenied},
	StatusProcessed:  {StatusComplected, StatusDenied},
	StatusComplected: {StatusDelivered, StatusDenied},
	StatusDelivered:  nil,
	StatusDenied:     nil,
}

func (s Status) String() string {
	if name, ok := statusNames[s]; ok {
		return name
	}
	return fmt.Sprintf("status(%d)", int32(s))
}

func (s Status) Valid() bool {
	_, ok := statusNames[s]
	return ok
}

// CanTransitionTo сообщает, разрешён ли переход статус-машиной. Переход в тот же
// статус считаем недопустимым — иначе повторный клик в админке маскирует ошибку.
func (s Status) CanTransitionTo(to Status) bool {
	for _, allowed := range statusTransitions[s] {
		if allowed == to {
			return true
		}
	}
	return false
}

// StockError говорит, какого именно товара не хватило: клиенту нужно показать позицию,
// а не голое «нет в наличии».
type StockError struct {
	ProductUUID string
	Requested   int32
	Available   int32
}

func (e *StockError) Error() string {
	return fmt.Sprintf("insufficient stock for %s: requested %d, available %d",
		e.ProductUUID, e.Requested, e.Available)
}

func (e *StockError) Unwrap() error { return ErrInsufficientStock }

// TransitionError несёт обе стороны запрещённого перехода.
type TransitionError struct {
	From Status
	To   Status
}

func (e *TransitionError) Error() string {
	return fmt.Sprintf("cannot move order from %s to %s", e.From, e.To)
}

func (e *TransitionError) Unwrap() error { return ErrInvalidStatusTransition }

// Item — строка заказа со снимком цены на момент оформления: каталог потом
// подорожает, а заказ должен остаться тем, что видел покупатель.
type Item struct {
	ProductID   int32  `json:"product_id"`
	ProductUUID string `json:"product_uuid"`
	Title       string `json:"title"`
	UnitPrice   string `json:"unit_price"`
	Quantity    int32  `json:"quantity"`
	LineTotal   string `json:"line_total"`
}

type Order struct {
	ID        int32     `json:"id"`
	OwnerID   *int32    `json:"owner_id,omitempty"`
	Status    Status    `json:"status"`
	Total     string    `json:"total"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Items     []Item    `json:"items,omitempty"`
}

// CheckoutRequest — вход оформления. IdempotencyKey задаёт клиент; повторный запрос
// с тем же ключом обязан вернуть тот же заказ, а не создать второй.
type CheckoutRequest struct {
	SessionID      string
	IdempotencyKey string
	OwnerID        *int32
}

// CheckoutLine — позиция, пришедшая из корзины. Цену корзина не диктует: её берут
// из товара под блокировкой уже внутри транзакции.
type CheckoutLine struct {
	ProductUUID string
	Quantity    int32
}

type OrderFilter struct {
	OwnerID *int32
	Status  *Status
}

type Page struct {
	Limit  int32
	Offset int32
}

// OrderList — страница заказов без позиций: список в админке рисует шапки,
// состав подтягивается уже по конкретному заказу.
type OrderList struct {
	Items []Order `json:"items"`
	Total int64   `json:"total"`
}

// OrderCreated — полезная нагрузка события order.created, которое checkout кладёт
// в outbox. Уходит в Kafka как есть, поэтому поля снимочные, без ссылок на каталог.
type OrderCreated struct {
	OrderID   int32     `json:"order_id"`
	SessionID string    `json:"session_id"`
	OwnerID   *int32    `json:"owner_id,omitempty"`
	Total     string    `json:"total"`
	Items     []Item    `json:"items"`
	CreatedAt time.Time `json:"created_at"`
}
