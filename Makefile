generate:
	go run cmd/sql-gen/main.go -root ./example -package sql -out stdout > example/sql.gen.go