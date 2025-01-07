module main

go 1.22.3

require (
	github.com/gorilla/mux v1.8.1
	github.com/jackc/pgx/v5 v5.7.2
	github.com/stretchr/testify v1.10.0
	go.uber.org/zap v1.27.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/text v0.21.0 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require internal/dbaser v1.0.0

require internal/memo v1.0.0

require internal/middles v1.0.0

replace internal/dbaser => ./internal/dbaser

replace internal/middles => ./internal/middles

replace internal/memo => ./internal/memo
