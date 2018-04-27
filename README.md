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
        Db password to connect with. Requires dbUser. Takes precedence over vault methods
  -dbPort string
    	Db port to connect to. (default "3306")
  -dbUser string
        Db user to connect with. Requires dbPass. Takes precedence over vault methods
  -vaultAddr string
        Vault address to use. (default "http://localhost:8200")
  -vaultRoleId string
        Role ID to access vault with. Requires vaultSecretId.
  -vaultSecretId string
        Secret ID to access vault with. Requires vaultRoleId.
  -vaultToken string
        Token to access vault with. Takes precedence over vaultRoleId and vaultSecredId.
```

Parameters can also be passed using environment variables (parameter name or
uppercase version), i.e. to start with actual credb credentials:

```
DBUSER=user DBPASS=pass DBNAME=db DBHOST=host DBPORT=port ./dbwebapp
```

Or:

```
dbUser=user dbPass=pass dbName=db dbHost=host dbPort=port ./dbwebapp
```

## Usage with vault

We support two ways of vault integration:

* tokens ([Vault Tokens](https://www.vaultproject.io/docs/concepts/tokens.html))
* app roles ([Vault AppRole Auth Method](https://www.vaultproject.io/docs/auth/approle.html))

### Vault tokens

To run with an explicit vault token (to obtain db credentials from vault):
```
VAULTTOKEN=76b0d0fa-e7ac-f29e-d3e2-d1b4fc98300c ./dbwebapp
```
In this case the secret is expected to be under *secrets/dbwebapp username=user password=pass*

The required secret engine in vault is kv:
```
vault secrets enable -path=secrets kv
vault kv put secrets/dbwebapp username=dbwebapp password=dbwebapp
```
The according policy would be:
```
path "secrets/dbwebapp" {
  capabilities = ["read"]
}
```
Created by:
```
vault policy write dbwebapp /policies/dbwebapp-policy.hcl
```

Tokens can be created by:
```
vault token create -policy=dbwebapp
```

The result of this command can be passed to the application via the VAULTTOKEN parameter.

### AppRole auth method
Similarly for the approle method we need first need to specify a policy:
```
path "auth/approle/login" {
  capabilities = ["create", "read"]
}

path "database/creds/vault-mysql-role" {
  capabilities = ["read"]
}
```
And to enable the auth backend, secrets engine and create a role:

```
vault write auth/approle/role/dbwebapp policies="dbwebapp" role_id="dbrole"
vault write auth/approle/role/dbwebapp/custom-secret-id secret_id=testsecret1
vault secrets enable database
vault write database/config/mysql-database \
    plugin_name=mysql-database-plugin \
    connection_url="{{username}}:{{password}}@tcp(mysql-server:3306)/" \
    allowed_roles="vault-mysql-role" \
    username="vault" \
    password="vault"
vault write database/roles/vault-mysql-role \
    db_name=mysql-database \
    creation_statements="CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';GRANT SELECT ON dbwebappdb.* TO '{{name}}'@'%';" \
    default_ttl="1h" \
    max_ttl="24h"
```

Finally, MySQL needs to allow vault access to generate credentials on demand:
```
CREATE DATABASE dbwebappdb;

CREATE USER 'vault'@'%' IDENTIFIED BY 'vault';
GRANT ALL PRIVILEGES ON dbwebappdb.* TO 'vault'@'%' WITH GRANT OPTION;
GRANT CREATE USER ON *.* to 'vault'@'%';
```

Usage with approle credentials (to create db credentials on thefly via the db secrets engine) :
```
VAULTADDR=http://localhost:8200 VAULTROLEID=roleId VAULTSECRETID=secretId ./dbwebapp
```
The role id and secred id are used to authenitcate at *auth/approle/login* and
the token returned from this call will be used to acquire temporary db
credentials from *database/creds/vault-mysql-role*.

Leases on renewable tokens are renewed regularly.

## Running/building

On linux systems it is sufficient to run *make* to build the go binary and a
minimal docker container.
