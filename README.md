# go-mysql

A pure Go MySQL driver for database/sql.

Requires Go >= 1.0.3 and MySQL >= 4.1

## Data Source Name

### Format

    mysql[s]://[user[:password]][@host][:port][/database][?param&...]

* `mysqls://` establishes an SSL connection
* `user` defaults to root
* `password` defaults to blank
* `host` defaults to localhost (use `(unix)` for unix domain sockets)
* `port` defaults to 3306

### Parameters

* `strict` : treat MySQL warnings as errors
* `allow-insecure-local-infile` : allow `LOAD DATA LOCAL INFILE`
* `ssl-insecure-skip-verify` : skip SSL certificate verification
* `socket` : unix domain socket (default `/var/run/mysqld/mysqld.sock`)
* `debug` : log requests and MySQL warnings to the standard logger
* `charset` : set connection character set (read note below)

### Examples

    mysql://gopher1@localhost
    mysql://gopher2:secret@localhost:3306/test?strict&debug
    mysqls://gopher1@localhost?ssl-insecure-skip-verify
    mysql://gopher2:secret@(unix)/test?socket=/var/lib/mysql/mysql.sock

## Notes

### About Time

A zero time.Time argument to Query/Exec is treated as a MySQL zero
timestamp (0000-00-00 00:00:00). A MySQL zero timestamp is returned as
a Go zero time.

Timestamps in MySQL are assumed to be in UTC. time.Time arguments are
stored as UTC and returned as UTC.

### Character Set

Strings are by default UTF-8 encoded in the MySQL connection; they are
automatically converted by the MySQL server as needed.

It is however common in legacy MySQL implementations (notably with php
clients) to have incorrectly encoded strings in the database and you
may need to trick the server not to translate to/from UTF-8. The
`charset` parameter allows you to set the character set on a connection
basis (a `SET NAMES` statement is executed on connect). Please review
http://mysql.rjweb.org/doc.php/charcoll before using this option.

## Installation

    go get github.com/serbaut/go-mysql

## Usage

    import (
        "database/sql"
        _ "github.com/serbaut/go-mysql"
    )

    func main() {
        db, err := sql.Open("mysql", "mysql://gopher2:secret@localhost/mydb")
        ...
    }

## Testing

    mysql@localhost> grant all on test.* to gopher1@localhost;
    mysql@localhost> grant all on test.* to gopher2@localhost identified by 'secret';

    $ go test
