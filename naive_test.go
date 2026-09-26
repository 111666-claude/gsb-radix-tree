package radixtree

import "sort"
import "strings"

// naiveTree 是「切片 + 排序 + 线性扫描」的朴素参考实现，
// 口径与 radix 树完全一致，仅用于差分比对。
type naiveTree struct {
	entries []Entry
}

func newNaive() *naiveTree { return &naiveTree{} }

func (n *naiveTree) Put(key string, value int) {
	for i := range n.entries {
		if n.entries[i].Key == key {
			n.entries[i].Value = value
			return
		}
	}
	n.entries = append(n.entries, Entry{Key: key, Value: value})
}

func (n *naiveTree) Get(key string) (int, bool) {
	for _, entry := range n.entries {
		if entry.Key == key {
			return entry.Value, true
		}
	}
	return 0, false
}

func (n *naiveTree) Delete(key string) bool {
	for i, entry := range n.entries {
		if entry.Key == key {
			n.entries = append(n.entries[:i], n.entries[i+1:]...)
			return true
		}
	}
	return false
}

func (n *naiveTree) Len() int { return len(n.entries) }

func (n *naiveTree) Keys() []string {
	keys := make([]string, 0, len(n.entries))
	for _, entry := range n.entries {
		keys = append(keys, entry.Key)
	}
	sort.Strings(keys)
	return keys
}

func (n *naiveTree) WithPrefix(prefix string) []Entry {
	out := []Entry{}
	for _, entry := range n.entries {
		if strings.HasPrefix(entry.Key, prefix) {
			out = append(out, entry)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func (n *naiveTree) LongestPrefix(query string) (string, bool) {
	best := ""
	found := false
	for _, entry := range n.entries {
		if strings.HasPrefix(query, entry.Key) && (!found || len(entry.Key) > len(best)) {
			best = entry.Key
			found = true
		}
	}
	return best, found
}
