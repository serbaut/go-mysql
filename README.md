# go-mysql

A pure Go MySQL driver for database/sql.

Requires Go >= 1.0.3 and MySQL >= 5.0.

## Data Source Name Format

    mysql[s]://[user[:password]][@host][:port][/database][?param&...]

* use mysqls:// to establish an SSL connection
* user defaults to root
* password defaults to blank
* host defaults to localhost
* port defaults to 3306
* use query paramteter ssl-insecure-skip-verify to skip SSL certificate verification
* use query paramteter debug to log request and MySQL warnings to stdout

## Features

* longtext and longblob > 16MB support
* SSL support

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

Note: The new go scheduler (as of 2013-03-13) needs test -cpu=2 to
give good benchmark results. Dmitriy is working on it.
