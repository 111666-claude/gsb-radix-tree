// Package radixtree 是字符串前缀索引。
package radixtree

import (
	"sort"
	"strings"
)

// Entry 是一条键值。
type Entry struct {
	Key   string
	Value int
}

// Tree 是前缀索引。缺陷：内部就是一根切片，所有操作都线性扫描。
type Tree struct {
	entries  []Entry
	compared int
}

// New 建一棵空树。
func New() *Tree { return &Tree{} }

// Len 是条目数。
func (t *Tree) Len() int { return len(t.entries) }

// Compared 是扫描次数（统计用）。
func (t *Tree) Compared() int { return t.compared }

// Put 写入。缺陷：不查重，重复键会插成两条。
func (t *Tree) Put(key string, value int) {
	t.compared++
	t.entries = append(t.entries, Entry{Key: key, Value: value})
}

// Get 读。缺陷：线性扫描，而且返回的是第一条匹配（旧值）。
func (t *Tree) Get(key string) (int, bool) {
	for _, entry := range t.entries {
		t.compared++
		if entry.Key == key {
			return entry.Value, true
		}
	}
	return 0, false
}

// Delete 删除。缺陷：只把值清零，长度不减。
func (t *Tree) Delete(key string) bool {
	for index, entry := range t.entries {
		t.compared++
		if entry.Key == key {
			t.entries[index].Value = 0
			return true
		}
	}
	return false
}

// WithPrefix 枚举前缀。缺陷：按插入顺序返回，不排序。
func (t *Tree) WithPrefix(prefix string) []Entry {
	out := []Entry{}
	for _, entry := range t.entries {
		t.compared++
		if strings.HasPrefix(entry.Key, prefix) {
			out = append(out, entry)
		}
	}
	return out
}

// LongestPrefix 返回最长匹配前缀的键。缺陷：返回第一个匹配上的，不是最长的。
func (t *Tree) LongestPrefix(key string) (string, bool) {
	for _, entry := range t.entries {
		t.compared++
		if strings.HasPrefix(key, entry.Key) {
			return entry.Key, true
		}
	}
	return "", false
}

// Keys 按字典序返回所有键。
func (t *Tree) Keys() []string {
	keys := make([]string, 0, len(t.entries))
	seen := map[string]bool{}
	for _, entry := range t.entries {
		if seen[entry.Key] {
			continue
		}
		seen[entry.Key] = true
		keys = append(keys, entry.Key)
	}
	sort.Strings(keys)
	return keys
}
