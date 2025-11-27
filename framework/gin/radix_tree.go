// Package gin - 高性能Radix Tree路由实现
// 基于Gin原版优化，支持零分配路由匹配
package gin

import (
	"strings"
	"unicode/utf8"
)

// nodeType 节点类型
type nodeType uint8

const (
	static   nodeType = iota // 静态节点: /users
	root                     // 根节点: /
	param                    // 参数节点: /:id
	catchAll                 // 通配符节点: /*path
)

// 注意：Param和Params类型已在gin_core.go中定义

// node Radix Tree节点结构
type node struct {
	path      string      // 节点路径
	indices   string      // 子节点索引字符串
	wildChild bool        // 是否有通配符子节点
	nType     nodeType    // 节点类型
	priority  uint32      // 节点优先级（访问频率）
	children  []*node     // 子节点列表
	handlers  HandlerChain // 处理函数链
	fullPath  string      // 完整路径（用于重建URL）
}

// HandlerChain 处理函数链类型
type HandlerChain []HandlerFunc

// Last 返回链中最后一个处理函数
func (c HandlerChain) Last() HandlerFunc {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}

// methodTree 方法树结构
type methodTree struct {
	method string
	root   *node
}

// methodTrees 方法树集合
type methodTrees []methodTree

// get 获取指定方法的树
func (trees methodTrees) get(method string) *node {
	for _, tree := range trees {
		if tree.method == method {
			return tree.root
		}
	}
	return nil
}

// min 返回两个数的最小值
func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

// longestCommonPrefix 计算两个字符串的最长公共前缀长度
func longestCommonPrefix(a, b string) int {
	i := 0
	max := min(len(a), len(b))
	for i < max && a[i] == b[i] {
		i++
	}
	return i
}

// addRoute 添加路由到radix tree
func (n *node) addRoute(path string, handlers HandlerChain) {
	fullPath := path
	n.priority++

	// 空树直接插入
	if len(n.path) == 0 && len(n.children) == 0 {
		n.insertChild(path, fullPath, handlers)
		n.nType = root
		return
	}

	parentFullPathIndex := 0

walk:
	for {
		// 找到最长公共前缀
		i := longestCommonPrefix(path, n.path)

		// 分割边：当前节点需要分割
		if i < len(n.path) {
			child := node{
				path:      n.path[i:],
				wildChild: n.wildChild,
				indices:   n.indices,
				children:  n.children,
				handlers:  n.handlers,
				priority:  n.priority - 1,
				fullPath:  n.fullPath,
			}

			// 更新当前节点
			n.children = []*node{&child}
			n.indices = string([]byte{n.path[i]})
			n.path = path[:i]
			n.handlers = nil
			n.wildChild = false
			n.fullPath = fullPath[:parentFullPathIndex+i]
		}

		// 插入新路由
		if i < len(path) {
			path = path[i:]
			c := path[0]

			// 检查参数节点
			if n.nType == param && c == '/' && len(n.children) == 1 {
				parentFullPathIndex += len(n.path)
				n = n.children[0]
				n.priority++
				continue walk
			}

			// 检查现有子节点
			for i, max := 0, len(n.indices); i < max; i++ {
				if c == n.indices[i] {
					parentFullPathIndex += len(n.path)
					i = n.incrementChildPrio(i)
					n = n.children[i]
					continue walk
				}
			}

			// 插入新子节点
			if c != ':' && c != '*' && n.nType != catchAll {
				// 添加到indices并创建子节点
				n.indices += string([]byte{c})
				child := &node{
					fullPath: fullPath,
				}
				n.addChild(child)
				n.incrementChildPrio(len(n.indices) - 1)
				n = child
			} else if n.wildChild {
				// 通配符冲突检查
				n = n.children[len(n.children)-1]
				n.priority++

				// 检查通配符是否匹配
				if len(path) >= len(n.path) && n.path == path[:len(n.path)] &&
					n.nType != catchAll &&
					(len(n.path) >= len(path) || path[len(n.path)] == '/') {
					continue walk
				}

				pathSeg := path
				if n.nType != catchAll {
					pathSeg = strings.SplitN(pathSeg, "/", 2)[0]
				}
				prefix := fullPath[:strings.Index(fullPath, pathSeg)] + n.path
				panic("'" + pathSeg + "' in new path '" + fullPath +
					"' conflicts with existing wildcard '" + n.path +
					"' in existing prefix '" + prefix + "'")
			}

			n.insertChild(path, fullPath, handlers)
			return
		}

		// 路径已存在，更新处理函数
		if n.handlers != nil {
			panic("handlers are already registered for path '" + fullPath + "'")
		}
		n.handlers = handlers
		n.fullPath = fullPath
		return
	}
}

// addChild 添加子节点
func (n *node) addChild(child *node) {
	n.children = append(n.children, child)
}

// incrementChildPrio 增加子节点优先级并重排序
func (n *node) incrementChildPrio(pos int) int {
	cs := n.children
	cs[pos].priority++
	prio := cs[pos].priority

	// 调整位置以保持优先级顺序
	newPos := pos
	for ; newPos > 0 && cs[newPos-1].priority < prio; newPos-- {
		cs[newPos], cs[newPos-1] = cs[newPos-1], cs[newPos]
	}

	// 同步调整indices
	if newPos != pos {
		n.indices = n.indices[:newPos] +
			n.indices[pos:pos+1] +
			n.indices[newPos:pos] + n.indices[pos+1:]
	}

	return newPos
}

// insertChild 插入子节点
func (n *node) insertChild(path, fullPath string, handlers HandlerChain) {
	for {
		// 查找参数
		wildcard, i, valid := findWildcard(path)
		if i < 0 { // 没有通配符
			break
		}

		// 通配符名称必须有效
		if !valid {
			panic("only one wildcard per path segment is allowed, has: '" +
				wildcard + "' in path '" + fullPath + "'")
		}

		// 检查通配符名称是否有效
		if len(wildcard) < 2 {
			panic("wildcards must be named with a non-empty name in path '" + fullPath + "'")
		}

		if wildcard[0] == ':' { // 参数
			if i > 0 {
				// 在参数前插入前缀
				n.path = path[:i]
				path = path[i:]
			}

			child := &node{
				nType:    param,
				path:     wildcard,
				fullPath: fullPath,
			}
			n.addChild(child)
			n.wildChild = true
			n = child
			n.priority++

			// 如果路径没有以通配符结尾，继续处理
			if len(wildcard) < len(path) {
				path = path[len(wildcard):]
				child := &node{
					priority: 1,
					fullPath: fullPath,
				}
				n.addChild(child)
				n = child
				continue
			}

			// 设置处理函数
			n.handlers = handlers
			return

		} else { // catchAll
			if i+len(wildcard) != len(path) {
				panic("catch-all routes are only allowed at the end of the path in path '" + fullPath + "'")
			}

			if len(n.path) > 0 && n.path[len(n.path)-1] == '/' {
				pathSeg := strings.SplitN(n.children[0].path, "/", 2)[0]
				panic("catch-all wildcard '" + path +
					"' in new path '" + fullPath +
					"' conflicts with existing path segment '" + pathSeg +
					"' in existing prefix '" + n.fullPath + pathSeg + "'")
			}

			// 当前路径必须以'/'结尾
			i--
			if path[i] != '/' {
				panic("no / before catch-all in path '" + fullPath + "'")
			}

			n.path = path[:i]

			// 第一个节点：catchAll之前的路径
			child := &node{
				wildChild: true,
				nType:     catchAll,
				fullPath:  fullPath,
			}

			n.addChild(child)
			n.indices = string('/')
			n = child
			n.priority++

			// 第二个节点：保存变量
			child = &node{
				path:     path[i:],
				nType:    catchAll,
				handlers: handlers,
				priority: 1,
				fullPath: fullPath,
			}
			n.children = []*node{child}
			return
		}
	}

	// 没有通配符，直接插入
	n.path = path
	n.handlers = handlers
	n.fullPath = fullPath
}

// findWildcard 查找通配符
func findWildcard(path string) (wildcard string, i int, valid bool) {
	for start, c := range []byte(path) {
		if c != ':' && c != '*' {
			continue
		}

		// 查找通配符结束位置
		valid = true
		for end, c := range []byte(path[start+1:]) {
			switch c {
			case '/':
				return path[start : start+1+end], start, valid
			case ':', '*':
				valid = false
			}
		}
		return path[start:], start, valid
	}
	return "", -1, false
}

// getValue 获取路由值和参数
func (n *node) getValue(path string, params func() *Params) (handlers HandlerChain, p *Params, tsr bool) {
walk:
	for {
		prefix := n.path
		if len(path) > len(prefix) {
			if path[:len(prefix)] == prefix {
				path = path[len(prefix):]

				// 尝试所有子节点
				idxc := path[0]
				for i, c := range []byte(n.indices) {
					if c == idxc {
						n = n.children[i]
						continue walk
					}
				}

				// 如果没有匹配的子节点且有通配符子节点
				if n.wildChild {
					n = n.children[len(n.children)-1]
					
					switch n.nType {
					case param:
						// 查找参数结束位置
						end := 0
						for end < len(path) && path[end] != '/' {
							end++
						}

						// 保存参数
						if params != nil {
							if p == nil {
								p = params()
								*p = (*p)[:0] // 重置slice但保持容量
							}
							
							val := path[:end]
							*p = append(*p, Param{
								Key:   n.path[1:], // 跳过 ':'
								Value: val,
							})
						}

						// 继续处理剩余路径
						if end < len(path) {
							if len(n.children) > 0 {
								path = path[end:]
								n = n.children[0]
								continue walk
							}

							// 没有更多子节点，但路径还有剩余
							tsr = (len(path) == 1 && path[0] == '/')
							return
						}

						if handlers = n.handlers; handlers != nil {
							return
						}
						
						if len(n.children) == 1 {
							// 尾随斜杠重定向
							n = n.children[0]
							tsr = (n.path == "/" && n.handlers != nil)
						}
						return

					case catchAll:
						// 保存catch-all参数
						if params != nil {
							if p == nil {
								p = params()
								*p = (*p)[:0]
							}
							
							*p = append(*p, Param{
								Key:   n.path[2:], // 跳过 '*/'
								Value: path,
							})
						}

						handlers = n.handlers
						return

					default:
						panic("invalid node type")
					}
				}

				// 尾随斜杠重定向检查
				if path == "/" && n.wildChild && n.nType != root {
					tsr = true
					return
				}

				// 没有找到匹配
				return
			}
		} else if path == prefix {
			// 精确匹配
			if handlers = n.handlers; handlers != nil {
				return
			}

			// 尾随斜杠重定向
			if path == "/" && n.wildChild && n.nType != root {
				tsr = true
				return
			}

			for i, c := range []byte(n.indices) {
				if c == '/' {
					n = n.children[i]
					tsr = (len(n.path) == 1 && n.handlers != nil) ||
						(n.nType == catchAll && n.children[0].handlers != nil)
					return
				}
			}

			return
		}

		// 没有匹配，检查尾随斜杠重定向
		tsr = (path == "/") ||
			(len(prefix) == len(path)+1 && prefix[len(path)] == '/' &&
				path == prefix[:len(prefix)-1] && n.handlers != nil)
		return
	}
}

// findCaseInsensitivePath 查找大小写不敏感的路径
func (n *node) findCaseInsensitivePath(path string, fixTrailingSlash bool) (fixedPath string, found bool) {
	const stackBufSize = 128

	// 使用较小的缓冲区
	buf := make([]byte, 0, stackBufSize)
	if l := len(path) + 1; l > stackBufSize {
		buf = make([]byte, 0, l)
	}

	ciPath := n.findCaseInsensitivePathRec(
		path,
		buf,       // 使用pre-allocated buffer
		[4]byte{}, // 预分配rune buffer
		fixTrailingSlash,
	)

	return ciPath, ciPath != ""
}

// findCaseInsensitivePathRec 递归查找大小写不敏感路径
func (n *node) findCaseInsensitivePathRec(path string, ciPath []byte, rb [4]byte, fixTrailingSlash bool) string {
	npLen := len(n.path)

walk:
	for len(path) >= npLen && (npLen == 0 || strings.EqualFold(path[1:npLen], n.path[1:])) {
		// 将当前节点路径添加到结果中
		path = path[npLen:]
		ciPath = append(ciPath, n.path...)

		if len(path) == 0 {
			// 已到达末尾
			if n.handlers != nil {
				return string(ciPath)
			}

			// 查找带斜杠的处理器
			for i, c := range []byte(n.indices) {
				if c == '/' {
					n = n.children[i]
					if (len(n.path) == 1 && n.handlers != nil) ||
						(n.nType == catchAll && n.children[0].handlers != nil) {
						return string(ciPath) + "/"
					}
					return ""
				}
			}

			// 检查尾随斜杠重定向
			if fixTrailingSlash {
				for i, c := range []byte(n.indices) {
					if c == '/' {
						n = n.children[i]
						if (len(n.path) == 1 && n.handlers != nil) ||
							(n.nType == catchAll && n.children[0].handlers != nil) {
							return string(ciPath) + "/"
						}
						break
					}
				}
			}
			return ""
		}

		// 如果此节点没有通配符子节点，我们可以继续查找下一个子节点
		if !n.wildChild {
			// 跳过rune为非ASCII的情况
			rb = shiftNRuneBytes(rb, path[0])
			if rb[3] != 0 {
				// 可能是rune的开始或中间，让我们完成它
				for i := 1; i < len(path); i++ {
					rb = shiftNRuneBytes(rb, path[i])
					if rb[3] == 0 {
						// 完整的rune
						path = path[i+1:]
						// 查找匹配的子节点
						goto nextChild
					}
				}
				return ""
			}

		nextChild:
			// 查找匹配的子节点
			for i, c := range []byte(n.indices) {
				// 大小写不敏感比较
				if c == path[0] || (c >= 'A' && c <= 'Z' && c+'a'-'A' == path[0]) ||
					(path[0] >= 'A' && path[0] <= 'Z' && path[0]+'a'-'A' == c) {
					// 找到匹配的子节点
					out := n.children[i].findCaseInsensitivePathRec(
						path, ciPath, rb, fixTrailingSlash,
					)
					if out != "" {
						return out
					}
				}
			}

			// 没有找到匹配的子节点
			if fixTrailingSlash && path == "/" && n.handlers != nil {
				return string(ciPath)
			}
			return ""
		}

		n = n.children[len(n.children)-1]
		switch n.nType {
		case param:
			// 查找param结束位置
			end := 0
			for end < len(path) && path[end] != '/' {
				end++
			}

			// 将参数值添加到路径中
			ciPath = append(ciPath, path[:end]...)

			// 继续处理剩余路径
			if end < len(path) {
				if len(n.children) > 0 {
					path = path[end:]
					n = n.children[0]
					npLen = len(n.path)
					continue walk
				}

				// 没有更多子节点
				if fixTrailingSlash && len(path) == 1 && path[0] == '/' {
					return string(ciPath) + "/"
				}
				return ""
			}

			if n.handlers != nil {
				return string(ciPath)
			} else if fixTrailingSlash && len(n.children) == 1 {
				// 没有处理器，检查TSR
				n = n.children[0]
				if n.path == "/" && n.handlers != nil {
					return string(ciPath) + "/"
				}
			}
			return ""

		case catchAll:
			ciPath = append(ciPath, path...)
			return string(ciPath)

		default:
			panic("invalid node type")
		}
	}

	// 没有匹配
	if fixTrailingSlash {
		if len(path)+1 == npLen && n.path[len(path)] == '/' &&
			strings.EqualFold(path[1:], n.path[1:len(path)]) && n.handlers != nil {
			return string(append(ciPath, n.path...))
		}
	}
	return ""
}

// shiftNRuneBytes 移位并添加新字节到rune缓冲区
func shiftNRuneBytes(rb [4]byte, b byte) [4]byte {
	switch {
	case b < utf8.RuneSelf:
		// ASCII
		return [4]byte{b}
	case b < 0xC0:
		// 延续字节
		return [4]byte{rb[1], rb[2], rb[3], b}
	default:
		// 新的多字节序列开始
		return [4]byte{b}
	}
}