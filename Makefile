.PHONY: cover demo lint test test-down test-up

cover:
	go tool cover -html=cover.out

demo: test-up
	cockroach sql --insecure -e 'create database if not exists certmagic; create user if not exists certmagic;'
	cockroach sql --insecure -e 'grant select, insert, update, delete on database certmagic to certmagic;'
	cockroach sql --insecure -d certmagic <tables.sql
	go run example/main.go

lint:
	golangci-lint run

test: test-up
	go test -p 1 -coverprofile=cover.out -mod=readonly .

test-down:
	docker-compose -p certmagic-storage-crdb-test -f docker-compose-test.yaml down

test-up:
	docker-compose -p certmagic-storage-crdb-test -f docker-compose-test.yaml up -d
