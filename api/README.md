# Go Backend with PostgreSQL Database

This is a Go backend service using Gin framework and PostgreSQL database with `lib/pq` driver.

`docker build -t watson-docker .`
`docker run --rm -p 8080:8080 watson-docker`

Create a migration based on this template: https://github.com/golang-migrate/migrate/blob/master/database/postgres/TUTORIAL.md

`migrate create -ext sql -dir db/migrations -seq create_users_table`
`migrate -database ${DATABASE_URL} -path db/migrations up`


docker build -t us-central1-docker.pkg.dev/watson-465400/watson-registry/watson-go-api:latest .
