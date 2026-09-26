// Package radixtree 是字符串前缀索引。
package radixtree

// Entry 是一条键值。
type Entry struct {
	Key   string
	Value int
}

// node 是压缩前缀树（radix tree）的节点；path 是入边上的压缩片段。
type node struct {
	path     string
	value    int
	hasValue bool
	children []*node // 按 path[0] 字典序排列
}

// Tree 是基于压缩 radix 树的前缀索引。
type Tree struct {
	root     *node
	count    int
	compared int
}

// New 建一棵空树。
func New() *Tree { return &Tree{root: &node{}} }

// Len 是条目数。
func (t *Tree) Len() int { return t.count }

// Compared 是查找过程中沿前缀路径逐字节比较的次数（统计用）。
func (t *Tree) Compared() int { return t.compared }

// findChild 在按首字节排序的子节点中二分查找；返回节点与插入下标。
func findChild(cur *node, b byte) (*node, int) {
	lo, hi := 0, len(cur.children)
	for lo < hi {
		mid := (lo + hi) / 2
		if cur.children[mid].path[0] < b {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if lo < len(cur.children) && cur.children[lo].path[0] == b {
		return cur.children[lo], lo
	}
	return nil, lo
}

// insertChild 把 child 按序插入 children 的 idx 处。
func insertChild(children []*node, idx int, child *node) []*node {
	children = append(children, nil)
	copy(children[idx+1:], children[idx:])
	children[idx] = child
	return children
}

// commonPrefixLen 返回 a、b 的公共前缀长度，并按实际逐字节比较次数计数。
func (t *Tree) commonPrefixLen(a, b string) int {
	limit := len(a)
	if len(b) < limit {
		limit = len(b)
	}
	i := 0
	for i < limit && a[i] == b[i] {
		i++
	}
	t.compared += i
	if i < limit { // 失配的那个字节也比较过一次
		t.compared++
	}
	return i
}

// Put 写入键值；重复键是更新，只留一条且值取最后一次写入。复杂度 O(键长)。
func (t *Tree) Put(key string, value int) {
	if key == "" {
		if !t.root.hasValue {
			t.count++
		}
		t.root.value = value
		t.root.hasValue = true
		return
	}
	cur := t.root
	rem := key
	for {
		child, idx := findChild(cur, rem[0])
		if child == nil {
			cur.children = insertChild(cur.children, idx, &node{
				path:     rem,
				value:    value,
				hasValue: true,
			})
			t.count++
			return
		}
		n := t.commonPrefixLen(rem, child.path)
		if n == len(child.path) {
			rem = rem[n:]
			if rem == "" {
				if !child.hasValue {
					t.count++
				}
				child.value = value
				child.hasValue = true
				return
			}
			cur = child
			continue
		}
		// 在 child.path 的第 n 个字节处分裂。
		mid := &node{path: child.path[:n]}
		child.path = child.path[n:]
		cur.children[idx] = mid
		if n == len(rem) {
			// 新键恰好结束在分裂点。
			mid.value = value
			mid.hasValue = true
			mid.children = []*node{child}
		} else {
			leaf := &node{path: rem[n:], value: value, hasValue: true}
			if leaf.path[0] < child.path[0] {
				mid.children = []*node{leaf, child}
			} else {
				mid.children = []*node{child, leaf}
			}
		}
		t.count++
		return
	}
}

// Get 读取键值。复杂度 O(键长)。
func (t *Tree) Get(key string) (int, bool) {
	if key == "" {
		return t.root.value, t.root.hasValue
	}
	cur := t.root
	rem := key
	for rem != "" {
		child, _ := findChild(cur, rem[0])
		if child == nil {
			return 0, false
		}
		n := t.commonPrefixLen(rem, child.path)
		if n != len(child.path) {
			return 0, false
		}
		rem = rem[n:]
		cur = child
	}
	return cur.value, cur.hasValue
}

// Delete 真正删除键值，返回键是否存在。复杂度 O(键长)。
func (t *Tree) Delete(key string) bool {
	if key == "" {
		if !t.root.hasValue {
			return false
		}
		t.root.hasValue = false
		t.root.value = 0
		t.count--
		return true
	}
	if !t.deleteRec(t.root, key) {
		return false
	}
	t.count--
	return true
}

// deleteRec 从 cur 沿 key 递归删除；key 非空。
func (t *Tree) deleteRec(cur *node, key string) bool {
	child, idx := findChild(cur, key[0])
	if child == nil {
		return false
	}
	n := t.commonPrefixLen(key, child.path)
	if n != len(child.path) {
		return false // 在边内部失配，键不可能存在
	}
	rest := key[n:]
	if rest != "" {
		if !t.deleteRec(child, rest) {
			return false
		}
		t.cleanup(cur, idx)
		return true
	}
	if !child.hasValue {
		return false
	}
	child.hasValue = false
	child.value = 0
	t.cleanup(cur, idx)
	return true
}

// cleanup 在删除后压缩 idx 处的子节点：无值且只有一个孩子就并入该孩子。
func (t *Tree) cleanup(parent *node, idx int) {
	child := parent.children[idx]
	if child.hasValue {
		return
	}
	switch len(child.children) {
	case 0:
		parent.children = append(parent.children[:idx], parent.children[idx+1:]...)
	case 1:
		only := child.children[0]
		child.path += only.path
		child.value = only.value
		child.hasValue = only.hasValue
		child.children = only.children
	}
}

// WithPrefix 按字典序返回所有以 prefix 开头的条目。
// 复杂度 O(前缀长度 + 命中条目数)，不扫描全部键。
func (t *Tree) WithPrefix(prefix string) []Entry {
	out := []Entry{}
	cur := t.root
	rem := prefix
	for rem != "" {
		child, _ := findChild(cur, rem[0])
		if child == nil {
			return out
		}
		n := t.commonPrefixLen(rem, child.path)
		if n == len(child.path) {
			rem = rem[n:]
			if rem == "" {
				t.collect(child, prefix, &out)
				return out
			}
			cur = child
			continue
		}
		if n == len(rem) {
			// 前缀在边内部结束：该子树全部命中。
			t.collect(child, prefix+child.path[n:], &out)
		}
		return out
	}
	t.collect(cur, prefix, &out)
	return out
}

// collect 按字典序 DFS 收集 cur 子树的全部条目。
func (t *Tree) collect(cur *node, prefix string, out *[]Entry) {
	if cur.hasValue {
		*out = append(*out, Entry{Key: prefix, Value: cur.value})
	}
	for _, child := range cur.children {
		t.collect(child, prefix+child.path, out)
	}
}

// Keys 按字典序返回所有键。复杂度 O(条目数)。
func (t *Tree) Keys() []string {
	entries := t.WithPrefix("")
	keys := make([]string, 0, len(entries))
	for _, entry := range entries {
		keys = append(keys, entry.Key)
	}
	return keys
}

// LongestPrefix 返回树中作为 query 前缀的最长键；没有匹配时 ok=false。
// 复杂度 O(查询串长度)，不扫描全部键。
func (t *Tree) LongestPrefix(query string) (string, bool) {
	cur := t.root
	pos := 0
	best := ""
	found := false
	if cur.hasValue {
		found = true
	}
	for pos < len(query) {
		child, _ := findChild(cur, query[pos])
		if child == nil {
			break
		}
		n := t.commonPrefixLen(query[pos:], child.path)
		if n < len(child.path) {
			break // 边未走通，更深的键不可能是 query 前缀
		}
		pos += n
		cur = child
		if cur.hasValue {
			best = query[:pos]
			found = true
		}
	}
	return best, found
}
