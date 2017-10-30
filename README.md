# Simple webapp connecting to a MySQL database

A simple webapp to test connecting to MySQL databases.

## General
In short the application does the following:

* Read parameters (also from environment variables)
* Ping the given db (with simple wait and retry logic)
* Report once a connection can be made (or exit in case it can't)
* Expose a */health* endpoint

The following parameters are supported:

```
Usage of ./dbwebapp:
  -dbHost string
    	Db host to connect to. (default "localhost")
  -dbName string
    	Db to use.
  -dbPass string
    	Db password to connect with.
  -dbPort string
    	Db port to connect to. (default "3306")
  -dbUser string
    	Db user to connect with.
```

Parameters can also be passed using environment variables, i.e.:

```
DBUSER=user DBPASS=pass DBNAME=db DBHOST=host DBPORT=port ./dbwebapp
```

Or:

```
dbUser=user dbPass=pass dbName=db dbHost=host dbPort=port ./dbwebapp
```

## Running/building

On linux systems it is sufficient to run *make* to build the go binary and a
minimal docker container.
