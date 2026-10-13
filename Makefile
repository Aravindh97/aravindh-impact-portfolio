.PHONY: test test-go test-python test-web build resume dev-core dev-intelligence dev-web compose-up clean
PY ?= python3

# Canonical Next.js + Go + Python platform.
test: test-go test-python test-web

test-go:
	gofmt -l . | (! grep .) && go vet ./... && go test -race -cover ./...

test-python:
	PYTHONPATH=. $(PY) -m pytest -q services/intelligence/tests

test-web:
	cd apps/web && npm run typecheck && npm run build

resume:
	go run ./devresume/cmd/resume -data devresume/resume.json -out apps/web/public/resume.pdf

build: resume test-web

# Run these in three terminals. All default to loopback.
dev-core:
	go run ./services/core/cmd/humanity-core
dev-intelligence:
	PYTHONPATH=. $(PY) -m uvicorn services.intelligence.app.main:app --host 127.0.0.1 --port 8091
dev-web:
	cd apps/web && npm run dev

# The only published port is host loopback 127.0.0.1:3000.
compose-up:
	docker compose up --build

clean:
	rm -rf apps/web/.next devresume/build/resume dist
