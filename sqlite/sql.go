package sqlite

import (
	"fmt"
)

func SQLiteCount(db *SQLiteDB, tbl_name string) (int64, error) {
	if conn, err := db.GetConn(true); nil == err {
		if row := conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tbl_name)); nil != row {
			var cnt int64
			if err := row.Scan(&cnt); nil == err {
				return cnt, nil
			} else {
				return 0, err
			}
		} else {
			return 0, fmt.Errorf("QueryRow return failed")
		}
	} else {
		return 0, err
	}
}
