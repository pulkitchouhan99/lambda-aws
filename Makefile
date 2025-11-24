.PHONY: up down seed invoke-cmd invoke-worker

up:
	docker-compose up -d

down:
	docker-compose down

seed:
	@echo "Seeding database..."
	@go run scripts/seed/seed.go

invoke-cmd:
	@echo "Invoking command lambda..."
	@# Example: make invoke-cmd func=createPatient payload='{"name":"John Doe"}'
	aws --endpoint-url=http://localhost:4566 lambda invoke --function-name $(func) --payload '$(payload)' response.json

invoke-worker:
	@echo "Invoking worker..."
	@# Example: make invoke-worker func=screeningEventWorker payload='{"event":"screening_created"}'
	aws --endpoint-url=http://localhost:4566 lambda invoke --function-name $(func) --payload '$(payload)' response.json

deploy:
	sam build
	sam deploy --stack-name local-stack --endpoint-url http://localhost:4566 --no-confirm-changeset
