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
	go build github.com/testra/testra/apps/api/...

go-vet:
	go vet github.com/testra/testra/apps/api/...

go-test:
	go test github.com/testra/testra/apps/api/...

go-race:
	go test -race github.com/testra/testra/apps/api/...

go-fmt:
	gofmt -w apps/api
