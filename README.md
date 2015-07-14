# Project Cars Dedicated Server History

## How to get go-pcdsh?

```
go get github.com/mattias/go-pcdsh
```

## How to migrate database?

Install this tool:
```
go get github.com/rubenv/sql-migrate/...
```
Copy dbconfig.yml.example to dbconfig.yml.
dbconfig.yml needs to be available in the current working directory when performing sql-migrate
Make sure the migrations directory is also in the current working directory.

Then do the following commands:
```
sql-migrate up
```

For more documentation about sql-migrate: https://github.com/rubenv/sql-migrate

## How to run go-pcdsh?

Copy conf.json.example to conf.json.
Then change the configuration in conf.json and dbconfig.yml.
conf.json needs to be in the same directory as the executable.

Next up just run go-pcdsh.

## What can it do?

Visit http://localhost:8080/session to see all available games

Visit http://localhost:8080/session/compiled/{game-id} to see the game compiled into a nice structure, game-id is taken from the /session list
Example: http://localhost:8080/session/compiled/1