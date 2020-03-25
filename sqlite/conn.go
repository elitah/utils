package sqlite

import (
	"container/list"
	"sync"

	"github.com/mattn/go-sqlite3"
)

type sqliteRawConnStore struct {
	sync.RWMutex

	list list.List

	conn *sqlite3.SQLiteConn
}

func (this *sqliteRawConnStore) Set(conn *sqlite3.SQLiteConn) {
	this.Lock()
	defer this.Unlock()

	if nil != this.conn {
		this.conn.Close()
	}

	conn.RegisterUpdateHook(this.HandleUpdate)

	this.conn = conn
}

func (this *sqliteRawConnStore) Get() *sqlite3.SQLiteConn {
	this.RLock()
	defer this.RUnlock()

	return this.conn
}

func (this *sqliteRawConnStore) RegisterNotice(ch chan string) {
	this.list.PushBack(ch)
}

func (this *sqliteRawConnStore) HandleUpdate(op int, db string, table string, rowid int64) {
	if "temp" != db {
		for e := this.list.Front(); nil != e; e = e.Next() {
			if ch, ok := e.Value.(chan string); ok {
				if cap(ch) > len(ch) {
					ch <- table
				}
			}
		}
	}
}

func (this *sqliteRawConnStore) Close() {
	for e := this.list.Front(); nil != e; e = e.Next() {
		if result := this.list.Remove(e); nil != result {
			if ch, ok := result.(chan string); ok {
				close(ch)
			}
		}
	}
}
