`migrate create -ext sql -dir database/migrations -seq create_users_table`
`migrate -database ${DATABASE_URL} -path database/migrations up`
