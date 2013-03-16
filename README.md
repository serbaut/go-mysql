# go-mysql

A pure Go MySQL driver for database/sql.

Requires Go >= 1.0.3 and MySQL >= 5.0.

## Data Source Name Format

    mysql[s]://[user[:password]][@host][:port][/database][?param&...]

* use `mysqls://` to establish an SSL connection
* `user` defaults to root
* `password` defaults to blank
* `host` defaults to localhost (use `(unix)` for unix domain sockets)
* `port` defaults to 3306

Parameters

* `allow-insecure-local-infile` : allow `LOAD DATA LOCAL INFILE`
* `ssl-insecure-skip-verify` : skip SSL certificate verification
* `socket` : unix domain socket (default `/var/run/mysqld/mysqld.sock`)
* `debug` : log requests and MySQL warnings to stdout

Examples

    mysql://gopher1@localhost
    mysql://gopher2:secret@localhost:3306/test?debug
    mysqls://gopher1@localhost?ssl-insecure-skip-verify
    mysql://gopher2:secret@(unix)/test?socket=/var/lib/mysql/mysql.sock

## Support for

* longtext and longblob > 16MB
* LOAD DATA LOCAL INFILE
* SSL

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
