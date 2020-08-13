package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

const (
	sqlite3Path = "/tmp/tmp_test.db"
)

func fileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func main() {
	if fileExist(sqlite3Path) {
		os.Remove(sqlite3Path)
	}
	db, err := sql.Open("sqlite3", sqlite3Path)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	sqlStmt := "CREATE TABLE wal(idx INTEGER, data BLOB);"
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := db.Prepare("INSERT INTO wal VALUES (?, ?)")
	for i := 0; i < 100; i++ {
		_,err = stmt.Exec(i,[]byte(fmt.Sprintf("test value %d",i)))
	}

}
