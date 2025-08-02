package util

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileExists checks whether a file or directory exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// IsFile tells whether the filename is a regular file
func IsFile(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

// IsDir tells whether the filename is a directory
func IsDir(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsReadable tells whether a file exists and is readable
func IsReadable(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

// IsWritable tells whether the filename is writable
func IsWritable(filename string) bool {
	if !FileExists(filename) {
		// Check if we can create a file in the directory
		dir := filepath.Dir(filename)
		tempFile := filepath.Join(dir, ".temp_write_test")
		file, err := os.Create(tempFile)
		if err != nil {
			return false
		}
		file.Close()
		os.Remove(tempFile)
		return true
	}

	file, err := os.OpenFile(filename, os.O_WRONLY, 0666)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

// Filesize gets file size
func Filesize(filename string) int64 {
	info, err := os.Stat(filename)
	if err != nil {
		return 0
	}
	return info.Size()
}

// Filemtime gets file modification time
func Filemtime(filename string) int64 {
	info, err := os.Stat(filename)
	if err != nil {
		return 0
	}
	return info.ModTime().Unix()
}

// Filectime gets inode change time of file
func Filectime(filename string) int64 {
	info, err := os.Stat(filename)
	if err != nil {
		return 0
	}
	// In Unix-like systems, this would be the change time
	// Go's standard library doesn't expose ctime directly
	return info.ModTime().Unix()
}

// Fileatime gets last access time of file
func Fileatime(filename string) int64 {
	info, err := os.Stat(filename)
	if err != nil {
		return 0
	}
	// Go's standard library doesn't expose atime directly
	// Return modification time as fallback
	return info.ModTime().Unix()
}

// FileGetContents reads entire file into a string
func FileGetContents(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// FilePutContents writes data to a file
func FilePutContents(filename string, data any, flags ...int) error {
	var content []byte

	switch v := data.(type) {
	case string:
		content = []byte(v)
	case []byte:
		content = v
	default:
		content = []byte(fmt.Sprintf("%v", v))
	}

	// Handle flags if provided
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if len(flags) > 0 {
		// FILE_APPEND equivalent
		if flags[0]&8 != 0 { // FILE_APPEND = 8
			flag = os.O_CREATE | os.O_WRONLY | os.O_APPEND
		}
	}

	file, err := os.OpenFile(filename, flag, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(content)
	return err
}

// File reads entire file into an array
func File(filename string, flags ...int) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		// Handle flags if provided
		if len(flags) == 0 || flags[0]&2 == 0 { // FILE_IGNORE_NEW_LINES = 2
			line += "\n"
		}

		if len(flags) == 0 || flags[0]&1 == 0 { // FILE_SKIP_EMPTY_LINES = 1
			lines = append(lines, line)
		} else if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}

	return lines, scanner.Err()
}

// Copy copies file
func Copy(source, dest string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// Rename renames a file or directory
func Rename(oldname, newname string) error {
	return os.Rename(oldname, newname)
}

// Unlink deletes a file
func Unlink(filename string) error {
	return os.Remove(filename)
}

// Mkdir creates a directory
func Mkdir(pathname string, mode os.FileMode, recursive ...bool) error {
	if len(recursive) > 0 && recursive[0] {
		return os.MkdirAll(pathname, mode)
	}
	return os.Mkdir(pathname, mode)
}

// Rmdir removes a directory
func Rmdir(dirname string) error {
	return os.Remove(dirname)
}

// Dirname returns the parent directory of path
func Dirname(path string) string {
	return filepath.Dir(path)
}

// Basename returns trailing name component of path
func Basename(path string, suffix ...string) string {
	base := filepath.Base(path)
	if len(suffix) > 0 && strings.HasSuffix(base, suffix[0]) {
		base = base[:len(base)-len(suffix[0])]
	}
	return base
}

// Pathinfo returns information about a file path
func Pathinfo(path string, options ...int) map[string]string {
	info := make(map[string]string)

	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(path)
	filename := base
	if ext != "" {
		filename = base[:len(base)-len(ext)]
		ext = ext[1:] // Remove the dot
	}

	const (
		PATHINFO_DIRNAME   = 1
		PATHINFO_BASENAME  = 2
		PATHINFO_EXTENSION = 4
		PATHINFO_FILENAME  = 8
	)

	if len(options) == 0 {
		info["dirname"] = dir
		info["basename"] = base
		info["extension"] = ext
		info["filename"] = filename
	} else {
		flag := options[0]
		if flag&PATHINFO_DIRNAME != 0 {
			info["dirname"] = dir
		}
		if flag&PATHINFO_BASENAME != 0 {
			info["basename"] = base
		}
		if flag&PATHINFO_EXTENSION != 0 {
			info["extension"] = ext
		}
		if flag&PATHINFO_FILENAME != 0 {
			info["filename"] = filename
		}
	}

	return info
}

// Realpath returns the absolute pathname
func Realpath(path string) (string, error) {
	return filepath.Abs(path)
}

// Glob finds pathnames matching a pattern
func Glob(pattern string, flags ...int) ([]string, error) {
	return filepath.Glob(pattern)
}

// Scandir returns an array of files and directories from the given path
func Scandir(directory string, sortingOrder ...int) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		files = append(files, entry.Name())
	}

	// Handle sorting order if provided
	if len(sortingOrder) > 0 && sortingOrder[0] == 1 { // SCANDIR_SORT_DESCENDING
		// Reverse the slice
		for i := len(files)/2 - 1; i >= 0; i-- {
			opp := len(files) - 1 - i
			files[i], files[opp] = files[opp], files[i]
		}
	}

	return files, nil
}

// Fopen opens file or URL
func Fopen(filename, mode string) (*os.File, error) {
	var flag int
	var perm os.FileMode = 0666

	switch mode {
	case "r":
		flag = os.O_RDONLY
	case "r+":
		flag = os.O_RDWR
	case "w":
		flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	case "w+":
		flag = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	case "a":
		flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	case "a+":
		flag = os.O_RDWR | os.O_CREATE | os.O_APPEND
	default:
		flag = os.O_RDONLY
	}

	return os.OpenFile(filename, flag, perm)
}

// Fclose closes a file pointer
func Fclose(file *os.File) error {
	return file.Close()
}

// Fread reads from file pointer
func Fread(file *os.File, length int) ([]byte, error) {
	buffer := make([]byte, length)
	n, err := file.Read(buffer)
	return buffer[:n], err
}

// Fwrite writes to a file pointer
func Fwrite(file *os.File, data []byte) (int, error) {
	return file.Write(data)
}

// Fgets reads line from file pointer
func Fgets(file *os.File) (string, error) {
	reader := bufio.NewReader(file)
	line, err := reader.ReadString('\n')
	return line, err
}

// Feof tests for end-of-file on a file pointer
func Feof(file *os.File) bool {
	// Try to read one byte and check for EOF
	buffer := make([]byte, 1)
	_, err := file.Read(buffer)
	if err == io.EOF {
		return true
	}
	// Seek back if we successfully read
	if err == nil {
		file.Seek(-1, io.SeekCurrent)
	}
	return false
}

// Ftell returns the current position of the file read/write pointer
func Ftell(file *os.File) (int64, error) {
	return file.Seek(0, io.SeekCurrent)
}

// Fseek seeks on a file pointer
func Fseek(file *os.File, offset int64, whence int) (int64, error) {
	return file.Seek(offset, whence)
}

// Rewind rewinds the position of a file pointer
func Rewind(file *os.File) error {
	_, err := file.Seek(0, io.SeekStart)
	return err
}

// Touch sets access and modification time of file
func Touch(filename string, mtime ...time.Time) error {
	var modTime time.Time
	if len(mtime) > 0 {
		modTime = mtime[0]
	} else {
		modTime = time.Now()
	}

	// Create file if it doesn't exist
	if !FileExists(filename) {
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		file.Close()
	}

	return os.Chtimes(filename, modTime, modTime)
}

// ChmodRecursive recursively changes file permissions
func ChmodRecursive(path string, mode os.FileMode) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err == nil {
			err = os.Chmod(name, mode)
		}
		return err
	})
}

// DiskFreeSpace returns available space on filesystem
func DiskFreeSpace(directory string) (int64, error) {
	// This is a simplified implementation
	// On Unix systems, you would use syscall.Statfs
	// For cross-platform compatibility, return file info
	info, err := os.Stat(directory)
	if err != nil {
		return 0, err
	}
	// This is not accurate disk space, just a placeholder
	_ = info
	return 0, fmt.Errorf("disk_free_space not implemented for this platform")
}

// TempNam creates a file with a unique filename
func TempNam(dir, prefix string) (string, error) {
	file, err := os.CreateTemp(dir, prefix)
	if err != nil {
		return "", err
	}
	name := file.Name()
	file.Close()
	return name, nil
}
