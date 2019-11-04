package cpu

import (
	"io/ioutil"
	"runtime"
	"strconv"
	"strings"
)

func GetCPUTicks() (idle, total uint64) {
	if "linux" == strings.ToLower(runtime.GOOS) {
		if contents, err := ioutil.ReadFile("/proc/stat"); nil == err {
			// 分割每一行
			lines := strings.Split(string(contents), "\n")
			// 遍历每一行
			for _, line := range lines {
				// 分割数字字符串
				fields := strings.Fields(line)
				// 如果是所有CPU
				if fields[0] == "cpu" {
					// 数值总个数
					numFields := len(fields)
					// 遍历参数
					for i := 1; i < numFields; i++ {
						// 转换数值
						if val, err := strconv.ParseUint(fields[i], 10, 64); nil == err {
							if 4 == i {
								idle = val
							}
							total += val
						}
					}
					return
				}
			}
		}
	}
	return
}
