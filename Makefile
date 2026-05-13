.PHONY: up down build test test-python test-go test-ts lint clean logs health

up:
	docker compose up -d --build

down:
	docker compose down

build:
	docker compose build

test: test-python test-go test-ts

test-python:
	cd api-gateway && pip install -r requirements.txt -q && pytest -v

test-go:
	cd analytics-engine && go test -v ./...

test-ts:
	cd notification-service && npm install --silent && npm test

lint: lint-python lint-go lint-ts

lint-python:
	cd api-gateway && flake8 --max-line-length=120 --exclude=__pycache__ .

lint-go:
	cd analytics-engine && go vet ./...

lint-ts:
	cd notification-service && npm run lint

logs:
	docker compose logs -f

health:
	@echo "API Gateway:"
	@curl -s http://localhost:8080/health | python3 -m json.tool 2>/dev/null || echo "  unreachable"
	@echo "Analytics Engine:"
	@curl -s http://localhost:8081/health | python3 -m json.tool 2>/dev/null || echo "  unreachable"
	@echo "Notification Service:"
	@curl -s http://localhost:8082/health | python3 -m json.tool 2>/dev/null || echo "  unreachable"

clean:
	docker compose down -v --rmi local
	rm -rf notification-service/node_modules notification-service/dist
	rm -rf api-gateway/__pycache__ api-gateway/.pytest_cache
