package main

import (
	"fmt"
	"time"

	"github.com/elitah/utils/cpu"
	"github.com/elitah/utils/platform"
	"github.com/elitah/utils/random"
)

func main() {
	fmt.Println("hello utils")

	testRandom()
	testPlatform()
	testCPU()
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
