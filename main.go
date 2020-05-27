package main

import (
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/elitah/utils/aes"
	"github.com/elitah/utils/atomic"
	"github.com/elitah/utils/bufferpool"
	"github.com/elitah/utils/cpu"
	"github.com/elitah/utils/exepath"
	"github.com/elitah/utils/hash"
	"github.com/elitah/utils/hex"
	"github.com/elitah/utils/httptools"
	"github.com/elitah/utils/logs"
	"github.com/elitah/utils/mutex"
	"github.com/elitah/utils/number"
	"github.com/elitah/utils/platform"
	"github.com/elitah/utils/random"
	"github.com/elitah/utils/sqlite"
	"github.com/elitah/utils/vhost"
	"github.com/elitah/utils/wait"
)

func main() {
	logs.SetLogger(logs.AdapterConsole, `{"level":99,"color":true}`)
	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(3)
	logs.Async()

	defer logs.Close()

	logs.Info("hello utils")

	testAES()

	testAtomic()
	testNumber()

	//testBufferPool()
	//testVhost()
	//testHttpTools()

	testWait()

	testExtPath()
	testHex()
	testRandom()
	testPlatform()
	testCPU()
	testMutex()
	testHash()

	testSQLite()
}

func testAES() {
	logs.Info("--- hello utils/aes test ----------------------------------------------------------------")

	if t0 := aes.NewAESTool("123456"); nil != t0 {
		if t1 := aes.NewAESTool("123456"); nil != t1 {
			t0.EncryptInit()

			t0.Write([]byte("exampleplaintext"))

			t0.Encrypt(nil)

			t1.Write(t0.Bytes())

			t1.Decrypt(nil)

			logs.Info(t1.String())
		}
	}
}

func testAtomic() {
	logs.Info("--- hello utils/atomic test ----------------------------------------------------------------")

	xs32 := atomic.AInt32(0)

	logs.Info(xs32.Add(1))
	logs.Info(xs32.CAS(1, 0))
	logs.Info(xs32.Load())
	logs.Info(xs32.Swap(1))
	logs.Info(xs32.Load())
	logs.Info(xs32.Sub(1))
	logs.Info(xs32.Load())

	xs64 := atomic.AInt64(0)

	logs.Info(xs64.Add(1))
	logs.Info(xs64.CAS(1, 0))
	logs.Info(xs64.Load())
	logs.Info(xs64.Swap(1))
	logs.Info(xs64.Load())
	logs.Info(xs64.Sub(1))
	logs.Info(xs64.Load())

	xu32 := atomic.AUint32(0)

	logs.Info(xu32.Add(1))
	logs.Info(xu32.CAS(1, 0))
	logs.Info(xu32.Load())
	logs.Info(xu32.Swap(1))
	logs.Info(xu32.Load())
	logs.Info(xu32.Sub(1))
	logs.Info(xu32.Load())

	xu64 := atomic.AUint64(0)

	logs.Info(xu64.Add(1))
	logs.Info(xu64.CAS(1, 0))
	logs.Info(xu64.Load())
	logs.Info(xu64.Swap(1))
	logs.Info(xu64.Load())
	logs.Info(xu64.Sub(1))
	logs.Info(xu64.Load())

	xptr := atomic.AUintptr(0)

	logs.Info(xptr.Add(1))
	logs.Info(xptr.CAS(1, 0))
	logs.Info(xptr.Load())
	logs.Info(xptr.Swap(1))
	logs.Info(xptr.Load())
	logs.Info(xptr.Sub(1))
	logs.Info(xptr.Load())
}

func testNumber() {
	logs.Info("--- hello utils/number test ----------------------------------------------------------------")

	logs.Info("IsNumber: %v", number.IsNumeric(nil))
	logs.Info("IsNumber: %v", number.IsNumeric(5))
	logs.Info("IsNumber: %v", number.IsNumeric(0x5))
	logs.Info("IsNumber: %v", number.IsNumeric(0.1))

	if v, err := number.ToInt64(-50); nil == err {
		logs.Info("ToInt64: %v", v)
	} else {
		logs.Info("ToInt64: error: %v", err)
	}

	if v, err := number.ToInt64(0x50); nil == err {
		logs.Info("ToInt64: %v", v)
	} else {
		logs.Info("ToInt64: error: %v", err)
	}

	if v, err := number.ToInt64("0x50"); nil == err {
		logs.Info("ToInt64: %v", v)
	} else {
		logs.Info("ToInt64: error: %v", err)
	}
}

func testBufferPool() {
	logs.Info("--- hello utils/bufferpool test ----------------------------------------------------------------")

	logs.Info("--- bufferpool.Get(): start -----------------------------------------------------------------")

	b := bufferpool.Get()

	logs.Info("--- bufferpool.Get(): done ---------------------------------------------------------------")

	logs.Info("--- bufferpool: test function ReadFromLimited ----------------------------------------------------")

	r := strings.NewReader("some io.Reader stream to be read\n")

	b.ReadFromLimited(r, 10)

	logs.Info("bufferpool.ReadFromLimited(): %s", b.String())

	b.Reset()

	b.ReadFromLimited(r, 10)

	logs.Info("bufferpool.ReadFromLimited(): %s", b.String())

	logs.Info("--- bufferpool: test TeeReader -------------------------------------------------")

	if _b := bufferpool.Get(); nil != _b {
		r = strings.NewReader("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

		b.Reset()

		if _r, err := _b.TeeReader(r, 10); nil == err {
			b.ReadFromLimited(_r, 20)
		}

		logs.Info("bufferpool.ReadFromLimited(): b: %s", b.String())
		logs.Info("bufferpool.ReadFromLimited(): _b: %s", _b.String())
	}

	logs.Info("--- bufferpool: test buffer reference count -------------------------------------")

	b.AddRefer(6)

	for i := 0; !b.IsFree(); i++ {
		logs.Info(i, b.Free())
	}

	time.Sleep(time.Second)

	os.Exit(-1)
}

func testVhost() {
	logs.Info("--- hello utils/vhost test ----------------------------------------------------------------")

	if listener, err := net.Listen("tcp", ":51180"); nil == err {
		for {
			if conn, err := listener.Accept(); nil == err {
				if httpConn, err := vhost.HTTP(conn); nil == err {
					logs.Info(httpConn.Host)
					httpConn.Close()
				} else {
					logs.Error(err)
				}
				conn.Close()
			} else {
				logs.Error(err)

				os.Exit(-1)
			}
		}
	} else {
		logs.Error(err)
	}
}

func testHttpTools() {
	logs.Info("--- hello utils/httptools test ----------------------------------------------------------------")

	logs.Info(http.ListenAndServe(":38082", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 获取通用处理器(调试模式)
		// if resp := httptools.NewHttpHandler(r, true); nil != resp {
		// 获取通用处理器
		if resp := httptools.NewHttpHandler(r, true); nil != resp {
			// 调试模式
			//resp.Debug(true)
			// 释放
			defer func() {
				if o := resp.Output(w); "" != o {
					logs.Info(o)
				}
				resp.Release()
			}()
			// 识别路径
			switch resp.GetPath() {
			case "/":
				if resp.HttpOnlyIs("GET") {
					resp.SendHttpRedirect("/test")
				}
				return
			case "/post":
				if resp.HttpOnlyIs("GET", "POST") {
					switch resp.Method {
					case "GET":
						resp.SendHTML(`<form action="/post" method="post" enctype="multipart/form-data">`)
						resp.SendHTML(`<p><input type="file" name="file"></p>`)
						resp.SendHTML(`<p><input type="text" name="name"></p>`)
						resp.SendHTML(`<p><input type="submit" value="submit"></p>`)
						resp.SendHTML(`</form>`)
					case "POST":
						if err := resp.GetUpload(func(part *multipart.Part) bool {
							logs.Info(part)
							return true
						}); nil == err {
							resp.SendHTML(`<h3>ok</h3>`)
						} else {
							logs.Error(err)
						}
					}
				}
				return
			case "/test":
				if resp.HttpOnlyIs("GET") {
					if err := resp.TemplateWrite([]byte(`<html>
	<head>
		<title>test</title>
	</head>
	<body>
		<p>hello test, <a href="{{ .Path }}">bye</a></p>
	</body>
</html>
`), struct {
						Path string
					}{
						Path: "/bye",
					}, "text/html"); nil != err {
						logs.Error(err)
					}
				}
				return
			case "/bye":
				if resp.HttpOnlyIs("GET") {
					resp.SendJSAlert("提示", "成功", "/")
					//
					go func() {
						time.Sleep(3 * time.Second)

						os.Exit(0)
					}()
				}
				return
			}
			//
			resp.NotFound()
			//
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	})))
}

func testWait() {
	logs.Info("--- hello utils/wait test ----------------------------------------------------------------")

	logs.Info("wait.Signal(): start")

	wait.Signal(
		wait.WithNotify(func(s os.Signal) bool {
			logs.Info(s)
			return true
		}),
		wait.WithSignal(syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM),
		wait.WithTicket(1, func(t time.Time) {
			logs.Info(t)
		}),
	)

	logs.Info("wait.Signal(): done")
}

func testExtPath() {
	logs.Info("--- hello utils/exepath test ----------------------------------------------------------------")

	logs.Info("exepath.GetExePath():\n\t%s\n", exepath.GetExePath())
	logs.Info("exepath.GetExeDir():\n\t%s\n", exepath.GetExeDir())
}

func testHex() {
	logs.Info("--- hello utils/hex test ----------------------------------------------------------------")

	data := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66}
	result := hex.EncodeToStringWithSeq(data, ' ')

	logs.Info("hex.EncodeNumberToStringWithSeq(1, le):\n\t%s\n", hex.EncodeNumberToStringWithSeq("1", ' ', true))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~2, le):\n\t%s\n", hex.EncodeNumberToStringWithSeq("12", ' ', true))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~3, le):\n\t%s\n", hex.EncodeNumberToStringWithSeq("123", ' ', true))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~4, le):\n\t%s\n", hex.EncodeNumberToStringWithSeq("1234", ' ', true))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~5, le):\n\t%s\n", hex.EncodeNumberToStringWithSeq("12345", ' ', true))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~6, le):\n\t%s\n", hex.EncodeNumberToStringWithSeq("123456", ' ', true))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~7, le):\n\t%s\n", hex.EncodeNumberToStringWithSeq("1234567", ' ', true))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~8, le):\n\t%s\n", hex.EncodeNumberToStringWithSeq("12345678", ' ', true))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~9, le):\n\t%s\n", hex.EncodeNumberToStringWithSeq("123456789", ' ', true))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~0, le):\n\t%s\n", hex.EncodeNumberToStringWithSeq("1234567890", ' ', true))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~01, le):\n\t%s\n", hex.EncodeNumberToStringWithSeq("12345678901", ' ', true))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~02, le):\n\t%s\n", hex.EncodeNumberToStringWithSeq("123456789012", ' ', true))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~03, le):\n\t%s\n", hex.EncodeNumberToStringWithSeq("1234567890123", ' ', true))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~04, le):\n\t%s\n", hex.EncodeNumberToStringWithSeq("12345678901234", ' ', true))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~05, le):\n\t%s\n", hex.EncodeNumberToStringWithSeq("123456789012345", ' ', true))

	logs.Info("hex.EncodeNumberToStringWithSeq(1, be):\n\t%s\n", hex.EncodeNumberToStringWithSeq("1", ' ', false))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~2, be):\n\t%s\n", hex.EncodeNumberToStringWithSeq("12", ' ', false))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~3, be):\n\t%s\n", hex.EncodeNumberToStringWithSeq("123", ' ', false))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~4, be):\n\t%s\n", hex.EncodeNumberToStringWithSeq("1234", ' ', false))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~5, be):\n\t%s\n", hex.EncodeNumberToStringWithSeq("12345", ' ', false))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~6, be):\n\t%s\n", hex.EncodeNumberToStringWithSeq("123456", ' ', false))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~7, be):\n\t%s\n", hex.EncodeNumberToStringWithSeq("1234567", ' ', false))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~8, be):\n\t%s\n", hex.EncodeNumberToStringWithSeq("12345678", ' ', false))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~9, be):\n\t%s\n", hex.EncodeNumberToStringWithSeq("123456789", ' ', false))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~0, be):\n\t%s\n", hex.EncodeNumberToStringWithSeq("1234567890", ' ', false))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~01, be):\n\t%s\n", hex.EncodeNumberToStringWithSeq("12345678901", ' ', false))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~02, be):\n\t%s\n", hex.EncodeNumberToStringWithSeq("123456789012", ' ', false))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~03, be):\n\t%s\n", hex.EncodeNumberToStringWithSeq("1234567890123", ' ', false))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~04, be):\n\t%s\n", hex.EncodeNumberToStringWithSeq("12345678901234", ' ', false))
	logs.Info("hex.EncodeNumberToStringWithSeq(1~05, be):\n\t%s\n", hex.EncodeNumberToStringWithSeq("123456789012345", ' ', false))

	logs.Info("hex.EncodeNumberToStringWithSeq():\n\t%s\n", hex.EncodeNumberToStringWithSeq(-123456789012345, ' ', true, 1))
	logs.Info("hex.EncodeNumberToStringWithSeq():\n\t%s\n", hex.EncodeNumberToStringWithSeq(-123456789012345, ' ', true, 3))
	logs.Info("hex.EncodeNumberToStringWithSeq():\n\t%s\n", hex.EncodeNumberToStringWithSeq(-123456789012345, ' ', true, 5))
	logs.Info("hex.EncodeNumberToStringWithSeq():\n\t%s\n", hex.EncodeNumberToStringWithSeq(-123456789012345, ' ', true, 7))
	logs.Info("hex.EncodeNumberToStringWithSeq():\n\t%s\n", hex.EncodeNumberToStringWithSeq(-123456789012345, ' ', false, 1))
	logs.Info("hex.EncodeNumberToStringWithSeq():\n\t%s\n", hex.EncodeNumberToStringWithSeq(-123456789012345, ' ', false, 3))
	logs.Info("hex.EncodeNumberToStringWithSeq():\n\t%s\n", hex.EncodeNumberToStringWithSeq(-123456789012345, ' ', false, 5))
	logs.Info("hex.EncodeNumberToStringWithSeq():\n\t%s\n", hex.EncodeNumberToStringWithSeq(-123456789012345, ' ', false, 7))

	logs.Info("hex.EncodeToString():\n\t%s\n", hex.EncodeToString(data))

	logs.Info("hex.EncodeToStringWithSeq():\n\t%s\n", result)

	if data, err := hex.DecodeStringWithSeq(result); nil == err {
		logs.Info("hex.DecodeStringWithSeq():\n\t%x\n", data)
	} else {
		logs.Info("hex.DecodeStringWithSeq():\n\t%v\n", err)
	}
}

func testRandom() {
	logs.Info("--- hello utils/random test ----------------------------------------------------------------")

	logs.Info("random.ModeALL(64):\n\t%s\n", random.NewRandomString(random.ModeALL, 64))
	logs.Info("random.ModeNoLower(64):\n\t%s\n", random.NewRandomString(random.ModeNoLower, 64))
	logs.Info("random.ModeNoUpper(64):\n\t%s\n", random.NewRandomString(random.ModeNoUpper, 64))
	logs.Info("random.ModeNoNumber(64):\n\t%s\n", random.NewRandomString(random.ModeNoNumber, 64))
	logs.Info("random.ModeNoLowerNumber(64):\n\t%s\n", random.NewRandomString(random.ModeNoLowerNumber, 64))
	logs.Info("random.ModeNoUpperNumber(64):\n\t%s\n", random.NewRandomString(random.ModeNoUpperNumber, 64))
	logs.Info("random.ModeNoLine(64):\n\t%s\n", random.NewRandomString(random.ModeNoLine, 64))
	logs.Info("random.ModeNoLowerLine(64):\n\t%s\n", random.NewRandomString(random.ModeNoLowerLine, 64))
	logs.Info("random.ModeNoUpperLine(64):\n\t%s\n", random.NewRandomString(random.ModeNoUpperLine, 64))
	logs.Info("random.ModeOnlyLower(64):\n\t%s\n", random.NewRandomString(random.ModeOnlyLower, 64))
	logs.Info("random.ModeOnlyUpper(64):\n\t%s\n", random.NewRandomString(random.ModeOnlyUpper, 64))
	logs.Info("random.ModeOnlyNumber(64):\n\t%s\n", random.NewRandomString(random.ModeOnlyNumber, 64))
	logs.Info("random.ModeHexUpper(64):\n\t%s\n", random.NewRandomString(random.ModeHexUpper, 64))
	logs.Info("random.ModeHexLower(64):\n\t%s\n", random.NewRandomString(random.ModeHexLower, 64))

	logs.Info("random.NewRandomUUID:\n\t%s\n", random.NewRandomUUID())

	logs.Info("--------------------------------------------------------------------------------------------")
}

func testPlatform() {
	logs.Info("--- hello utils/random test ----------------------------------------------------------------")

	logs.Info(platform.GetPlatformInfo())

	logs.Info("--------------------------------------------------------------------------------------------")
}

func testCPU() {
	logs.Info("--- hello utils/cpu test ----------------------------------------------------------------")

	idle0, total0 := cpu.GetCPUTicks()
	time.Sleep(1 * time.Second)
	idle1, total1 := cpu.GetCPUTicks()

	idleTicks := float64(idle1 - idle0)
	totalTicks := float64(total1 - total0)
	cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks

	logs.Info("CPU usage is %.2f [busy: %.0f, total: %.0f]\n", cpuUsage, totalTicks-idleTicks, totalTicks)

	logs.Info("--------------------------------------------------------------------------------------------")
}

func testMutex() {
	var r1 mutex.Mutex
	var r2 mutex.TMutex
	var n1 int
	var n2 int

	logs.Info("--- hello utils/mutex test ----------------------------------------------------------------")

	for i := 0; 30 > i; i++ {
		go func() {
			s := time.Duration(random.NewRandomInt(900) + 100)
			for {
				if r1.TryLock() {
					n1++
					r1.Unlock()
				}
				time.Sleep(s * time.Millisecond)
			}
		}()
	}

	for i := 0; 30 > i; i++ {
		go func() {
			s := time.Duration(random.NewRandomInt(900) + 100)
			for {
				r1.Lock()
				n1++
				r1.Unlock()
				time.Sleep(s * time.Millisecond)
			}
		}()
	}

	for i := 0; 30 > i; i++ {
		go func() {
			s := time.Duration(random.NewRandomInt(900) + 100)
			for {
				if r2.TryLock() {
					n2++
					r2.Unlock()
				}
				time.Sleep(s * time.Millisecond)
			}
		}()
	}

	for i := 0; 30 > i; i++ {
		go func() {
			s := time.Duration(random.NewRandomInt(900) + 100)
			for {
				r2.Lock()
				n2++
				r2.Unlock()
				time.Sleep(s * time.Millisecond)
			}
		}()
	}

	for i := 0; 20 > i; i++ {
		logs.Info("---", i)
		if r1.TryLock() {
			logs.Info(n1)
			r1.Unlock()
		}
		if r2.TryLock() {
			logs.Info(n2)
			r2.Unlock()
		}
		time.Sleep(1 * time.Second)
	}
}

func testHash() {
	logs.Info("--- hello utils/hash test ----------------------------------------------------------------")

	hash.SetGobFormat(true)

	logs.Info(hash.HashToBytes("md5", "123", "456", 123, 456))
	logs.Info(hash.HashToBytes("sha1", "123", "456", 123, 456))
	logs.Info(hash.HashToBytes("sha256", "123", "456", 123, 456))
	logs.Info(hash.HashToBytes("sha512", "123", "456", 123, 456))

	logs.Info(hash.HashToString("md5", "123", "456", 123, 456))
	logs.Info(hash.HashToString("sha1", "123", "456", 123, 456))
	logs.Info(hash.HashToString("sha256", "123", "456", 123, 456))
	logs.Info(hash.HashToString("sha512", "123", "456", 123, 456))

	hash.SetGobFormat(false)

	logs.Info(hash.HashToBytes("md5", "123", "456", 123, 456))
	logs.Info(hash.HashToBytes("sha1", "123", "456", 123, 456))
	logs.Info(hash.HashToBytes("sha256", "123", "456", 123, 456))
	logs.Info(hash.HashToBytes("sha512", "123", "456", 123, 456))

	logs.Info(hash.HashToString("md5", "123", "456", 123, 456))
	logs.Info(hash.HashToString("sha1", "123", "456", 123, 456))
	logs.Info(hash.HashToString("sha256", "123", "456", 123, 456))
	logs.Info(hash.HashToString("sha512", "123", "456", 123, 456))
}

func testSQLite() {
	if db := sqlite.NewSQLiteDB(
		sqlite.WithBackup("test.db", 10, 2048, 32),
	); nil != db {
		db.CreateTable("test1", `id INTEGER PRIMARY KEY AUTOINCREMENT,
								key INTEGER NOT NULL,
								timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))`)
		db.CreateTable("test2", `id INTEGER PRIMARY KEY AUTOINCREMENT,
								key INTEGER NOT NULL,
								timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))`, true)
		db.CreateTable("test3", `id INTEGER PRIMARY KEY AUTOINCREMENT,
								key INTEGER NOT NULL,
								timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))`)
		db.CreateTable("test4", `id INTEGER PRIMARY KEY AUTOINCREMENT,
								key INTEGER NOT NULL,
								timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))`)
		db.CreateTable("test5", `id INTEGER PRIMARY KEY AUTOINCREMENT,
								key INTEGER NOT NULL,
								timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))`, true)
		db.CreateTable("test6", `id INTEGER PRIMARY KEY AUTOINCREMENT,
								key INTEGER NOT NULL,
								timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))`)
		db.CreateTable("test7", `id INTEGER PRIMARY KEY AUTOINCREMENT,
								key INTEGER NOT NULL,
								timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))`, true)
		db.CreateTable("test8", `id INTEGER PRIMARY KEY AUTOINCREMENT,
								key INTEGER NOT NULL,
								timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))`)
		db.CreateTable("test9", `id INTEGER PRIMARY KEY AUTOINCREMENT,
								key INTEGER NOT NULL,
								timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))`, true)

		if n, err := db.StartBackup(true); nil == err {
			logs.Warn("表同步完成，同步条数为%d", n)
		} else {
			logs.Error(err)
		}

		sig := make(chan os.Signal, 1)

		signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

		for {
			select {
			case c := <-sig:
				logs.Warn("Signal: ", c, ", Closing!!!")

				db.Close()

				return
			case <-time.After(100 * time.Millisecond):
				//default:
				if conn, err := db.GetConn(true); nil == err {
					conn.Exec("INSERT INTO test1 (key) VALUES (?);", time.Now().Unix())
					conn.Exec("INSERT INTO test2 (key) VALUES (?);", time.Now().Unix())
					conn.Exec("INSERT INTO test3 (key) VALUES (?);", time.Now().Unix())
					conn.Exec("INSERT INTO test4 (key) VALUES (?);", time.Now().Unix())
					conn.Exec("INSERT INTO test5 (key) VALUES (?);", time.Now().Unix())
					conn.Exec("INSERT INTO test6 (key) VALUES (?);", time.Now().Unix())
					conn.Exec("INSERT INTO test7 (key) VALUES (?);", time.Now().Unix())
					conn.Exec("INSERT INTO test8 (key) VALUES (?);", time.Now().Unix())
					conn.Exec("INSERT INTO test9 (key) VALUES (?);", time.Now().Unix())
				} else {
					logs.Error(err)
				}
			}
		}
	}
}
