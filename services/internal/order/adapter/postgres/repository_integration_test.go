//go:build integration

package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/honnek/lumewear-shop/services/internal/order/domain"
	"github.com/honnek/lumewear-shop/services/migrations"
)

const (
	redHat  = "11111111-1111-1111-1111-111111111111"
	blueHat = "22222222-2222-2222-2222-222222222222"
	oldShoe = "33333333-3333-3333-3333-333333333333"
	deleted = "44444444-4444-4444-4444-444444444444"
)

// setupRepo поднимает postgres со снимком схемы монолита и сидом, накатывает
// goose-миграции заказа и отдаёт репозиторий вместе с пулом для проверок из тестов.
func setupRepo(t *testing.T) (*Repository, *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	pg, err := tcpostgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		tcpostgres.WithDatabase("orders"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.WithInitScripts("schema.sql", "testdata/seed.sql"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	t.Cleanup(func() { _ = pg.Terminate(ctx) })

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("conn string: %v", err)
	}
	if err := migrations.Up(dsn); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)

	return New(pool), pool
}

func TestCheckout(t *testing.T) {
	repo, pool := setupRepo(t)
	ctx := context.Background()

	t.Run("reserves stock and snapshots prices", func(t *testing.T) {
		order, err := repo.Checkout(ctx, req("key-1", "s1"), []domain.CheckoutLine{
			{ProductUUID: redHat, Quantity: 2},
			{ProductUUID: blueHat, Quantity: 1},
		})
		if err != nil {
			t.Fatal(err)
		}

		if order.Status != domain.StatusCreated {
			t.Errorf("status = %s, want created", order.Status)
		}
		// 2 × 9.99 + 1 × 12.50
		if order.Total != "32.48" {
			t.Errorf("total = %s, want 32.48", order.Total)
		}
		if len(order.Items) != 2 || order.Items[0].UnitPrice != "9.99" || order.Items[0].LineTotal != "19.98" {
			t.Errorf("items = %+v", order.Items)
		}
		if got := stock(t, pool, redHat); got != 8 {
			t.Errorf("red hat stock = %d, want 8", got)
		}

		// Тот же заказ, прочитанный обратно, должен совпадать со снимком.
		back, err := repo.Get(ctx, order.ID)
		if err != nil {
			t.Fatal(err)
		}
		if back.Total != order.Total || len(back.Items) != 2 {
			t.Errorf("re-read mismatch: %+v", back)
		}
	})

	t.Run("writes order.created to outbox", func(t *testing.T) {
		order, err := repo.Checkout(ctx, req("key-outbox", "s-outbox"), []domain.CheckoutLine{
			{ProductUUID: oldShoe, Quantity: 1},
		})
		if err != nil {
			t.Fatal(err)
		}

		var eventType, payload string
		err = pool.QueryRow(ctx,
			`SELECT event_type, payload::text FROM outbox
			 WHERE aggregate_type = 'order' AND aggregate_id = $1 AND published_at IS NULL`,
			fmt.Sprint(order.ID)).Scan(&eventType, &payload)
		if err != nil {
			t.Fatalf("outbox row: %v", err)
		}
		if eventType != "order.created" {
			t.Errorf("event_type = %s", eventType)
		}

		var event domain.OrderCreated
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			t.Fatalf("payload %s: %v", payload, err)
		}
		if event.OrderID != order.ID || event.SessionID != "s-outbox" || event.Total != "5.00" {
			t.Errorf("event = %+v", event)
		}
		if len(event.Items) != 1 || event.Items[0].ProductUUID != oldShoe {
			t.Errorf("event items = %+v", event.Items)
		}
	})

	t.Run("same key returns the same order", func(t *testing.T) {
		lines := []domain.CheckoutLine{{ProductUUID: blueHat, Quantity: 1}}

		first, err := repo.Checkout(ctx, req("key-repeat", "s2"), lines)
		if err != nil {
			t.Fatal(err)
		}
		before := stock(t, pool, blueHat)

		second, err := repo.Checkout(ctx, req("key-repeat", "s2"), lines)
		if err != nil {
			t.Fatal(err)
		}
		if second.ID != first.ID {
			t.Errorf("order id = %d, want %d", second.ID, first.ID)
		}
		if after := stock(t, pool, blueHat); after != before {
			t.Errorf("stock moved on replay: %d -> %d", before, after)
		}
	})

	t.Run("lookup by idempotency key", func(t *testing.T) {
		created, err := repo.Checkout(ctx, req("key-lookup", "s2"), []domain.CheckoutLine{
			{ProductUUID: blueHat, Quantity: 1},
		})
		if err != nil {
			t.Fatal(err)
		}

		found, err := repo.OrderByIdempotencyKey(ctx, "key-lookup")
		if err != nil {
			t.Fatal(err)
		}
		if found.ID != created.ID || len(found.Items) != 1 {
			t.Errorf("got %+v, want order %d with items", found, created.ID)
		}

		if _, err := repo.OrderByIdempotencyKey(ctx, "never-used"); !errors.Is(err, domain.ErrOrderNotFound) {
			t.Errorf("err = %v, want ErrOrderNotFound", err)
		}
	})

	t.Run("not enough stock", func(t *testing.T) {
		_, err := repo.Checkout(ctx, req("key-toomuch", "s3"), []domain.CheckoutLine{
			{ProductUUID: oldShoe, Quantity: 999},
		})
		if !errors.Is(err, domain.ErrInsufficientStock) {
			t.Fatalf("err = %v, want ErrInsufficientStock", err)
		}

		var se *domain.StockError
		if !errors.As(err, &se) || se.ProductUUID != oldShoe || se.Requested != 999 {
			t.Fatalf("err = %+v, want details about %s", err, oldShoe)
		}
	})

	t.Run("deleted product is not orderable", func(t *testing.T) {
		_, err := repo.Checkout(ctx, req("key-gone", "s4"), []domain.CheckoutLine{
			{ProductUUID: deleted, Quantity: 1},
		})
		if !errors.Is(err, domain.ErrProductNotFound) {
			t.Fatalf("err = %v, want ErrProductNotFound", err)
		}
	})

	t.Run("failed line rolls back the whole order", func(t *testing.T) {
		before := stock(t, pool, redHat)

		_, err := repo.Checkout(ctx, req("key-mixed", "s5"), []domain.CheckoutLine{
			{ProductUUID: redHat, Quantity: 1},
			{ProductUUID: oldShoe, Quantity: 500},
		})
		if !errors.Is(err, domain.ErrInsufficientStock) {
			t.Fatalf("err = %v", err)
		}
		if after := stock(t, pool, redHat); after != before {
			t.Errorf("red hat stock = %d, want untouched %d", after, before)
		}

		var n int
		if err := pool.QueryRow(ctx,
			`SELECT count(*) FROM checkout_idempotency WHERE key = 'key-mixed'`).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Errorf("idempotency key persisted after rollback")
		}
	})
}

// Гвоздь фазы: остаётся одна единица товара, десять покупателей жмут «купить»
// одновременно. Пройти должен ровно один — это то, что сломано в PHP-версии.
func TestCheckoutLastUnitUnderConcurrency(t *testing.T) {
	repo, pool := setupRepo(t)
	ctx := context.Background()

	setStock(t, pool, redHat, 1)

	const buyers = 10
	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		ok     []int32
		denied int
	)

	start := make(chan struct{})
	for i := 0; i < buyers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start

			order, err := repo.Checkout(ctx,
				req(fmt.Sprintf("race-%d", i), fmt.Sprintf("session-%d", i)),
				[]domain.CheckoutLine{{ProductUUID: redHat, Quantity: 1}})

			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				ok = append(ok, order.ID)
			case errors.Is(err, domain.ErrInsufficientStock):
				denied++
			default:
				t.Errorf("unexpected error: %v", err)
			}
		}(i)
	}
	close(start)
	wg.Wait()

	if len(ok) != 1 {
		t.Fatalf("%d successful checkouts, want exactly 1 (orders %v)", len(ok), ok)
	}
	if denied != buyers-1 {
		t.Errorf("denied = %d, want %d", denied, buyers-1)
	}
	if got := stock(t, pool, redHat); got != 0 {
		t.Errorf("stock = %d, want 0", got)
	}

	var orders int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM order_product WHERE product_id = (SELECT id FROM product WHERE uuid = $1)`,
		redHat).Scan(&orders); err != nil {
		t.Fatal(err)
	}
	if orders != 1 {
		t.Errorf("order_product rows = %d, want 1", orders)
	}
}

// Один ключ идемпотентности, два параллельных запроса: второй обязан получить
// тот же заказ, а не второй такой же.
func TestCheckoutSameKeyConcurrently(t *testing.T) {
	repo, pool := setupRepo(t)
	ctx := context.Background()

	var (
		wg  sync.WaitGroup
		mu  sync.Mutex
		ids []int32
	)

	start := make(chan struct{})
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start

			order, err := repo.Checkout(ctx, req("one-key", "s-race"),
				[]domain.CheckoutLine{{ProductUUID: blueHat, Quantity: 1}})
			if err != nil {
				t.Errorf("checkout: %v", err)
				return
			}
			mu.Lock()
			ids = append(ids, order.ID)
			mu.Unlock()
		}()
	}
	close(start)
	wg.Wait()

	if len(ids) != 2 || ids[0] != ids[1] {
		t.Fatalf("order ids = %v, want two identical", ids)
	}
	if got := stock(t, pool, blueHat); got != 9 {
		t.Errorf("stock = %d, want 9 (reserved once)", got)
	}
}

func TestGetAndList(t *testing.T) {
	repo, _ := setupRepo(t)
	ctx := context.Background()

	owner := int32(7)
	mine := domain.CheckoutRequest{SessionID: "s-owner", IdempotencyKey: "k-owner", OwnerID: &owner}
	own, err := repo.Checkout(ctx, mine, []domain.CheckoutLine{{ProductUUID: redHat, Quantity: 1}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Checkout(ctx, req("k-guest", "s-guest"),
		[]domain.CheckoutLine{{ProductUUID: blueHat, Quantity: 1}}); err != nil {
		t.Fatal(err)
	}

	t.Run("missing order", func(t *testing.T) {
		if _, err := repo.Get(ctx, 9999); !errors.Is(err, domain.ErrOrderNotFound) {
			t.Fatalf("err = %v, want ErrOrderNotFound", err)
		}
	})

	t.Run("filter by owner", func(t *testing.T) {
		list, err := repo.List(ctx, domain.OrderFilter{OwnerID: &owner}, domain.Page{Limit: 10})
		if err != nil {
			t.Fatal(err)
		}
		if list.Total != 1 || list.Items[0].ID != own.ID {
			t.Fatalf("got %+v", list)
		}
	})

	t.Run("filter by status", func(t *testing.T) {
		created := domain.StatusCreated
		list, err := repo.List(ctx, domain.OrderFilter{Status: &created}, domain.Page{Limit: 10})
		if err != nil {
			t.Fatal(err)
		}
		if list.Total != 2 {
			t.Fatalf("total = %d, want 2", list.Total)
		}

		denied := domain.StatusDenied
		empty, err := repo.List(ctx, domain.OrderFilter{Status: &denied}, domain.Page{Limit: 10})
		if err != nil {
			t.Fatal(err)
		}
		if empty.Total != 0 {
			t.Fatalf("total = %d, want 0", empty.Total)
		}
	})
}

func TestUpdateStatus(t *testing.T) {
	repo, _ := setupRepo(t)
	ctx := context.Background()

	order, err := repo.Checkout(ctx, req("k-status", "s-status"),
		[]domain.CheckoutLine{{ProductUUID: redHat, Quantity: 1}})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("walks the happy path", func(t *testing.T) {
		for _, want := range []domain.Status{domain.StatusProcessed, domain.StatusComplected, domain.StatusDelivered} {
			got, err := repo.UpdateStatus(ctx, order.ID, want)
			if err != nil {
				t.Fatalf("-> %s: %v", want, err)
			}
			if got.Status != want {
				t.Fatalf("status = %s, want %s", got.Status, want)
			}
		}
	})

	t.Run("delivered is terminal", func(t *testing.T) {
		_, err := repo.UpdateStatus(ctx, order.ID, domain.StatusDenied)
		if !errors.Is(err, domain.ErrInvalidStatusTransition) {
			t.Fatalf("err = %v, want ErrInvalidStatusTransition", err)
		}

		var te *domain.TransitionError
		if !errors.As(err, &te) || te.From != domain.StatusDelivered {
			t.Fatalf("err = %+v", err)
		}
	})

	t.Run("skipping a step is rejected", func(t *testing.T) {
		fresh, err := repo.Checkout(ctx, req("k-skip", "s-skip"),
			[]domain.CheckoutLine{{ProductUUID: blueHat, Quantity: 1}})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := repo.UpdateStatus(ctx, fresh.ID, domain.StatusDelivered); !errors.Is(err, domain.ErrInvalidStatusTransition) {
			t.Fatalf("err = %v, want ErrInvalidStatusTransition", err)
		}
	})

	t.Run("missing order", func(t *testing.T) {
		if _, err := repo.UpdateStatus(ctx, 9999, domain.StatusProcessed); !errors.Is(err, domain.ErrOrderNotFound) {
			t.Fatalf("err = %v, want ErrOrderNotFound", err)
		}
	})
}

func req(key, session string) domain.CheckoutRequest {
	return domain.CheckoutRequest{SessionID: session, IdempotencyKey: key}
}

func stock(t *testing.T, pool *pgxpool.Pool, uuid string) int32 {
	t.Helper()
	var n int32
	if err := pool.QueryRow(context.Background(),
		`SELECT stock FROM product WHERE uuid = $1`, uuid).Scan(&n); err != nil {
		t.Fatalf("read stock: %v", err)
	}
	return n
}

func setStock(t *testing.T, pool *pgxpool.Pool, uuid string, n int32) {
	t.Helper()
	if _, err := pool.Exec(context.Background(),
		`UPDATE product SET stock = $2 WHERE uuid = $1`, uuid, n); err != nil {
		t.Fatalf("set stock: %v", err)
	}
}
