package main

import (
	"fmt"
	"time"

	"github.com/elitah/utils/cpu"
	"github.com/elitah/utils/exepath"
	"github.com/elitah/utils/hex"
	"github.com/elitah/utils/mutex"
	"github.com/elitah/utils/platform"
	"github.com/elitah/utils/random"
)

func main() {
	fmt.Println("hello utils")

	testExtPath()
	testHex()
	testRandom()
	testPlatform()
	testCPU()
	testMutex()
}

func testExtPath() {
	fmt.Println("--- hello utils/exepath test ----------------------------------------------------------------")

	fmt.Printf("exepath.GetExePath():\n\t%s\n", exepath.GetExePath())
	fmt.Printf("exepath.GetExeDir():\n\t%s\n", exepath.GetExeDir())
}

func testHex() {
	fmt.Println("--- hello utils/hex test ----------------------------------------------------------------")

	data := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66}
	result := hex.EncodeToStringWithSeq(data, ' ')

	fmt.Printf("hex.EncodeToString():\n\t%s\n", hex.EncodeToString(data))
	fmt.Printf("hex.EncodeToStringWithSeq():\n\t%s\n", result)
	if data, err := hex.DecodeStringWithSeq(result); nil == err {
		fmt.Printf("hex.DecodeStringWithSeq():\n\t%x\n", data)
	} else {
		fmt.Printf("hex.DecodeStringWithSeq():\n\t%v\n", err)
	}
}

func testRandom() {
	fmt.Println("--- hello utils/random test ----------------------------------------------------------------")

	fmt.Printf("random.ModeALL(64):\n\t%s\n", random.NewRandomString(random.ModeALL, 64))
	fmt.Printf("random.ModeNoLower(64):\n\t%s\n", random.NewRandomString(random.ModeNoLower, 64))
	fmt.Printf("random.ModeNoUpper(64):\n\t%s\n", random.NewRandomString(random.ModeNoUpper, 64))
	fmt.Printf("random.ModeNoNumber(64):\n\t%s\n", random.NewRandomString(random.ModeNoNumber, 64))
	fmt.Printf("random.ModeNoLowerNumber(64):\n\t%s\n", random.NewRandomString(random.ModeNoLowerNumber, 64))
	fmt.Printf("random.ModeNoUpperNumber(64):\n\t%s\n", random.NewRandomString(random.ModeNoUpperNumber, 64))
	fmt.Printf("random.ModeNoLine(64):\n\t%s\n", random.NewRandomString(random.ModeNoLine, 64))
	fmt.Printf("random.ModeNoLowerLine(64):\n\t%s\n", random.NewRandomString(random.ModeNoLowerLine, 64))
	fmt.Printf("random.ModeNoUpperLine(64):\n\t%s\n", random.NewRandomString(random.ModeNoUpperLine, 64))
	fmt.Printf("random.ModeOnlyLower(64):\n\t%s\n", random.NewRandomString(random.ModeOnlyLower, 64))
	fmt.Printf("random.ModeOnlyUpper(64):\n\t%s\n", random.NewRandomString(random.ModeOnlyUpper, 64))
	fmt.Printf("random.ModeOnlyNumber(64):\n\t%s\n", random.NewRandomString(random.ModeOnlyNumber, 64))
	fmt.Printf("random.ModeHexUpper(64):\n\t%s\n", random.NewRandomString(random.ModeHexUpper, 64))
	fmt.Printf("random.ModeHexLower(64):\n\t%s\n", random.NewRandomString(random.ModeHexLower, 64))

	fmt.Printf("random.NewRandomUUID:\n\t%s\n", random.NewRandomUUID())

	fmt.Println("--------------------------------------------------------------------------------------------")
}

func testPlatform() {
	fmt.Println("--- hello utils/random test ----------------------------------------------------------------")

	fmt.Println(platform.GetPlatformInfo())

	fmt.Println("--------------------------------------------------------------------------------------------")
}

func testCPU() {
	fmt.Println("--- hello utils/cpu test ----------------------------------------------------------------")

	idle0, total0 := cpu.GetCPUTicks()
	time.Sleep(1 * time.Second)
	idle1, total1 := cpu.GetCPUTicks()

	idleTicks := float64(idle1 - idle0)
	totalTicks := float64(total1 - total0)
	cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks

	fmt.Printf("CPU usage is %.2f%% [busy: %.0f, total: %.0f]\n", cpuUsage, totalTicks-idleTicks, totalTicks)

	fmt.Println("--------------------------------------------------------------------------------------------")
}

func testMutex() {
	var r1 mutex.Mutex
	var r2 mutex.TMutex
	var n1 int
	var n2 int

	fmt.Println("--- hello utils/mutex test ----------------------------------------------------------------")

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

	for i := 0; 1000 > i; i++ {
		fmt.Println("---", i)
		if r1.TryLock() {
			fmt.Println(n1)
			r1.Unlock()
		}
		if r2.TryLock() {
			fmt.Println(n2)
			r2.Unlock()
		}
		time.Sleep(1 * time.Second)
	}
}
