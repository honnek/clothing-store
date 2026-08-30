package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/shopspring/decimal"

	"github.com/honnek/lumewear-shop/services/internal/cart/domain"
)

// Repository хранит сырую корзину: соответствие uuid товара → количество.
type Repository interface {
	Add(ctx context.Context, session, uuid string, delta int32) error
	SetQty(ctx context.Context, session, uuid string, qty int32) error
	Remove(ctx context.Context, session, uuid string) error
	Items(ctx context.Context, session string) (map[string]int32, error)
}

// Catalog — порт в catalog-сервис: цена/название товара. Реализуется gRPC-клиентом.
type Catalog interface {
	GetProduct(ctx context.Context, uuid string) (domain.CatalogProduct, error)
}

type Cart struct {
	repo    Repository
	catalog Catalog
}

func New(repo Repository, catalog Catalog) *Cart {
	return &Cart{repo: repo, catalog: catalog}
}

func (c *Cart) AddItem(ctx context.Context, session, uuid string, qty int32) (domain.Cart, error) {
	if qty <= 0 {
		return domain.Cart{}, domain.ErrInvalidQuantity
	}
	// Цену в корзине не храним — проверяем лишь, что товар вообще существует.
	if _, err := c.catalog.GetProduct(ctx, uuid); err != nil {
		return domain.Cart{}, err
	}
	if err := c.repo.Add(ctx, session, uuid, qty); err != nil {
		return domain.Cart{}, err
	}
	return c.GetCart(ctx, session)
}

func (c *Cart) UpdateItem(ctx context.Context, session, uuid string, qty int32) (domain.Cart, error) {
	if qty <= 0 {
		// Нулевое/отрицательное количество трактуем как удаление позиции.
		if err := c.repo.Remove(ctx, session, uuid); err != nil {
			return domain.Cart{}, err
		}
		return c.GetCart(ctx, session)
	}
	if _, err := c.catalog.GetProduct(ctx, uuid); err != nil {
		return domain.Cart{}, err
	}
	if err := c.repo.SetQty(ctx, session, uuid, qty); err != nil {
		return domain.Cart{}, err
	}
	return c.GetCart(ctx, session)
}

func (c *Cart) RemoveItem(ctx context.Context, session, uuid string) (domain.Cart, error) {
	if err := c.repo.Remove(ctx, session, uuid); err != nil {
		return domain.Cart{}, err
	}
	return c.GetCart(ctx, session)
}

// GetCart собирает корзину: тянет количества из repo и обогащает их ценой/названием
// из catalog, считая построчные и общую суммы. Товар, пропавший из каталога, в выдачу
// не попадает.
func (c *Cart) GetCart(ctx context.Context, session string) (domain.Cart, error) {
	qtys, err := c.repo.Items(ctx, session)
	if err != nil {
		return domain.Cart{}, err
	}

	uuids := make([]string, 0, len(qtys))
	for uuid := range qtys {
		uuids = append(uuids, uuid)
	}
	sort.Strings(uuids)

	items := make([]domain.Item, 0, len(uuids))
	total := decimal.Zero
	for _, uuid := range uuids {
		p, err := c.catalog.GetProduct(ctx, uuid)
		if errors.Is(err, domain.ErrProductNotFound) {
			continue
		}
		if err != nil {
			return domain.Cart{}, err
		}

		price, err := decimal.NewFromString(p.Price)
		if err != nil {
			return domain.Cart{}, fmt.Errorf("parse price %q: %w", p.Price, err)
		}

		qty := qtys[uuid]
		line := price.Mul(decimal.NewFromInt(int64(qty)))
		total = total.Add(line)

		items = append(items, domain.Item{
			ProductUUID: uuid,
			Title:       p.Title,
			UnitPrice:   p.Price,
			Quantity:    qty,
			LineTotal:   line.StringFixed(2),
		})
	}

	return domain.Cart{SessionID: session, Items: items, Total: total.StringFixed(2)}, nil
}
