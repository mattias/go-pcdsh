# Project Cars Dedicated Server History

## How to run?

Download go-pcdsh and run it:

```
go-pcdsh
``` 

## How to build?

```
go get github.com/emicklei/go-restful
go get github.com/bndr/gopencils
go get github.com/go-sql-driver/mysql
go install github.com/mattias/go-pcdsh
```

## How to migrate database?

Install this tool:
```
go get github.com/rubenv/sql-migrate/...
```
Then do the following commands:
```
sql-migrate up
```

For more documentation about sql-migrate: https://github.com/rubenv/sql-migrate
