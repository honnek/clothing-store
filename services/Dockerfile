# Общий multi-stage билд для всех Go-сервисов. Нужный бинарь выбирается через
# --build-arg SERVICE=<каталог в cmd/>.
ARG GO_VERSION=1.21

FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG SERVICE=probe
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/app ./cmd/${SERVICE}

FROM gcr.io/distroless/static-debian12:nonroot AS final
COPY --from=build /out/app /app
ENTRYPOINT ["/app"]
