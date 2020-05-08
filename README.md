# jsonToMySQL 

This is a simple webserver that takes a preformatted http get request and parses the JSON and puts in into a preconfigured MySQL database.
&nbsp;

## Usage
Run this server on the configured port (see below) and effect http requests against the server. The server will respond to the client with either an 'Ok' or 'Failed' response indicating a record was created/updated or the transaction failed. 
&nbsp;

## Syntax
```
$ go.exe run /path/to/jsonToMySQL.go [arguments]
```

## Files
Two files are needed to use this service 'config.json' and 'tables.json' which need to be in the go script directory.

config.json will look like this:
```
{
    "server_port": 8081,
    "mysql_username": "user1",
    "mysql_password": "secret-password",
    "mysql_server": "127.0.0.1",
    "mysql_database": "storage_DB",
    "mysql_port": 3306
}
```
**Note:** 'server_port' is the port this service will run on.
&nbsp;  

tables.json will look like this:
```
{
    "EA5CEB4C-3C7D-4098-85C6-ABC66F0E686A": {
        "Table": "Table_1",
        "Fields": ["stock", "amount", "price"],
        "Alias": ["symbol", "openQuantity", "currentPrice"]
    },
    "00EEF14B-A216-447E-978B-2312FFF3F517": {
        "Table": "Table_2",
        "Fields": ["id", "name", "vals"],
        "Alias": ["identity", "fullname", "info"]
    }
}
```
**Notes:** 
The GUIDs can be anything but will define what the request authorization is set to.

Aliases are direct maps to the fields. So for instance, if you have a json request for
the second table with 'fullname' identified, it will be input into the 'name' field
in 'Table_2' in MySQL.


## HTTP
HTTP requests can be done through curl.

Example:
```
$ curl -s -H "Authorization: 00EEF14B-A216-447E-978B-2312FFF3F517" -d '{"identity":100, "fullname": "Michel Noel", "info":"Favorite saying Cool beans", "Other":"Not captured"}' http://localhost:8081
```
&nbsp;

Will result with:
___
storage_DB->Table_2
```
┌─────┬─────────────┬────────────────────────────┐
│ id  │ name        │ vals                       │
├─────┼─────────────┼────────────────────────────┤
│ 100 │ Michel Noel │ Favorite saying Cool beans │
└─────┴─────────────┴────────────────────────────┘
```
**Note:** Tables need to exist in the MySQL database with the correct fields/types. 

## Arguments

##### Help
Use `--help` to show the help menu.

##### Verbose
Use `--verbose` to show more information on the webserver.

##### Version
Use `--version` to show the version info.

---

**Warranty:**
VENDOR MAKES NO WARRANTIES, EITHER EXPRESS OR IMPLIED, INCLUDING WITHOUT LIMITATION ANY IMPLIED WARRANTY OF MERCHANTABILITY OR FITNESS FOR A PARTICULAR PURPOSE.

Written by Michel Noel © 2020
