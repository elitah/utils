package exepath

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func GetExePath() string {
	var base string

	switch runtime.GOOS {
	case "linux":
		if file, err := os.Readlink("/proc/self/exe"); nil == err {
			base = file
		}
		if "" == base {
			if cwd, err := os.Readlink("/proc/self/cwd"); nil == err {
				if content, err := ioutil.ReadFile("/proc/self/comm"); nil == err {
					if comm := strings.TrimSpace(string(content)); "" != comm {
						base = filepath.Join(cwd, comm)
					}
				}
			}
		}
	default:
		if file, err := exec.LookPath(os.Args[0]); nil == err {
			base = file
		}
	}

	if "" != base {
		if path, err := filepath.Abs(base); nil == err {
			return path
		}
	}

	return ""
}

func GetExeDir() string {
	if file := GetExePath(); "" != file {
		return filepath.Dir(file)
	}
	return ""
}
