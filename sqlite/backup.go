package sqlite

import (
	"time"

	"github.com/mattn/go-sqlite3"
)

func SQLiteBackup(master, slave *sqlite3.SQLiteConn, step, delay int) error {
	if bk, err := slave.Backup("main", master, "main"); nil == err {
		//
		defer bk.Finish()
		//
		for {
			if ok, err := bk.Step(step); nil == err {
				if ok {
					return nil
				} else {
					time.Sleep(time.Duration(delay) * time.Millisecond)
				}
			} else {
				return err
			}
		}
	} else {
		return err
	}
}
