package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/elitah/utils/logs"
)

type sqliteTableInfo struct {
	name string
	sql  string
	sync bool
	cnt  int64
}

func SQLiteSync(master, slave *sql.DB, dir string) (int64, error) {
	if err := master.Ping(); nil == err {
		if err := slave.Ping(); nil == err {
			var list []*sqliteTableInfo

			var tbl_name, sql string

			if rows, err := master.Query("SELECT tbl_name, sql FROM sqlite_master WHERE type=='table';"); nil == err {
				// 遍历结果
				for rows.Next() {
					if err := rows.Scan(&tbl_name, &sql); nil == err {
						if "sqlite_sequence" != tbl_name && "" != sql {
							list = append(list, &sqliteTableInfo{
								name: tbl_name,
								sql:  sql,
							})
						}
					} else {
						return 0, err
					}
				}
				//
				rows.Close()
				//
				for _, item := range list {
					if row := slave.QueryRow("SELECT sql FROM sqlite_master WHERE type=='table' AND tbl_name==?;", item.name); nil != row {
						if err := row.Scan(&sql); nil == err {
							//
							item.sync = sql == item.sql
							//
							if row := slave.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s;", item.name)); nil != row {
								row.Scan(&item.cnt)
							}
						} else {
							//logs.Info(err)
						}
					}
				}
				//
				cnt := int64(0)
				//
				for _, item := range list {
					if 0 < item.cnt {
						if item.sync {
							//...
							logs.Info("正在同步: ", item.name)
							//
							if rows, err := slave.Query(fmt.Sprintf("SELECT * FROM %s;", item.name)); nil == err {
								if types, err := rows.ColumnTypes(); nil == err {
									if tx, err := master.Begin(); nil == err {
										sql := "INSERT INTO " + item.name + " ("
										for i, _ := range types {
											if 0 != i {
												sql += ", "
											}
											sql += types[i].Name()
										}
										sql += ") VALUES ("
										for i, _ := range types {
											if 0 != i {
												sql += ", "
											}
											sql += "?"
										}
										sql += ");"
										if stmt, err := tx.Prepare(sql); nil == err {
											value := make([]interface{}, len(types))
											for i, _ := range value {
												switch strings.ToLower(types[i].DatabaseTypeName()) {
												case "integer":
													value[i] = new(int64)
												case "float":
													value[i] = new(float64)
												case "blob":
													value[i] = new([]byte)
												case "text":
													value[i] = new(string)
												case "timestamp", "datetime", "date":
													value[i] = new(time.Time)
												case "boolean":
													value[i] = new(bool)
												default:
													logs.Error("不支持的类型:", types[i].DatabaseTypeName())
												}
											}
											//
											for rows.Next() {
												if err := rows.Scan(value...); nil == err {
													if _, err := stmt.Exec(value...); nil == err {
														cnt++
													} else {
														logs.Error(err)
													}
												} else {
													logs.Error(err)
												}
											}
											//
											if err := tx.Commit(); nil != err {
												logs.Error(err)
											}
											//
											stmt.Close()
										} else {
											logs.Error(err)
										}
									} else {
										logs.Error(err)
									}
								} else {
									logs.Error(err)
								}
								rows.Close()
							} else {
								logs.Error(err)
							}
						} else if "" != dir {
							//...
							logs.Warn("正在备份: ", item.name)
							//
							list := make([]interface{}, 0, int(item.cnt))
							//
							if rows, err := slave.Query(fmt.Sprintf("SELECT * FROM %s;", item.name)); nil == err {
								if types, err := rows.ColumnTypes(); nil == err {
									for rows.Next() {
										value := make([]interface{}, len(types))
										for i, _ := range value {
											switch strings.ToLower(types[i].DatabaseTypeName()) {
											case "integer":
												value[i] = new(int64)
											case "float":
												value[i] = new(float64)
											case "blob":
												value[i] = new([]byte)
											case "text":
												value[i] = new(string)
											case "timestamp", "datetime", "date":
												value[i] = new(time.Time)
											case "boolean":
												value[i] = new(bool)
											default:
												logs.Info("不支持的类型:", types[i].DatabaseTypeName())
											}
										}
										if err := rows.Scan(value...); nil == err {
											list = append(list, value)
										}
									}
								} else {
									logs.Error(err)
								}
								//
								rows.Close()
							} else {
								logs.Error(err)
							}
							if data, err := json.Marshal(&struct {
								SQL   string      `json:"sql"`
								Count int64       `json:"count"`
								List  interface{} `json:"list"`
							}{
								SQL:   item.sql,
								Count: item.cnt,
								List:  list,
							}); nil == err {
								ioutil.WriteFile(
									fmt.Sprintf(
										"%s/sqlite_backup_%s_%d.json",
										dir,
										item.name,
										time.Now().Unix(),
									),
									data,
									0644,
								)
							} else {
								logs.Error(err)
							}
						}
					}
				}
				return cnt, nil
			} else {
				return 0, err
			}
		} else {
			return 0, err
		}
	} else {
		return 0, err
	}
}
