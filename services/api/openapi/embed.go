// Package openapi отдаёт сгенерированные OpenAPI-спеки сервисов, вшитые в бинарь
// (чтобы distroless-образу не нужны были внешние файлы).
package openapi

import _ "embed"

//go:embed catalog/v1/catalog.swagger.json
var Catalog []byte

//go:embed cart/v1/cart.swagger.json
var Cart []byte

//go:embed order/v1/order.swagger.json
var Order []byte
