# cf-test-postgres
This is simple test application for CF. The `cf-test-postgres` connects to PostgreSQL database and creates REST endpoints for 
adding and retrieving data to/from database.

## Configuration
In `main.go` you can find consts for database connection

```
const (
	host     = "10.0.0.4"
	port     = 5432
	user     = "postgres"
	password = "password"
	dbname   = "postgres"
)
```

## Deployment
```
cf push cf-test-postgres -u process
```

## How to add new entry to database

```
curl -X POST -H "Content-Type: application/json" -d '{"name":"John","surname":"Doe", "age":23, "email":"test@email.com", "avatar":"svg-1"}' http://cf-test-postgres.domain/user
```

## How to get all data

```
curl http://cf-test-postgres.domain/users
```

