// Package radixtree 是字符串前缀索引。
//
// 内部用压缩基数树（compressed trie / radix tree）实现：
//   - Get / Put / Delete 沿查询串下行，比较量只与键长有关，O(键长)；
//   - WithPrefix 走到前缀对应位置后只遍历前缀子树，按字典序产出；
//   - LongestPrefix 沿查询串下行时记录沿途命中的终端节点，不扫描全部键。
package radixtree

import "sort"

// Entry 是一条键值。
type Entry struct {
	Key   string
	Value int
}

// node 是基数树节点。prefix 是父边压在该节点上的字符串，children 的键是
// 子节点 prefix 的第一个字节（根节点 prefix 为空）。
type node struct {
	prefix   string
	value    int
	terminal bool
	children map[byte]*node
}

func newNode(prefix string) *node {
	return &node{prefix: prefix, children: map[byte]*node{}}
}

// Tree 是前缀索引。
type Tree struct {
	root     *node
	size     int
	compared int
}

// New 建一棵空树。
func New() *Tree {
	return &Tree{root: newNode("")}
}

// Len 是条目数。
func (t *Tree) Len() int { return t.size }

// Compared 是下行过程中与边标签比较过的次数（每走一条边计一次，
// 对应线性实现里“与多少条记录比较过”）（统计用）。
func (t *Tree) Compared() int { return t.compared }

// matchEdge 沿一条边比较，返回公共前缀长度；每比较一条边计一次 compared。
// 单次调用只处理一条边，因此整条下行路径是 O(键长)，不会扫到其它分支。
func (t *Tree) matchEdge(a, b string) int {
	t.compared++
	limit := len(a)
	if len(b) < limit {
		limit = len(b)
	}
	index := 0
	for index < limit {
		if a[index] != b[index] {
			break
		}
		index++
	}
	return index
}

// commonPrefix 返回 a、b 的公共前缀长度。
func (t *Tree) commonPrefix(a, b string) int {
	return t.matchEdge(a, b)
}

// Put 写入；重复键是更新，只留一条、值取最后一次写入。O(键长)。
func (t *Tree) Put(key string, value int) {
	current := t.root
	remaining := key
	for {
		if len(remaining) == 0 {
			if !current.terminal {
				t.size++
			}
			current.terminal = true
			current.value = value
			return
		}
		child := current.children[remaining[0]]
		if child == nil {
			child = newNode(remaining)
			child.terminal = true
			child.value = value
			current.children[remaining[0]] = child
			t.size++
			return
		}
		shared := t.commonPrefix(child.prefix, remaining)
		switch {
		case shared == len(child.prefix):
			// 整条边吃完，继续向下走。
			current = child
			remaining = remaining[shared:]
		case shared == len(remaining):
			// 新键是这条边的前缀：在中间插一个终端节点。
			middle := newNode(remaining)
			middle.terminal = true
			middle.value = value
			child.prefix = child.prefix[shared:]
			middle.children[child.prefix[0]] = child
			current.children[remaining[0]] = middle
			t.size++
			return
		default:
			// 在 shared 处分裂。
			middle := newNode(remaining[:shared])
			child.prefix = child.prefix[shared:]
			leaf := newNode(remaining[shared:])
			leaf.terminal = true
			leaf.value = value
			middle.children[child.prefix[0]] = child
			middle.children[remaining[shared]] = leaf
			current.children[remaining[0]] = middle
			t.size++
			return
		}
	}
}

// Get 读取。O(键长)。
func (t *Tree) Get(key string) (int, bool) {
	current := t.root
	remaining := key
	for {
		if len(remaining) == 0 {
			return current.value, current.terminal
		}
		child := current.children[remaining[0]]
		if child == nil {
			return 0, false
		}
		shared := t.commonPrefix(child.prefix, remaining)
		if shared != len(child.prefix) {
			return 0, false
		}
		current = child
		remaining = remaining[shared:]
	}
}

// frame 记录下行路径上的节点以及从父节点进入它所用的字节（根为 0）。
type frame struct {
	node *node
	via  byte
}

// Delete 真正删除：Len 减一、Keys 不再出现、Get 读不到。O(键长)。
func (t *Tree) Delete(key string) bool {
	path := []frame{{node: t.root}}
	current := t.root
	remaining := key
	for {
		if len(remaining) == 0 {
			if !current.terminal {
				return false
			}
			current.terminal = false
			current.value = 0
			t.size--
			break
		}
		child := current.children[remaining[0]]
		if child == nil {
			return false
		}
		shared := t.commonPrefix(child.prefix, remaining)
		if shared != len(child.prefix) {
			return false
		}
		path = append(path, frame{node: child, via: remaining[0]})
		current = child
		remaining = remaining[shared:]
	}

	// 自底向上剪枝：空的非根节点删掉；删掉后若父节点变成
	// “非终端且只有一个孩子”，就把父子合并回一条边。
	for depth := len(path) - 1; depth >= 1; depth-- {
		target := path[depth].node
		if target.terminal || len(target.children) > 0 {
			break
		}
		parent := path[depth-1].node
		delete(parent.children, path[depth].via)

		if parent == t.root || parent.terminal || len(parent.children) != 1 {
			break
		}
		var only *node
		for _, only = range parent.children {
		}
		parent.prefix += only.prefix
		parent.terminal = only.terminal
		parent.value = only.value
		parent.children = only.children
		grand := path[depth-2].node
		grand.children[path[depth-1].via] = parent
		break
	}
	return true
}

// sortedChildKeys 返回按字典序排好的子边首字节。
func sortedChildKeys(children map[byte]*node) []byte {
	out := make([]byte, 0, len(children))
	for first := range children {
		out = append(out, first)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// collect 按字典序深度优先收集子树中的终端条目。
// base 是当前节点对应的完整前缀（已含 current.prefix，根为 ""）。
func collect(current *node, base string, out *[]Entry) {
	if current.terminal {
		*out = append(*out, Entry{Key: base, Value: current.value})
	}
	for _, first := range sortedChildKeys(current.children) {
		child := current.children[first]
		collect(child, base+child.prefix, out)
	}
}

// WithPrefix 按字典序返回所有前缀匹配的条目，只遍历前缀子树。
func (t *Tree) WithPrefix(prefix string) []Entry {
	out := []Entry{}
	current := t.root
	remaining := prefix
	base := ""
	for {
		if len(remaining) == 0 {
			collect(current, base, &out)
			return out
		}
		child := current.children[remaining[0]]
		if child == nil {
			return out
		}
		shared := t.commonPrefix(child.prefix, remaining)
		switch {
		case shared == len(child.prefix):
			base += child.prefix
			current = child
			remaining = remaining[shared:]
		case shared == len(remaining):
			// 前缀是边标签的前缀：整棵子树都在范围内。
			collect(child, base+child.prefix, &out)
			return out
		default:
			return out
		}
	}
}

// Keys 按字典序返回所有键。
func (t *Tree) Keys() []string {
	entries := t.WithPrefix("")
	keys := make([]string, 0, len(entries))
	for _, entry := range entries {
		keys = append(keys, entry.Key)
	}
	return keys
}

// LongestPrefix 返回查询串已有的最长终端前缀；无匹配时 ok=false。
// 沿查询串下行，不扫描全部键。
func (t *Tree) LongestPrefix(query string) (string, bool) {
	bestKey := ""
	bestLen := -1
	if t.root.terminal {
		bestLen = 0
	}
	current := t.root
	remaining := query
	depth := 0
	for len(remaining) > 0 {
		child := current.children[remaining[0]]
		if child == nil {
			break
		}
		shared := t.commonPrefix(child.prefix, remaining)
		depth += shared
		if shared < len(child.prefix) {
			// 边没有吃完；只有当查询串也恰好走到尽头时该节点才算命中，
			// 而此时 shared == len(remaining)，节点尚未下行，下面单独处理。
			break
		}
		current = child
		remaining = remaining[shared:]
		if current.terminal && depth > bestLen {
			bestLen = depth
			bestKey = query[:depth]
		}
	}
	if bestLen < 0 {
		return "", false
	}
	return bestKey, true
}
