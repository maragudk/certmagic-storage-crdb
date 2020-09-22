.PHONY: cover lint test test-down test-up

cover:
	go tool cover -html=cover.out

lint:
	golangci-lint run

test: test-up
	go test -p 1 -coverprofile=cover.out -mod=readonly .

test-down:
	docker-compose -p certmagic-storage-crdb-test -f docker-compose-test.yaml down

test-up:
	docker-compose -p certmagic-storage-crdb-test -f docker-compose-test.yaml up -d
