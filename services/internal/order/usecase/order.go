package usecase

import (
	"context"
	"errors"
	"log/slog"
	"sort"

	"github.com/honnek/lumewear-shop/services/internal/order/domain"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// Repository — то, что usecase ждёт от хранилища. Checkout здесь одной операцией не от
// лени: резерв остатков, заказ и outbox обязаны лечь в одну транзакцию, а границу
// транзакции знает только адаптер.
type Repository interface {
	Checkout(ctx context.Context, req domain.CheckoutRequest, lines []domain.CheckoutLine) (domain.Order, error)
	OrderByIdempotencyKey(ctx context.Context, key string) (domain.Order, error)
	Get(ctx context.Context, id int32) (domain.Order, error)
	List(ctx context.Context, f domain.OrderFilter, p domain.Page) (domain.OrderList, error)
	UpdateStatus(ctx context.Context, id int32, to domain.Status) (domain.Order, error)
}

// Cart — порт в cart-service: состав корзины и её очистка после оформления.
type Cart interface {
	Items(ctx context.Context, session string) (map[string]int32, error)
	Clear(ctx context.Context, session string) error
}

type Order struct {
	repo Repository
	cart Cart
	log  *slog.Logger
}

func New(repo Repository, cart Cart, log *slog.Logger) *Order {
	return &Order{repo: repo, cart: cart, log: log}
}

func (o *Order) Checkout(ctx context.Context, req domain.CheckoutRequest) (domain.Order, error) {
	if req.IdempotencyKey == "" {
		return domain.Order{}, domain.ErrIdempotencyKeyRequired
	}

	qtys, err := o.cart.Items(ctx, req.SessionID)
	if err != nil {
		return domain.Order{}, err
	}

	lines := make([]domain.CheckoutLine, 0, len(qtys))
	for uuid, qty := range qtys {
		if qty <= 0 {
			continue
		}
		lines = append(lines, domain.CheckoutLine{ProductUUID: uuid, Quantity: qty})
	}
	if len(lines) == 0 {
		// Пустая корзина у ретрая — обычное дело: первый запрос успел оформить заказ
		// и почистить её, а ответ до клиента не дошёл. Ключ идемпотентности решает,
		// какой из двух случаев перед нами.
		order, err := o.repo.OrderByIdempotencyKey(ctx, req.IdempotencyKey)
		switch {
		case err == nil:
			return order, nil
		case errors.Is(err, domain.ErrOrderNotFound):
			return domain.Order{}, domain.ErrEmptyCart
		default:
			return domain.Order{}, err
		}
	}
	// Порядок из map случайный, а состав заказа должен быть воспроизводимым.
	sort.Slice(lines, func(i, j int) bool { return lines[i].ProductUUID < lines[j].ProductUUID })

	order, err := o.repo.Checkout(ctx, req, lines)
	if err != nil {
		return domain.Order{}, err
	}

	// Корзина — не источник истины, заказ уже зафиксирован. Если redis не ответил,
	// валить успешный checkout нельзя: клиент увидит ошибку и уйдёт в ретрай.
	if err := o.cart.Clear(ctx, req.SessionID); err != nil {
		o.log.WarnContext(ctx, "cart not cleared after checkout",
			slog.String("session_id", req.SessionID),
			slog.Int("order_id", int(order.ID)),
			slog.Any("error", err))
	}

	return order, nil
}

func (o *Order) Get(ctx context.Context, id int32) (domain.Order, error) {
	return o.repo.Get(ctx, id)
}

func (o *Order) List(ctx context.Context, f domain.OrderFilter, p domain.Page) (domain.OrderList, error) {
	if p.Limit <= 0 || p.Limit > maxLimit {
		p.Limit = defaultLimit
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
	return o.repo.List(ctx, f, p)
}

func (o *Order) UpdateStatus(ctx context.Context, id int32, to domain.Status) (domain.Order, error) {
	if !to.Valid() {
		return domain.Order{}, domain.ErrInvalidStatusTransition
	}
	return o.repo.UpdateStatus(ctx, id, to)
}
