COMMIT_MSG :=""

postgres16: 
	docker run --name postgres16 --network bank-network -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -p 5555:5432 -d postgres:16-alpine3.19

createdb:
	docker exec -it postgres16 createdb --username=root --owner=root simplebank

dropdb:
	docker exec -it postgres16 dropdb simplebank

migrateup:
	migrate -path db/migration  -database "postgresql://root:secret@localhost:5555/simplebank?sslmode=disable" --verbose up

migratedown:
	migrate -path db/migration  -database "postgresql://root:secret@localhost:5555/simplebank?sslmode=disable" --verbose down

migrateup1:
	migrate -path db/migration  -database "postgresql://root:secret@localhost:5555/simplebank?sslmode=disable" --verbose up 1

migratedown1:
	migrate -path db/migration  -database "postgresql://root:secret@localhost:5555/simplebank?sslmode=disable" --verbose down 1

sqlc:
	sqlc generate

server:
	go run main.go

test: 
	go test -v -cover ./...

mock:
	mockgen -package mockdb -destination db/mock/store.go BankAppGo/db/sqlc Store 
	
push_git:
	git add .
	git commit -m "$(COMMIT_MSG)" 
	git push

.PHONY: postgres16 createdb dropdb migrateup migratedown migrateup1 migratedown1 sqlc test server mock push_git