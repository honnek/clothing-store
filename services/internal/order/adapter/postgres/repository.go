package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/honnek/lumewear-shop/services/internal/order/adapter/postgres/orderdb"
	"github.com/honnek/lumewear-shop/services/internal/order/domain"
)

const (
	outboxAggregate   = "order"
	eventOrderCreated = "order.created"
)

type Repository struct {
	pool *pgxpool.Pool
	q    *orderdb.Queries
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, q: orderdb.New(pool)}
}

// Checkout оформляет заказ ровно один раз на ключ идемпотентности.
//
// Быстрый путь — ключ уже известен, отдаём созданный ранее заказ. Иначе идём в
// транзакцию; если параллельный запрос с тем же ключом успел закоммититься между
// проверкой и нашей вставкой, мы получим 23505, откатимся целиком (включая резерв
// остатков) и вернём его заказ.
func (r *Repository) Checkout(ctx context.Context, req domain.CheckoutRequest, lines []domain.CheckoutLine) (domain.Order, error) {
	switch o, err := r.byIdempotencyKey(ctx, req.IdempotencyKey); {
	case err == nil:
		return o, nil
	case !errors.Is(err, pgx.ErrNoRows):
		return domain.Order{}, err
	}

	o, err := r.checkoutTx(ctx, req, lines)
	if isUniqueViolation(err) {
		return r.byIdempotencyKey(ctx, req.IdempotencyKey)
	}
	return o, err
}

func (r *Repository) checkoutTx(ctx context.Context, req domain.CheckoutRequest, lines []domain.CheckoutLine) (domain.Order, error) {
	uuids := make([]pgtype.UUID, 0, len(lines))
	for _, l := range lines {
		var u pgtype.UUID
		if err := u.Scan(l.ProductUUID); err != nil {
			return domain.Order{}, fmt.Errorf("checkout %s: %w", l.ProductUUID, domain.ErrProductNotFound)
		}
		uuids = append(uuids, u)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Order{}, fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // после Commit это no-op

	q := r.q.WithTx(tx)

	locked, err := q.LockProductsForCheckout(ctx, uuids)
	if err != nil {
		return domain.Order{}, fmt.Errorf("lock products: %w", err)
	}
	products := make(map[string]orderdb.LockProductsForCheckoutRow, len(locked))
	for _, p := range locked {
		products[p.Uuid] = p
	}

	items := make([]domain.Item, 0, len(lines))
	total := decimal.Zero
	for _, l := range lines {
		p, ok := products[l.ProductUUID]
		if !ok {
			return domain.Order{}, fmt.Errorf("checkout %s: %w", l.ProductUUID, domain.ErrProductNotFound)
		}
		if p.Stock < l.Quantity {
			return domain.Order{}, &domain.StockError{
				ProductUUID: l.ProductUUID,
				Requested:   l.Quantity,
				Available:   p.Stock,
			}
		}

		price, err := decimal.NewFromString(p.Price)
		if err != nil {
			return domain.Order{}, fmt.Errorf("parse price %q: %w", p.Price, err)
		}
		line := price.Mul(decimal.NewFromInt(int64(l.Quantity)))
		total = total.Add(line)

		items = append(items, domain.Item{
			ProductID:   p.ID,
			ProductUUID: p.Uuid,
			Title:       p.Title,
			UnitPrice:   p.Price,
			Quantity:    l.Quantity,
			LineTotal:   line.StringFixed(2),
		})
	}

	for _, it := range items {
		if err := q.ReserveStock(ctx, orderdb.ReserveStockParams{Qty: it.Quantity, ID: it.ProductID}); err != nil {
			return domain.Order{}, fmt.Errorf("reserve stock for %d: %w", it.ProductID, err)
		}
	}

	totalPrice := total.InexactFloat64()
	head, err := q.InsertOrder(ctx, orderdb.InsertOrderParams{
		OwnerID:    req.OwnerID,
		Status:     int32(domain.StatusCreated),
		TotalPrice: &totalPrice,
	})
	if err != nil {
		return domain.Order{}, fmt.Errorf("insert order: %w", err)
	}

	for _, it := range items {
		var price pgtype.Numeric
		if err := price.Scan(it.UnitPrice); err != nil {
			return domain.Order{}, fmt.Errorf("price %q: %w", it.UnitPrice, err)
		}
		err := q.InsertOrderProduct(ctx, orderdb.InsertOrderProductParams{
			AppOrderID:  head.ID,
			ProductID:   it.ProductID,
			Quantity:    it.Quantity,
			PricePerOne: price,
		})
		if err != nil {
			return domain.Order{}, fmt.Errorf("insert order product %d: %w", it.ProductID, err)
		}
	}

	if err := q.InsertIdempotencyKey(ctx, orderdb.InsertIdempotencyKeyParams{
		Key:       req.IdempotencyKey,
		SessionID: req.SessionID,
		OrderID:   head.ID,
	}); err != nil {
		return domain.Order{}, err
	}

	order := domain.Order{
		ID:        head.ID,
		OwnerID:   req.OwnerID,
		Status:    domain.StatusCreated,
		Total:     total.StringFixed(2),
		CreatedAt: head.CreatedAt.Time,
		UpdatedAt: head.UpdatedAt.Time,
		Items:     items,
	}

	payload, err := json.Marshal(domain.OrderCreated{
		OrderID:   order.ID,
		SessionID: req.SessionID,
		OwnerID:   req.OwnerID,
		Total:     order.Total,
		Items:     items,
		CreatedAt: order.CreatedAt,
	})
	if err != nil {
		return domain.Order{}, fmt.Errorf("marshal event: %w", err)
	}
	if err := q.InsertOutboxEvent(ctx, orderdb.InsertOutboxEventParams{
		AggregateType: outboxAggregate,
		AggregateID:   fmt.Sprint(order.ID),
		EventType:     eventOrderCreated,
		Payload:       payload,
	}); err != nil {
		return domain.Order{}, fmt.Errorf("insert outbox: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Order{}, fmt.Errorf("commit: %w", err)
	}
	return order, nil
}

func (r *Repository) Get(ctx context.Context, id int32) (domain.Order, error) {
	return r.get(ctx, r.q, id)
}

func (r *Repository) List(ctx context.Context, f domain.OrderFilter, p domain.Page) (domain.OrderList, error) {
	var status *int32
	if f.Status != nil {
		s := int32(*f.Status)
		status = &s
	}

	rows, err := r.q.ListOrders(ctx, orderdb.ListOrdersParams{
		OwnerID: f.OwnerID,
		Status:  status,
		Lim:     p.Limit,
		Off:     p.Offset,
	})
	if err != nil {
		return domain.OrderList{}, fmt.Errorf("list orders: %w", err)
	}

	total, err := r.q.CountOrders(ctx, orderdb.CountOrdersParams{OwnerID: f.OwnerID, Status: status})
	if err != nil {
		return domain.OrderList{}, fmt.Errorf("count orders: %w", err)
	}

	out := make([]domain.Order, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.Order{
			ID:        row.ID,
			OwnerID:   row.OwnerID,
			Status:    domain.Status(row.Status),
			Total:     row.Total,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		})
	}
	return domain.OrderList{Items: out, Total: total}, nil
}

// UpdateStatus держит строку заказа под FOR UPDATE, пока проверяет переход:
// тот же заказ параллельно правит админка монолита, и без блокировки два перехода
// могут разойтись с одного и того же исходного статуса.
func (r *Repository) UpdateStatus(ctx context.Context, id int32, to domain.Status) (domain.Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Order{}, fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // после Commit это no-op

	q := r.q.WithTx(tx)

	current, err := q.LockOrderStatus(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	if err != nil {
		return domain.Order{}, fmt.Errorf("lock order: %w", err)
	}

	from := domain.Status(current)
	if !from.CanTransitionTo(to) {
		return domain.Order{}, &domain.TransitionError{From: from, To: to}
	}

	if err := q.UpdateOrderStatus(ctx, orderdb.UpdateOrderStatusParams{Status: int32(to), ID: id}); err != nil {
		return domain.Order{}, fmt.Errorf("update status: %w", err)
	}

	order, err := r.get(ctx, q, id)
	if err != nil {
		return domain.Order{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Order{}, fmt.Errorf("commit: %w", err)
	}
	return order, nil
}

func (r *Repository) byIdempotencyKey(ctx context.Context, key string) (domain.Order, error) {
	id, err := r.q.OrderIDByIdempotencyKey(ctx, key)
	if err != nil {
		return domain.Order{}, err
	}
	return r.get(ctx, r.q, id)
}

func (r *Repository) get(ctx context.Context, q *orderdb.Queries, id int32) (domain.Order, error) {
	row, err := q.GetOrder(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	if err != nil {
		return domain.Order{}, fmt.Errorf("get order: %w", err)
	}

	rows, err := q.ListOrderItems(ctx, id)
	if err != nil {
		return domain.Order{}, fmt.Errorf("list order items: %w", err)
	}

	items := make([]domain.Item, 0, len(rows))
	for _, it := range rows {
		price, err := decimal.NewFromString(it.PricePerOne)
		if err != nil {
			return domain.Order{}, fmt.Errorf("parse price %q: %w", it.PricePerOne, err)
		}
		items = append(items, domain.Item{
			ProductID:   it.ProductID,
			ProductUUID: it.Uuid,
			Title:       it.Title,
			UnitPrice:   it.PricePerOne,
			Quantity:    it.Quantity,
			LineTotal:   price.Mul(decimal.NewFromInt(int64(it.Quantity))).StringFixed(2),
		})
	}

	return domain.Order{
		ID:        row.ID,
		OwnerID:   row.OwnerID,
		Status:    domain.Status(row.Status),
		Total:     row.Total,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
		Items:     items,
	}, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
