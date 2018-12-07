package DB

import (
	"fmt"
	"database/sql"
	_ "encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/BurntSushi/toml"
	"time"
)

type Conn struct {
	S struct {
		IP     string `toml:"ip"`
		Host   string `toml:"host"`
		Port   string `toml:"port"`
		DbName string `toml:"dbName"`
	} `toml:"S"`
	M struct {
		IP     string `toml:"ip"`
		Host   string `toml:"host"`
		Port   string `toml:"port"`
		DbName string `toml:"dbName"`
	} `toml:"M"`
}

type Config struct {
	Db struct {
		User            string `toml:"user"`
		Pass            string `toml:"pass"`
		ConnMaxLifetime int    `toml:"connMaxLifetime"`
		CSMEM           Conn   `toml:"CS_MEM"`
		SPORTMEM        Conn   `toml:"SPORT_MEM"`
		SPORTRECORD     Conn   `toml:"SPORT_RECORD"`
		Univer          Conn   `toml:"Univer"`
		IPLMAIN         Conn   `toml:"IPL_MAIN"`
		SLOT_TOOL       Conn   `toml:"SLOT_TOOL"`
		SLOT_GAME       Conn   `toml:"SLOT_GAME"`
		Almond		Conn   `toml:"Almond"`
	} `toml:"db"`
}

var CS_MEM_S *sql.DB
var CS_MEM_M *sql.DB
var Univer_S *sql.DB
var SPORT_RECORD_M *sql.DB
var IPL_MAIN_M *sql.DB
var SLOT_TOOL_M *sql.DB
var SLOT_GAME_M *sql.DB
var Almond_M *sql.DB
var Almond_S *sql.DB


func InitDB(path string) {
	var conf Config

	if _, err := toml.DecodeFile(path, &conf); err != nil {
		// handle error
		fmt.Println(err)
	}

	CS_MEM_S = setConnect(conf, conf.Db.CSMEM, false)
	CS_MEM_M = setConnect(conf, conf.Db.CSMEM, true)
	Univer_S = setConnect(conf, conf.Db.Univer, false)
	SPORT_RECORD_M = setConnect(conf, conf.Db.SPORTRECORD, true)
	IPL_MAIN_M = setConnect(conf, conf.Db.IPLMAIN, true)
	SLOT_TOOL_M = setConnect(conf, conf.Db.SLOT_TOOL, true)
	SLOT_GAME_M = setConnect(conf, conf.Db.SLOT_GAME, true)
	Almond_M = setConnect(conf, conf.Db.Almond, true)
	Almond_S = setConnect(conf, conf.Db.Almond, false)

	fmt.Printf("Load DB config success.\n DB: %v\n\n", conf)
}

// Connection pool
func setConnect(conf Config, conn Conn, master bool) *sql.DB {
	var con, dbName string

	if master {
		con = conf.Db.User + ":" + conf.Db.Pass + "@tcp(" + conn.M.IP + ":" + conn.M.Port
		con += ")/" + conn.M.DbName
		dbName = conn.M.DbName
	} else {
		con = conf.Db.User + ":" + conf.Db.Pass + "@tcp(" + conn.S.IP + ":" + conn.S.Port
		con += ")/" + conn.S.DbName
		dbName = conn.S.DbName
	}

	db, err := sql.Open("mysql", con); if err != nil {
		fmt.Println("DB "+dbName+" Connect Error", err)
	}

	// Connection limit
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)

	// Timeout for long connection
	setConnMaxLifetime(conf, db)

	return db
}

// Set timeout for long connection.
func setConnMaxLifetime(conf Config, db *sql.DB)  {
	var connMaxLifetime int

	if conf.Db.ConnMaxLifetime == 0 {
		connMaxLifetime = 180 // Default timeout
	} else {
		connMaxLifetime = conf.Db.ConnMaxLifetime
	}

	db.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)
}
