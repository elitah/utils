package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/elitah/utils/logs"

	"github.com/mattn/go-sqlite3"
)

const (
	FlagStatus = iota // 当前计划状态
	PlanAutoBackup

	FlagMax
)

type options struct {
	flag_groups [FlagMax]uint32

	backup_path  string
	backup_max   int
	backup_step  int
	backup_delay int

	dbchan_master string
	dbchan_backup string
}

type Option func(c *options)

func WithBackup(path string, args ...int) Option {
	if "" != path {
		return func(opts *options) {
			opts.backup_path = path

			if 1 <= len(args) {
				opts.backup_max = args[0]
			}

			if 2 <= len(args) {
				if opts.backup_step < args[1] {
					opts.backup_step = args[1]
				}
			}

			if 3 <= len(args) {
				if opts.backup_delay < args[2] {
					opts.backup_delay = args[2]
				}
			}
		}
	}

	return nil
}

type SQLiteDB struct {
	sync.Mutex

	store sqliteRawConnStore

	conn *sql.DB

	opts options
}

func NewSQLiteDB(opts ...Option) *SQLiteDB {
	r := &SQLiteDB{
		opts: options{
			backup_step:  1024, // 单步备份长度
			backup_delay: 10,   // 单步备份被打断后延迟时间（毫秒）
		},
	}

	for _, opt := range opts {
		if nil != opt {
			opt(&r.opts)
		}
	}

	r.opts.dbchan_master = fmt.Sprintf("sqlite3_master_%p", r)

	// 注册sqlite主驱动
	sql.Register(r.opts.dbchan_master, &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			r.store.Set(conn)
			return nil
		},
	})

	// 注册sqlite备份驱动
	if "" != r.opts.backup_path {
		//
		r.opts.dbchan_backup = fmt.Sprintf("sqlite3_backup_%p", r)
		//
		sql.Register(r.opts.dbchan_backup, &sqlite3.SQLiteDriver{
			ConnectHook: func(slave *sqlite3.SQLiteConn) error {
				if master := r.store.Get(); nil != master {
					return SQLiteBackup(master, slave, r.opts.backup_step, r.opts.backup_delay)
				} else {
					return fmt.Errorf("no master connection be found")
				}
			},
		})
	}

	return r
}

func (this *SQLiteDB) GetConn(flags ...bool) (conn *sql.DB, err error) {
	if 0x0 == atomic.LoadUint32(&this.opts.flag_groups[FlagStatus]) {
		this.Lock()

		if nil != this.conn {
			conn = this.conn
		} else {
			if conn, err = sql.Open(this.opts.dbchan_master, "file::memory:?mode=memory&cache=shared"); nil == err {
				//
				conn.SetMaxOpenConns(1)
				//
				this.conn = conn
			}
		}

		this.Unlock()
	} else {
		err = fmt.Errorf("database was closed")
	}

	if nil != err || (0 < len(flags) && flags[0]) {
		return
	}

	err = conn.Ping()

	return
}

func (this *SQLiteDB) StartBackup(auto bool) (int64, error) {
	if "" != this.opts.backup_path {
		// 检查文件是否存在
		if info, err := os.Stat(this.opts.backup_path); nil == err {
			if 0 == info.Size() {
				return 0, nil
			}
		} else {
			return 0, err
		}
		// 开始同步
		if master, err := this.GetConn(true); nil == err {
			if slave, err := sql.Open("sqlite3", this.opts.backup_path); nil == err {
				if n, err := SQLiteSync(master, slave, filepath.Dir(this.opts.backup_path)); nil == err {
					// 如果开启了自动备份
					if auto {
						// 检查自动备份是否启用
						if atomic.CompareAndSwapUint32(&this.opts.flag_groups[PlanAutoBackup], 0x0, 0x1) {
							go func() {
								// 缓冲
								var sb strings.Builder
								// 计时
								var unix_trigger int64
								// 监测通道
								ch := make(chan string, 1)
								// 表变化监测模块
								tableList := make(map[string]bool)
								// 注册监听
								this.store.RegisterNotice(ch)
								// 3秒循环
								ticker := time.NewTicker(3 * time.Second)
								// 循环监测数据库变动
								for {
									select {
									case name, ok := <-ch:
										if ok {
											// 标记
											tableList[name] = true
											// 倒计时15秒
											if 0 == unix_trigger {
												t := time.Now().Add(15 * time.Second)
												//
												unix_trigger = t.Unix()
												//
												logs.Info("数据库出现变化，15秒后(%s)开始备份", t.Format("2006-01-02 15:04:05"))
											}
										} else {
											return
										}
									case <-ticker.C:
										// 如果倒计时已结束
										if 0 != unix_trigger && unix_trigger < time.Now().Unix() {
											// 复位计时
											unix_trigger = 0
											// 清空缓冲
											sb.Reset()
											// 检查标记
											for key, value := range tableList {
												if value {
													// 标点符号
													if 0 < sb.Len() {
														sb.WriteRune(',')
													}
													// 表名
													sb.WriteString(key)
													// 清除标记
													tableList[key] = false
												}
											}
											// 信息输出
											logs.Info("开始备份数据库，发生改变的表为: %s", sb.String())
											// 开始备份
											if err := this.Backup(); nil != err {
												logs.Error(err)
											}
										}
									}
								}
							}()
						}
					}
					return n, nil
				} else {
					return 0, fmt.Errorf("unable sync database, %w", err)
				}
			} else {
				return 0, fmt.Errorf("unable open database, %w", err)
			}
		} else {
			return 0, fmt.Errorf("unable get database connection, %w", err)
		}
	}
	return 0, nil
}

func (this *SQLiteDB) CreateTable(name, structure string, flags ...bool) bool {
	//
	if conn, err := this.GetConn(); nil == err {
		//
		if 0 < len(flags) && flags[0] {
			name = fmt.Sprintf("temp.%s", name)
		}
		//
		sql := fmt.Sprintf("CREATE TABLE %s (%s);", name, strings.Join(strings.Fields(structure), " "))
		//
		if _, err := conn.Exec(sql); nil == err {
			return true
		} else {
			logs.Error(err)
		}
	}

	return false
}

func (this *SQLiteDB) Backup() error {
	if "" != this.opts.backup_path {
		if 0 < this.opts.backup_max {
			if ext := filepath.Ext(this.opts.backup_path); "" != ext {
				var path1, path2 string
				//
				basepath := this.opts.backup_path[:len(this.opts.backup_path)-len(ext)]
				//
				for i := this.opts.backup_max - 1; 0 <= i; i-- {
					//
					if 0 == i {
						path1 = this.opts.backup_path
						path2 = fmt.Sprintf("%s.000%s", basepath, ext)
					} else {
						path1 = fmt.Sprintf("%s.%03d%s", basepath, i-1, ext)
						path2 = fmt.Sprintf("%s.%03d%s", basepath, i, ext)
					}
					//
					os.Rename(path1, path2)
				}
			}
		}

		start := time.Now()

		logs.Info("数据库备份开始...")

		defer func() {
			logs.Info("数据库备份完成, 耗时: %v...", time.Since(start))
		}()

		this.Lock()
		defer this.Unlock()

		// 打开备份数据库
		if db, err := sql.Open(this.opts.dbchan_backup, this.opts.backup_path); nil == err {
			// 关闭数据库
			defer db.Close()
			// 通过Ping方法激活数据库
			if err := db.Ping(); nil == err {
				return nil
			} else {
				return err
			}
		} else {
			return err
		}
	}

	return fmt.Errorf("no backup file path be provided")
}

func (this *SQLiteDB) Close() {
	if atomic.CompareAndSwapUint32(&this.opts.flag_groups[FlagStatus], 0x0, 0x1) {
		this.store.Close()

		this.Backup()

		this.Lock()

		this.conn.Close()

		this.Unlock()
	}
}
