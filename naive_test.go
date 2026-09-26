package radixtree

import (
	"math/rand"
	"sort"
	"strings"
	"testing"
)

// naiveTree 是「切片 + 排序 + 线性扫描」的朴素参考实现：
// 重复键就地更新、删除物理移除，所有查询都先扫描/再排序，结果作为对照基准。
type naiveTree struct {
	entries []Entry
}

func newNaive() *naiveTree { return &naiveTree{} }

func (n *naiveTree) put(key string, value int) {
	for index := range n.entries {
		if n.entries[index].Key == key {
			n.entries[index].Value = value
			return
		}
	}
	n.entries = append(n.entries, Entry{Key: key, Value: value})
}

func (n *naiveTree) get(key string) (int, bool) {
	for _, entry := range n.entries {
		if entry.Key == key {
			return entry.Value, true
		}
	}
	return 0, false
}

func (n *naiveTree) delete(key string) bool {
	for index, entry := range n.entries {
		if entry.Key == key {
			n.entries = append(n.entries[:index], n.entries[index+1:]...)
			return true
		}
	}
	return false
}

func (n *naiveTree) keys() []string {
	out := make([]string, 0, len(n.entries))
	for _, entry := range n.entries {
		out = append(out, entry.Key)
	}
	sort.Strings(out)
	return out
}

func (n *naiveTree) withPrefix(prefix string) []Entry {
	out := []Entry{}
	for _, entry := range n.entries {
		if strings.HasPrefix(entry.Key, prefix) {
			out = append(out, entry)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func (n *naiveTree) longestPrefix(query string) (string, bool) {
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

type op struct {
	kind  int
	key   string
	value int
}

var compareKeyPool = []string{
	"", "a", "ap", "app", "apple", "apples", "applesauce",
	"b", "ban", "banana", "band", "bandana",
	"c", "cat", "catalog", "dog", "",
	"key000001", "key000010", "key000100", "key001000", "key010000",
}

func assertSameState(t *testing.T, tree *Tree, naive *naiveTree, step int) {
	t.Helper()
	if tree.Len() != len(naive.entries) {
		t.Fatalf("第 %d 步 Len 不一致：radix=%d naive=%d", step, tree.Len(), len(naive.entries))
	}
	treeKeys := tree.Keys()
	naiveKeys := naive.keys()
	if len(treeKeys) != len(naiveKeys) {
		t.Fatalf("第 %d 步 Keys 长度不一致：%v vs %v", step, treeKeys, naiveKeys)
	}
	for i := range treeKeys {
		if treeKeys[i] != naiveKeys[i] {
			t.Fatalf("第 %d 步 Keys 不一致：%v vs %v", step, treeKeys, naiveKeys)
		}
	}
	for _, key := range naiveKeys {
		rv, rok := tree.Get(key)
		nv, nok := naive.get(key)
		if rok != nok || rv != nv {
			t.Fatalf("第 %d 步 Get(%q) 不一致：%d/%v vs %d/%v", step, key, rv, rok, nv, nok)
		}
	}
	for _, prefix := range compareKeyPool {
		rt := tree.WithPrefix(prefix)
		nt := naive.withPrefix(prefix)
		if len(rt) != len(nt) {
			t.Fatalf("第 %d 步 WithPrefix(%q) 长度不一致：%v vs %v", step, prefix, rt, nt)
		}
		for i := range rt {
			if rt[i] != nt[i] {
				t.Fatalf("第 %d 步 WithPrefix(%q) 不一致：%v vs %v", step, prefix, rt, nt)
			}
		}
		rl, rok := tree.LongestPrefix(prefix)
		nl, nok := naive.longestPrefix(prefix)
		if rok != nok || rl != nl {
			t.Fatalf("第 %d 步 LongestPrefix(%q) 不一致：%q/%v vs %q/%v",
				step, prefix, rl, rok, nl, nok)
		}
	}
}

// TestCrossCheckAgainstNaive 用 1000 组随机操作序列，逐条比对
// Get / Keys / WithPrefix / LongestPrefix。
func TestCrossCheckAgainstNaive(t *testing.T) {
	for trial := 0; trial < 1000; trial++ {
		rng := rand.New(rand.NewSource(int64(1000 + trial)))
		tree := New()
		naive := newNaive()
		pool := make([]string, len(compareKeyPool))
		copy(pool, compareKeyPool)
		rng.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
		pool = pool[:8+rng.Intn(6)]

		steps := 30 + rng.Intn(40)
		for step := 0; step < steps; step++ {
			key := pool[rng.Intn(len(pool))]
			switch rng.Intn(10) {
			case 0, 1, 2, 3: // 40% put
				value := rng.Intn(1000)
				tree.Put(key, value)
				naive.put(key, value)
			case 4, 5: // 20% delete
				if tree.Delete(key) != (naive.delete(key)) {
					t.Fatalf("第 %d 组第 %d 步 Delete(%q) 返回值不一致", trial, step, key)
				}
			case 6, 7: // 20% get
				rv, rok := tree.Get(key)
				nv, nok := naive.get(key)
				if rv != nv || rok != nok {
					t.Fatalf("第 %d 组第 %d 步 Get(%q) 不一致：%d/%v vs %d/%v",
						trial, step, key, rv, rok, nv, nok)
				}
			case 8: // 10% prefix
				rt := tree.WithPrefix(key)
				nt := naive.withPrefix(key)
				if len(rt) != len(nt) {
					t.Fatalf("第 %d 组第 %d 步 WithPrefix(%q) 长度不一致", trial, step, key)
				}
				for i := range rt {
					if rt[i] != nt[i] {
						t.Fatalf("第 %d 组第 %d 步 WithPrefix(%q) 不一致：%v vs %v",
							trial, step, key, rt, nt)
					}
				}
			default: // 10% longest
				rl, rok := tree.LongestPrefix(key)
				nl, nok := naive.longestPrefix(key)
				if rl != nl || rok != nok {
					t.Fatalf("第 %d 组第 %d 步 LongestPrefix(%q) 不一致：%q/%v vs %q/%v",
						trial, step, key, rl, rok, nl, nok)
				}
			}
		}
		assertSameState(t, tree, naive, trial)
	}
}

// TestSampleScriptAgainstNaive 把 sample/script.txt 的操作序列也对照一遍。
func TestSampleScriptAgainstNaive(t *testing.T) {
	tree := New()
	naive := newNaive()
	apply := func(kind int, key string, value int) {
		switch kind {
		case 0:
			tree.Put(key, value)
			naive.put(key, value)
		case 1:
			if tree.Delete(key) != naive.delete(key) {
				t.Fatal("sample 序列 Delete 不一致")
			}
		}
	}
	apply(0, "a", 5)
	apply(0, "apples", 6)
	apply(0, "apple", 2)
	apply(0, "app", 3)
	apply(0, "banana", 1)
	apply(0, "banana", 9)

	if key, ok := tree.LongestPrefix("applesauce"); !ok || key != "apples" {
		t.Fatalf("sample: longest applesauce 不对：%q %v", key, ok)
	}
	if key, ok := naive.longestPrefix("applesauce"); !ok || key != "apples" {
		t.Fatalf("naive: longest applesauce 不对：%q %v", key, ok)
	}
	rt := tree.WithPrefix("app")
	nt := naive.withPrefix("app")
	want := []Entry{{Key: "app", Value: 3}, {Key: "apple", Value: 2}, {Key: "apples", Value: 6}}
	for i := range want {
		if rt[i] != want[i] || nt[i] != want[i] {
			t.Fatalf("sample: prefix app 不对：%v / %v", rt, nt)
		}
	}

	apply(1, "apple", 0)
	if got := tree.Keys(); len(got) != 4 || got[0] != "a" || got[3] != "banana" {
		t.Fatalf("sample: 删 apple 后 Keys 不对：%v", got)
	}
	assertSameState(t, tree, naive, 0)
	if tree.Len() != 5-1 {
		t.Fatalf("sample: len 应为 4，实际 %d", tree.Len())
	}
}
