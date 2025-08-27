// Package util 提供MVC框架的文件系统抽象功能
//
// 本文件提供了对http.FileSystem接口的扩展实现，主要用于：
// - 文件系统抽象：提供统一的文件访问接口
// - 目录遍历：递归遍历目录结构
// - 静态文件服务：支持Web静态文件服务
// - 文件系统操作：提供类似filepath.Walk的功能
//
// 设计目标：
// - 兼容标准库：实现http.FileSystem接口
// - 递归遍历：支持深度优先的目录遍历
// - 错误处理：优雅处理文件系统错误
// - 性能优化：减少不必要的文件系统调用
//
// 使用场景：
// - 静态文件服务器
// - 文件索引构建
// - 资源文件扫描
// - 模板文件加载
package util

import (
	"net/http"
	"os"
	"path/filepath"
)

// FileSystem 文件系统结构体
//
// 实现了http.FileSystem接口，提供对本地文件系统的访问能力。
// 这个实现直接使用os.Open来打开文件，适用于访问本地文件系统。
//
// 接口实现：
//   - http.FileSystem.Open(name string) (http.File, error)
//
// 使用示例：
//
//	fs := mvc.FileSystem{}
//	file, err := fs.Open("/path/to/file.txt")
//	if err != nil {
//		// 处理错误
//	}
//	defer file.Close()
//
// 注意事项：
//   - 直接访问本地文件系统，需要注意安全性
//   - 支持相对路径和绝对路径
//   - 返回的http.File支持Seek、Read、Close等操作
type FileSystem struct{}

// Open 打开指定路径的文件或目录
//
// 实现http.FileSystem接口的Open方法，用于打开文件或目录。
// 这个方法直接委托给os.Open，提供对本地文件系统的直接访问。
//
// 参数：
//   - name: string - 要打开的文件或目录路径
//
// 返回值：
//   - http.File: 打开的文件对象，支持Read、Seek、Close等操作
//   - error: 如果文件不存在或权限不足等情况下返回错误
//
// 使用示例：
//
//	fs := mvc.FileSystem{}
//
//	// 打开文件
//	file, err := fs.Open("static/css/style.css")
//	if err != nil {
//		log.Printf("无法打开文件: %v", err)
//		return
//	}
//	defer file.Close()
//
//	// 读取文件内容
//	content, err := io.ReadAll(file)
//	if err != nil {
//		log.Printf("读取文件失败: %v", err)
//	}
//
// 错误情况：
//   - 文件不存在：返回os.ErrNotExist
//   - 权限不足：返回权限相关错误
//   - 路径无效：返回路径格式错误
func (d FileSystem) Open(name string) (http.File, error) {
	return os.Open(name)
}

// Walk 遍历文件系统目录结构
//
// 类似于filepath.Walk函数，但适用于http.FileSystem接口。
// 递归遍历指定根目录下的所有文件和子目录，对每个文件和目录调用walkFn函数。
//
// 参数：
//   - fs: http.FileSystem - 要遍历的文件系统
//   - root: string - 遍历的起始根目录路径
//   - walkFn: filepath.WalkFunc - 对每个文件和目录调用的函数
//
// 返回值：
//   - error: 遍历过程中的错误，如果walkFn返回filepath.SkipDir则正常结束
//
// walkFn函数签名：
//
//	func(path string, info os.FileInfo, err error) error
//
// walkFn返回值处理：
//   - nil: 继续遍历
//   - filepath.SkipDir: 跳过当前目录（如果是目录的话）
//   - 其他error: 停止遍历并返回该错误
//
// 使用示例：
//
//	fs := mvc.FileSystem{}
//
//	err := mvc.Walk(fs, "static", func(path string, info os.FileInfo, err error) error {
//		if err != nil {
//			log.Printf("访问 %s 时出错: %v", path, err)
//			return nil // 继续遍历其他文件
//		}
//
//		if info.IsDir() {
//			fmt.Printf("目录: %s\n", path)
//		} else {
//			fmt.Printf("文件: %s (大小: %d 字节)\n", path, info.Size())
//		}
//
//		return nil
//	})
//
//	if err != nil {
//		log.Printf("遍历失败: %v", err)
//	}
//
// 遍历顺序：
//   - 深度优先遍历
//   - 先访问目录，再访问目录内的文件
//   - 目录内容的访问顺序由文件系统决定
//
// 错误处理：
//   - 如果根目录无法访问，直接返回错误
//   - 如果子目录无法访问，调用walkFn并传递错误信息
//   - walkFn可以决定如何处理错误（继续、跳过或停止）
func Walk(fs http.FileSystem, root string, walkFn filepath.WalkFunc) error {
	f, err := fs.Open(root)
	if err != nil {
		return err
	}
	info, err := f.Stat()
	if err != nil {
		err = walkFn(root, nil, err)
	} else {
		err = walk(fs, root, info, walkFn)
	}
	if err == filepath.SkipDir {
		return nil
	}
	return err
}

// walk 递归遍历目录的内部实现函数
//
// 这是Walk函数的核心递归实现，负责深度优先遍历目录结构。
// 对于每个文件和目录，都会调用提供的walkFn函数进行处理。
//
// 参数：
//   - fs: http.FileSystem - 文件系统接口
//   - path: string - 当前正在遍历的路径
//   - info: os.FileInfo - 当前路径的文件信息
//   - walkFn: filepath.WalkFunc - 处理每个文件和目录的回调函数
//
// 返回值：
//   - error: 遍历过程中的错误或walkFn返回的控制信号
//
// 遍历逻辑：
//  1. 如果是文件，直接调用walkFn并返回结果
//  2. 如果是目录，先调用walkFn处理目录本身
//  3. 然后读取目录内容并递归处理每个子项
//  4. 根据walkFn的返回值决定是否继续遍历
//
// 错误处理策略：
//   - 无法打开目录：调用walkFn传递错误，根据其返回值决定后续行为
//   - 无法读取目录内容：调用walkFn传递错误信息
//   - walkFn返回filepath.SkipDir：跳过当前目录的子项遍历
//   - walkFn返回其他错误：立即停止遍历并返回该错误
//
// 性能考虑：
//   - 使用defer确保目录资源及时释放
//   - 一次性读取所有目录项（Readdir(-1)）以减少系统调用
//   - 深度优先遍历，内存占用相对稳定
func walk(fs http.FileSystem, path string, info os.FileInfo, walkFn filepath.WalkFunc) error {
	var err error

	// 如果是文件，直接处理并返回
	if !info.IsDir() {
		return walkFn(path, info, nil)
	}

	// 如果是目录，尝试打开并读取内容
	dir, err := fs.Open(path)
	if err != nil {
		// 无法打开目录，将错误传递给walkFn
		if err1 := walkFn(path, info, err); err1 != nil {
			return err1
		}
		return err
	}
	defer dir.Close()

	// 读取目录中的所有条目
	dirs, err := dir.Readdir(-1)

	// 先处理目录本身
	err1 := walkFn(path, info, err)

	// 错误处理逻辑：
	// - 如果目录读取失败(err != nil)，walk无法进入该目录
	// - 如果walkFn返回错误(err1 != nil)，表示要跳过该目录或停止遍历
	// - 任何一个条件满足都需要返回，具体行为由walkFn的返回值决定
	if err != nil || err1 != nil {
		// walkFn的返回值控制调用者的行为：
		// - walkFn可能忽略err并返回nil以继续遍历
		// - 如果walkFn返回SkipDir，将由调用者处理
		// - 因此应该返回walkFn的返回值（err1）
		return err1
	}

	// 递归遍历目录中的每个条目
	for _, fileInfo := range dirs {
		filename := filepath.Join(path, fileInfo.Name())
		if err = walk(fs, filename, fileInfo, walkFn); err != nil {
			// 如果是文件出错，或者是目录出错但不是SkipDir，则停止遍历
			if !fileInfo.IsDir() || err != filepath.SkipDir {
				return err
			}
			// 如果是目录且返回SkipDir，则继续遍历下一个条目
		}
	}
	return nil
}
