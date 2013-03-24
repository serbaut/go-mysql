package mysql

// driver specific test. general database/sql tests are in ./sqltest.

import (
	"./sqltest"
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

const (
	dsn1 = "mysql://gopher1@localhost?strict"
	dsn2 = "mysql://gopher2:secret@localhost:3306/test?strict"
	dsn3 = "mysqls://gopher1@localhost?strict&ssl-insecure-skip-verify"
	dsn4 = "mysql://gopher2:secret@(unix)/test?strict&socket=/var/lib/mysql/mysql.sock"
)

func TestTypes(t *testing.T) {
	db, err := sql.Open("mysql", dsn2)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var blob_size = 16 * 1024 * 1024
	var max_allowed_packet int

	if err := db.QueryRow("select @@max_allowed_packet").Scan(&max_allowed_packet); err != nil {
		t.Fatal(err)
	}
	if max_allowed_packet < blob_size {
		t.Error("max_allowed_packet=16M is needed in my.cnf to test blobs fully")
		blob_size = max_allowed_packet
	}

	var typeTests = []struct {
		def  string
		sval string
		val  interface{}
	}{
		{"varchar(6)", "'abc\000'", "abc\000"},
		{"char(6)", "'abc\000'", "abc\000"},
		{"varchar(255) character set utf8", "'åäößαβγ \u65e5\u672c\u8a9e Привет мир'", "åäößαβγ \u65e5\u672c\u8a9e Привет мир"},
		{"varchar(255) character set latin1", "'åäöß'", "åäöß"},
		{"varchar(255) character set koi8r", "'Привет мир'", "Привет мир"},
		{"tinyint", "127", int8(127)},
		{"smallint", "32767", int16(32767)},
		{"mediumint", "8388607", int32(8388607)},
		{"int", "2147483647", int32(2147483647)},
		{"int unsigned", "4294967295", uint32(4294967295)},
		{"bigint", "9223372036854775807", int64(9223372036854775807)},
		{"float", "0.123456", float32(0.123456)},
		{"double", "0.12345678901234", 0.12345678901234},
		{"decimal(7,6)", "0.123456", float32(0.123456)},
		{"bool", "true", true},
		{"bit(10)", "256", []byte{1, 0}},
		{"timestamp", "'2001-02-03 01:02:03'", time.Date(2001, 2, 3, 1, 2, 3, 0, time.UTC)},
		{"datetime", "NULL", time.Time{}},
		{"datetime", "'0000-00-00 00:00:00'", time.Time{}},
		{"datetime", "'1111-02-03 01:02:03'", time.Date(1111, 2, 3, 1, 2, 3, 0, time.UTC)},
		{"datetime", "'2001-02-03 00:02:03'", time.Date(2001, 2, 3, 1, 2, 3, 0, time.FixedZone("", +3600))},
		{"date", "'2222-02-03'", time.Date(2222, 2, 3, 0, 0, 0, 0, time.UTC)},
		{"date", "NULL", time.Time{}},
		{"date", "'0000-00-00'", time.Time{}},
		// pending support for time.Duration in database/sql:
		//{"time", "'-100:01:59'", time.Duration(-(100*time.Hour + 1*time.Minute + 59*time.Second))},
		//{"time", "NULL", time.Duration(0)},
		{"enum('a', 'b')", "'b'", "b"},
		{"set('a', 'b')", "'b'", "b"},
		{"binary(5)", "'abc'", []byte{'a', 'b', 'c', 0, 0}},
		{"blob", "'blob'", []byte("blob")},
		{"text", "'text'", "text"},
		{"varchar(8000)", "'" + strings.Repeat(".", 8000) + "'", strings.Repeat(".", 8000)},
		{"longblob", "#longblob", bytes.Repeat([]byte{'.'}, blob_size)},
		{"longtext", "#longblob", strings.Repeat(".", blob_size)},
	}

	for _, tt := range typeTests {
		for _, mode := range []string{"string", "arg", "stmt"} {
			if tt.sval == "#longblob" && (mode == "string" || testing.Short()) {
				continue
			}
			tx, err := db.Begin()
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec("drop table if exists gotest"); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(fmt.Sprintf("create temporary table gotest (null1 int, x %s, null2 int)", tt.def)); err != nil {
				t.Fatal(err)
			}

			var r *sql.Rows

			switch mode {
			case "string":
				if _, err = tx.Exec(fmt.Sprintf("insert into gotest values (null, %s, null)", tt.sval)); err != nil {
					t.Fatal(err)
				}
				if r, err = tx.Query("select * from gotest"); err != nil {
					t.Fatal(err)
				}
			case "arg":
				if _, err = tx.Exec("insert into gotest values (?, ?, ?)", nil, tt.val, nil); err != nil {
					t.Fatal(err)
				}
				if r, err = tx.Query("select * from gotest"); err != nil {
					t.Fatal(err)
				}
			case "stmt":
				var st1, st2 *sql.Stmt
				if st1, err = tx.Prepare("insert into gotest values (?, ?, ?)"); err != nil {
					t.Fatal(err)
				}
				if st2, err = tx.Prepare("select * from gotest"); err != nil {
					t.Fatal(err)
				}
				if _, err = st1.Exec(nil, tt.val, nil); err != nil {
					t.Fatal(err)
				}
				if r, err = st2.Query(); err != nil {
					t.Fatal(err)
				}
				if err = st1.Close(); err != nil {
					t.Fatal(err)
				}
				if err = st2.Close(); err != nil {
					t.Fatal(err)
				}
			}

			if !r.Next() {
				if err = r.Err(); err != nil {
					t.Fatal(err)
				}
				t.Error("expected row")
			}

			var null1, null2 sql.NullInt64

			switch want := tt.val.(type) {
			case string: // DeepEqual too slow for blobs
				var got string
				if err := r.Scan(&null1, &got, &null2); err != nil {
					t.Fatal(err)
				}
				if got != want {
					t.Errorf("%v: got %v, want %v", tt, got, want)
				}
			case []byte: // DeepEqual too slow for blobs
				var got []byte
				if err := r.Scan(&null1, &got, &null2); err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got, want) {
					t.Errorf("%v: got %v, want %v", tt, got, want)
				}
			case float32:
				var got float32
				if err := r.Scan(&null1, &got, &null2); err != nil {
					t.Fatal(err)
				}
				if fmt.Sprintf("%.6f", got) != tt.sval {
					t.Errorf("%v: got %v, want %v", tt, got, want)
				}
			case float64:
				var got float64
				if err := r.Scan(&null1, &got, &null2); err != nil {
					t.Fatal(err)
				}
				if fmt.Sprintf("%.14f", got) != tt.sval {
					t.Errorf("%v: got %v, want %v", tt, got, want)
				}
			case time.Time:
				var got time.Time
				if err := r.Scan(&null1, &got, &null2); err != nil {
					t.Fatal(err)
				}
				if got.UTC() != want.UTC() {
					t.Errorf("%v: got %v, want %v", tt, got.UTC(), want.UTC())
				}
			default:
				v := reflect.New(reflect.ValueOf(tt.val).Type())
				if err := r.Scan(&null1, v.Interface(), &null2); err != nil {
					t.Fatal(err)
				}
				if got, want := reflect.Indirect(v).Interface(), tt.val; !reflect.DeepEqual(got, want) {
					t.Errorf("%v: got %v, want %v", tt, got, want)
				}
			}
			if got, want := null1.Valid, false; got != want {
				t.Errorf("%v: got %v, want %v", tt, got, want)
			}
			if got, want := null2.Valid, false; got != want {
				t.Errorf("%v: got %v, want %v", tt, got, want)
			}

			if r.Next() {
				t.Fatal("expected exactly 1 row")
			}
			if err = r.Close(); err != nil {
				t.Fatal(err)
			}
			if err = tx.Commit(); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestUtf8mb4(t *testing.T) {
	db, err := sql.Open("mysql", dsn2)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err = db.Exec("create temporary table gotest (name varchar(255) character set utf8mb4)"); err != nil {
		if err.Error() == "ERROR 1115 (42000): Unknown character set: 'utf8mb4'" {
			t.Log("skipping utf8mb4 test, server does not support it")
			return
		}
		t.Fatal(err)
	}

	want := "\U00101234"
	var got string

	if _, err = db.Exec("insert into gotest values ('" + want + "')"); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("select * from gotest where name = ?", want).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestLoadData(t *testing.T) {
	db, err := sql.Open("mysql", dsn2+"&allow-insecure-local-infile")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	f, err := ioutil.TempFile("", "go-mysql")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("1\ttest\n")
	f.Close()
	defer os.Remove(f.Name())

	if _, err = db.Exec("create temporary table gotest (id int, name varchar(255))"); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(fmt.Sprintf("load data local infile '%s' into table gotest", f.Name())); err != nil {
		t.Fatal(err)
	}
}

func TestSSL(t *testing.T) {
	db, err := sql.Open("mysql", dsn1)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var n, s string
	if err := db.QueryRow("show variables like 'have_ssl'").Scan(&n, &s); err != nil {
		t.Fatal(err)
	}
	if s != "YES" {
		t.Log("skipping SSL test, server does not support SSL")
		return
	}

	dbs, err := sql.Open("mysql", dsn3)
	if err != nil {
		t.Fatal(err)
	}
	defer dbs.Close()
	if err := dbs.QueryRow("show status like 'ssl_version'").Scan(&n, &s); err != nil {
		t.Fatal(err)
	}
	if got, want := s, "TLSv1"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestGoroutines(t *testing.T) {
	db, err := sql.Open("mysql", dsn2)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	const N = 100
	c := make(chan int, N)
	for i := 0; i < N; i++ {
		go func(i int) {
			if err := db.QueryRow("select ?+1", i).Scan(&i); err != nil {
				t.Error(err)
			}
			c <- i
		}(i)
	}
	sum := 0
	for i := 0; i < N; i++ {
		sum += <-c
	}
	if got, want := sum, N*(N+1)/2; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestUnixSocket(t *testing.T) {
	if _, err := os.Stat("/var/lib/mysql/mysql.sock"); err != nil {
		t.Log("skipping unix domain socket test")
		return
	}
	db, err := sql.Open("mysql", dsn4)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err = db.Exec("select 1"); err != nil {
		t.Fatal(err)
	}
}

func TestStrict(t *testing.T) {
	db, err := sql.Open("mysql", dsn2)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = db.Exec("drop table if exists doesnotexist"); err != nil {
		t.Fatal(err)
	}

	if _, err = db.Exec("select 1 + 'foo'"); err == nil {
		t.Fatal("expected error")
	}

	if _, err = db.Exec("select ? + ?", 1, "foo"); err == nil {
		// mysql doesnt seem to generate warnings for this?!
		//t.Fatal("expected error")
	}

	if _, err = db.Exec("create temporary table gotest (name varchar(2) )"); err != nil {
		t.Fatal(err)
	}

	if _, err = db.Exec("insert into gotest (name) values ('overflow')"); err == nil {
		t.Fatal("expected error")
	}

	if _, err = db.Exec("insert into gotest (name) values (?)", "overflow"); err == nil {
		t.Fatal("expected error")
	}
}

func TestNullTime(t *testing.T) {
	db, err := sql.Open("mysql", dsn2)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err = db.Exec("create temporary table gotest (ts datetime, d date)"); err != nil {
		t.Fatal(err)
	}

	if _, err = db.Exec("insert into gotest set ts=null, d=null"); err != nil {
		t.Fatal(err)
	}

	if _, err = db.Exec("insert into gotest set ts='0000-00-00 00:00:00', d='0000-00-00'"); err != nil {
		t.Fatal(err)
	}

	if _, err = db.Exec("insert into gotest set ts=?, d=?", nil, nil); err != nil {
		t.Fatal(err)
	}

	if _, err = db.Exec("insert into gotest set ts=?, d=?", time.Time{}, time.Time{}); err != nil {
		t.Fatal(err)
	}

	checkZero := func(r *sql.Rows) {
		for r.Next() {
			var ts, d time.Time
			if err := r.Scan(&ts, &d); err != nil {
				t.Fatal(err)
			}
			if !ts.IsZero() {
				t.Fatalf("got %v, expected %v", ts, time.Time{})
			}
		}
		if err := r.Err(); err != nil {
			t.Fatal(err)
		}
	}

	if r, err := db.Query("select * from gotest"); err != nil {
		t.Fatal(err)
	} else {
		checkZero(r)
	}
	if r, err := db.Query("select * from gotest where ?", true); err != nil {
		t.Fatal(err)
	} else {
		checkZero(r)
	}

	var n int
	if err := db.QueryRow("select count(*) from gotest where ts = ? and d = ?", time.Time{}, time.Time{}).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("got %v, want %v zero dates", n, 2)
	}
}

func TestSuite(t *testing.T) {
	db, err := sql.Open("mysql", dsn2)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	sqltest.RunTests(t, db, sqltest.MYSQL)
}
