package version

const VerboseVersionBanner string = `%s%s
__     ____     ___    _           _       
\ \   / /\ \   / / |  | |         | |      
 \ \_/ /  \ \_/ /| |__| | ___ _ __| |_ ____
  \   /    \   / |  __  |/ _ \ '__| __|_  /
   | |      | |  | |  | |  __/ |  | |_ / / 
   |_|      |_|  |_|  |_|\___|_|   \__/___| v{{ .Version }}
                                           
%s%s
├── GoVersion : {{ .GoVersion }}
├── GOOS      : {{ .GOOS }}
├── GOARCH    : {{ .GOARCH }}
├── NumCPU    : {{ .NumCPU }}
├── GOPATH    : {{ .GOPATH }}
├── GOROOT    : {{ .GOROOT }}
├── Compiler  : {{ .Compiler }}
└── Date      : {{ Now "Monday, 2 Jan 2006" }}%s%s
`

const ShortVersionBanner = `______
| ___ \
| |_/ /  ___   ___
| ___ \ / _ \ / _ \
| |_/ /|  __/|  __/
\____/  \___| \___| v{{ .Version }}
`

const (
	// Version 框架版本
	Version = "1.4.0"
)
