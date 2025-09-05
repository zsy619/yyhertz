// Package xio 提供了文件系统和IO操作的工具函数
//
// 主要功能：
//   - 文件操作：读取、写入、复制、移动、删除等
//   - 目录操作：创建、删除、遍历、权限设置等
//   - 路径处理：路径解析、拼接、规范化等
//   - 文件信息：大小、修改时间、权限查询等
//   - 文件系统：磁盘空间、文件系统信息等
//
// 基础用法：
//
//	import "your-project/framework/pkg/xio"
//
//	// 文件基本操作
//	content, err := xio.FileGetContents("config.txt")
//	err = xio.FilePutContents("output.txt", "Hello World", 0644)
//	exists := xio.FileExists("data.json")
//
//	// 目录操作
//	err = xio.Mkdir("uploads", 0755)
//	files, err := xio.ScanDir("./data")
//	size, err := xio.DirSize("./logs")
//
//	// 文件信息
//	size := xio.FileSize("image.jpg")
//	time := xio.FileMTime("config.yaml")
//	isDir := xio.IsDir("uploads")
//	isFile := xio.IsFile("readme.md")
//
// 高级用法：
//
//	// 文件复制和移动
//	err = xio.Copy("source.txt", "backup.txt")
//	err = xio.Rename("old.txt", "new.txt")
//
//	// 权限管理
//	err = xio.Chmod("script.sh", 0755)
//	err = xio.Chown("data.db", 1000, 1000)
//
//	// 路径处理
//	abs, err := xio.RealPath("../config.ini")
//	dir := xio.Dirname("/usr/local/bin/app")
//	base := xio.Basename("/home/user/file.txt")
//
//	// 文件系统信息
//	space, err := xio.DiskTotalSpace("/")
//	free, err := xio.DiskFreeSpace("/var")
//
//	// 临时文件
//	tmp, err := xio.TempNam("/tmp", "upload_")
//	tmpDir := xio.SysTempDir()
//
package xio