.PHONY: dev build test lint typecheck clean migrate \
  go-build go-vet go-test go-fmt go-race

dev:
	pnpm dev

build:
	pnpm build

test:
	pnpm test

lint:
	pnpm lint

typecheck:
	pnpm typecheck

clean:
	pnpm clean

migrate:
	go run ./apps/api/cmd/migrator

go-build:
	cd apps/api && go build ./...

go-vet:
	cd apps/api && go vet ./...

go-test:
	cd apps/api && go test ./...

go-race:
	cd apps/api && go test -race ./...

go-fmt:
	cd apps/api && gofmt -w .
