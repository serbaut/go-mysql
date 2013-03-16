package mysql

// from /usr/include/mysql/mysql_com.h

const (
	CLIENT_LONG_PASSWORD     = 1      /* new more secure passwords */
	CLIENT_FOUND_ROWS        = 2      /* Found instead of affected rows */
	CLIENT_LONG_FLAG         = 4      /* Get all column flags */
	CLIENT_CONNECT_WITH_DB   = 8      /* One can specify db on connect */
	CLIENT_NO_SCHEMA         = 16     /* Don't allow database.table.column */
	CLIENT_COMPRESS          = 32     /* Can use compression protocol */
	CLIENT_ODBC              = 64     /* Odbc client */
	CLIENT_LOCAL_FILES       = 128    /* Can use LOAD DATA LOCAL */
	CLIENT_IGNORE_SPACE      = 256    /* Ignore spaces before '(' */
	CLIENT_PROTOCOL_41       = 512    /* New 4.1 protocol */
	CLIENT_INTERACTIVE       = 1024   /* This is an interactive client */
	CLIENT_SSL               = 2048   /* Switch to SSL after handshake */
	CLIENT_IGNORE_SIGPIPE    = 4096   /* IGNORE sigpipes */
	CLIENT_TRANSACTIONS      = 8192   /* Client knows about transactions */
	CLIENT_RESERVED          = 16384  /* Old flag for 4.1 protocol  */
	CLIENT_SECURE_CONNECTION = 32768  /* New 4.1 authentication */
	CLIENT_MULTI_STATEMENTS  = 65536  /* Enable/disable multi-stmt support */
	CLIENT_MULTI_RESULTS     = 131072 /* Enable/disable multi-results */
)

const (
	COM_SLEEP = iota
	COM_QUIT
	COM_INIT_DB
	COM_QUERY
	COM_FIELD_LIST
	COM_CREATE_DB
	COM_DROP_DB
	COM_REFRESH
	COM_SHUTDOWN
	COM_STATISTICS
	COM_PROCESS_INFO
	COM_CONNECT
	COM_PROCESS_KILL
	COM_DEBUG
	COM_PING
	COM_TIME
	COM_DELAYED_INSERT
	COM_CHANGE_USER
	COM_BINLOG_DUMP
	COM_TABLE_DUMP
	COM_CONNECT_OUT
	COM_REGISTER_SLAVE
	COM_STMT_PREPARE
	COM_STMT_EXECUTE
	COM_STMT_SEND_LONG_DATA
	COM_STMT_CLOSE
	COM_STMT_RESET
	COM_SET_OPTION
	COM_STMT_FETCH
)

const (
	MYSQL_TYPE_DECIMAL     = 0
	MYSQL_TYPE_TINY        = 1
	MYSQL_TYPE_SHORT       = 2
	MYSQL_TYPE_LONG        = 3
	MYSQL_TYPE_FLOAT       = 4
	MYSQL_TYPE_DOUBLE      = 5
	MYSQL_TYPE_NULL        = 6
	MYSQL_TYPE_TIMESTAMP   = 7
	MYSQL_TYPE_LONGLONG    = 8
	MYSQL_TYPE_INT24       = 9
	MYSQL_TYPE_DATE        = 10
	MYSQL_TYPE_TIME        = 11
	MYSQL_TYPE_DATETIME    = 12
	MYSQL_TYPE_YEAR        = 13
	MYSQL_TYPE_NEWDATE     = 14
	MYSQL_TYPE_VARCHAR     = 15
	MYSQL_TYPE_BIT         = 16
	MYSQL_TYPE_NEWDECIMAL  = 246
	MYSQL_TYPE_ENUM        = 247
	MYSQL_TYPE_SET         = 248
	MYSQL_TYPE_TINY_BLOB   = 249
	MYSQL_TYPE_MEDIUM_BLOB = 250
	MYSQL_TYPE_LONG_BLOB   = 251
	MYSQL_TYPE_BLOB        = 252
	MYSQL_TYPE_VAR_STRING  = 253
	MYSQL_TYPE_STRING      = 254
	MYSQL_TYPE_GEOMETRY    = 255
)

const (
	NOT_NULL_FLAG           = 1         /* Field can't be NULL */
	PRI_KEY_FLAG            = 2         /* Field is part of a primary key */
	UNIQUE_KEY_FLAG         = 4         /* Field is part of a unique key */
	MULTIPLE_KEY_FLAG       = 8         /* Field is part of a key */
	BLOB_FLAG               = 16        /* Field is a blob */
	UNSIGNED_FLAG           = 32        /* Field is unsigned */
	ZEROFILL_FLAG           = 64        /* Field is zerofill */
	BINARY_FLAG             = 128       /* Field is binary   */
	ENUM_FLAG               = 256       /* field is an enum */
	AUTO_INCREMENT_FLAG     = 512       /* field is a autoincrement field */
	TIMESTAMP_FLAG          = 1024      /* Field is a timestamp */
	SET_FLAG                = 2048      /* field is a set */
	NO_DEFAULT_VALUE_FLAG   = 4096      /* Field doesn't have default value */
	ON_UPDATE_NOW_FLAG      = 8192      /* Field is set to NOW on UPDATE */
	NUM_FLAG                = 32768     /* Field is num (for clients) */
	PART_KEY_FLAG           = 16384     /* Intern; Part of some key */
	GROUP_FLAG              = 32768     /* Intern: Group field */
	UNIQUE_FLAG             = 65536     /* Intern: Used by sql_yacc */
	BINCMP_FLAG             = 131072    /* Intern: Used by sql_yacc */
	GET_FIXED_FIELDS_FLAG   = (1 << 18) /* Used to get fields in item tree */
	FIELD_IN_PART_FUNC_FLAG = (1 << 19) /* Field part of partition func */
	FIELD_IN_ADD_INDEX      = (1 << 20) /* Intern: Field used in ADD INDEX */
	FIELD_IS_RENAMED        = (1 << 21) /* Intern: Field is being renamed */
)

const (
	CURSOR_TYPE_NO_CURSOR  = 0
	CURSOR_TYPE_READ_ONLY  = 1
	CURSOR_TYPE_FOR_UPDATE = 2
	CURSOR_TYPE_SCROLLABLE = 4
)

const MAX_PACKET_SIZE = 1<<24 - 1
const MAX_DATA_CHUNK = 1 << 19
const CHARSET_UTF8_GENERAL_CI = 33

const (
	OK           = 0x00
	EOF          = 0xfe
	LOCAL_INFILE = 0xfb
	ERR          = 0xff
)
