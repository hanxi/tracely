package version

// Version 应用版本号，通过 -ldflags 在编译时注入
var Version = "0.0.0"

// BuildTime 构建时间，通过 -ldflags 在编译时注入
var BuildTime = "unknown"

// GitCommit Git 提交哈希，通过 -ldflags 在编译时注入
var GitCommit = "unknown"

// GoVersion Go 版本，通过 -ldflags 在编译时注入
var GoVersion = "unknown"
