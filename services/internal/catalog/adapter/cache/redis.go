package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/honnek/ranked-choice-shop/services/internal/catalog/domain"
	"github.com/honnek/ranked-choice-shop/services/internal/catalog/usecase"
)

// Cache — read-through декоратор над репозиторием. Каталог редко меняется и пишется
// только из PHP, поэтому свежесть держим на TTL, без явной инвалидации. Если redis
// недоступен — молча идём в исходный репозиторий: кеш не должен ронять выдачу.
type Cache struct {
	inner usecase.Repository
	rdb   *goredis.Client
	ttl   time.Duration
}

func New(inner usecase.Repository, rdb *goredis.Client, ttl time.Duration) *Cache {
	return &Cache{inner: inner, rdb: rdb, ttl: ttl}
}

func (c *Cache) GetProduct(ctx context.Context, uuid string) (domain.Product, error) {
	key := "catalog:product:" + uuid

	var p domain.Product
	if c.read(ctx, key, &p) {
		return p, nil
	}

	p, err := c.inner.GetProduct(ctx, uuid)
	if err != nil {
		return domain.Product{}, err
	}
	c.write(ctx, key, p)
	return p, nil
}

func (c *Cache) ListCategories(ctx context.Context) ([]domain.Category, error) {
	key := "catalog:categories"

	var cats []domain.Category
	if c.read(ctx, key, &cats) {
		return cats, nil
	}

	cats, err := c.inner.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	c.write(ctx, key, cats)
	return cats, nil
}

func (c *Cache) ListProducts(ctx context.Context, f domain.ProductFilter, p domain.Page) (domain.ProductList, error) {
	key := productsKey(f, p)

	var list domain.ProductList
	if c.read(ctx, key, &list) {
		return list, nil
	}

	list, err := c.inner.ListProducts(ctx, f, p)
	if err != nil {
		return domain.ProductList{}, err
	}
	c.write(ctx, key, list)
	return list, nil
}

// read возвращает true только при попадании в кеш. Любая ошибка redis или промах —
// false, и вызывающий идёт в исходный репозиторий.
func (c *Cache) read(ctx context.Context, key string, dst any) bool {
	b, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return false
	}
	return json.Unmarshal(b, dst) == nil
}

func (c *Cache) write(ctx context.Context, key string, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	_ = c.rdb.Set(ctx, key, b, c.ttl).Err()
}

func productsKey(f domain.ProductFilter, p domain.Page) string {
	cat, pub, search := "*", "*", "*"
	if f.CategoryID != nil {
		cat = fmt.Sprint(*f.CategoryID)
	}
	if f.Published != nil {
		pub = fmt.Sprint(*f.Published)
	}
	if f.Search != nil {
		search = *f.Search
	}
	return fmt.Sprintf("catalog:products:cat=%s:pub=%s:q=%s:lim=%d:off=%d", cat, pub, search, p.Limit, p.Offset)
}
