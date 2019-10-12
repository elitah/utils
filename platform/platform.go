package platform

import (
	"runtime"
	"strings"
)

type OSFamilyType int
type ArchFamilyType int

const (
	Linux OSFamilyType = iota
	Android
	Windows
	Darwin
	Solaris
	Plan9
	FreeBSD
	OpenBSD
	NetBSD
	NACL
	AIX
	DragonFly

	UnkownOSFamily
)

const (
	AMD64 ArchFamilyType = iota
	ARM
	ARM64
	I386
	MIPS
	MIPS64
	PPC64
	S390X
	WASM

	UnkownArchFamily
)

type PlatformInfo struct {
	OSFamily   OSFamilyType
	OS         string
	ArchFamily ArchFamilyType
	Arch       string
	BigEndian  bool
}

func GetPlatformInfo() PlatformInfo {
	p := PlatformInfo{
		OSFamily:   GetOSFamily(),
		OS:         runtime.GOOS,
		ArchFamily: GetArchFamily(),
		Arch:       runtime.GOARCH,
	}
	p.BigEndian = IsBigEndian(p.ArchFamily)
	return p
}

func GetOSFamily() OSFamilyType {
	switch strings.ToLower(runtime.GOOS) {
	case "linux":
		return Linux
	case "android":
		return Android
	case "windows":
		return Windows
	case "darwin":
		return Darwin
	case "solaris":
		return Solaris
	case "plan9":
		return Plan9
	case "freebsd":
		return FreeBSD
	case "openbsd":
		return OpenBSD
	case "netbsd":
		return NetBSD
	case "nacl":
		return NACL
	case "aix":
		return AIX
	case "dragonfly":
		return DragonFly
	}
	return UnkownOSFamily
}

func GetArchFamily() ArchFamilyType {
	switch strings.ToLower(runtime.GOARCH) {
	case "amd64", "amd64p32":
		return AMD64
	case "arm":
		return ARM
	case "arm64":
		return ARM64
	case "386":
		return I386
	case "mips", "mipsle":
		return MIPS
	case "mips64", "mips64le":
		return MIPS64
	case "ppc64", "ppc64le":
		return PPC64
	case "s390x":
		return S390X
	case "wasm":
		return WASM
	}
	return UnkownArchFamily
}

func IsBigEndian(archFamily ArchFamilyType) bool {
	switch archFamily {
	case MIPS:
		if "mipsle" == runtime.GOARCH {
			return true
		}
	case MIPS64:
		if "mips64le" == runtime.GOARCH {
			return true
		}
	case PPC64:
		if "ppc64le" == runtime.GOARCH {
			return true
		}
	}
	return false
}
