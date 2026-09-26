package radixtree

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func TestPutDuplicateUpdates(t *testing.T) {
	tree := New()
	tree.Put("banana", 1)
	tree.Put("banana", 9)
	if tree.Len() != 1 {
		t.Fatalf("重复键只该留一条，Len=%d", tree.Len())
	}
	value, ok := tree.Get("banana")
	if !ok || value != 9 {
		t.Fatalf("重复键应取最后一次写入：%d %v", value, ok)
	}
	if keys := tree.Keys(); len(keys) != 1 || keys[0] != "banana" {
		t.Fatalf("Keys 里只能有一条 banana：%v", keys)
	}
}

func TestDeleteActuallyRemoves(t *testing.T) {
	tree := New()
	tree.Put("apple", 2)
	tree.Put("apples", 6)
	if !tree.Delete("apple") {
		t.Fatal("删除存在的键应返回 true")
	}
	if tree.Len() != 1 {
		t.Fatalf("删除后 Len 应减一：%d", tree.Len())
	}
	if _, ok := tree.Get("apple"); ok {
		t.Fatal("删除后 Get 必须失败")
	}
	for _, key := range tree.Keys() {
		if key == "apple" {
			t.Fatal("删除后 Keys 里不该再出现 apple")
		}
	}
	if tree.Delete("apple") {
		t.Fatal("再次删除应返回 false")
	}
}

func TestWithPrefixIsSorted(t *testing.T) {
	tree := New()
	tree.Put("apples", 6)
	tree.Put("apple", 2)
	tree.Put("app", 3)
	got := tree.WithPrefix("app")
	want := []Entry{
		{Key: "app", Value: 3},
		{Key: "apple", Value: 2},
		{Key: "apples", Value: 6},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("WithPrefix 应按字典序：%v", got)
	}
}

func TestLongestPrefixMatch(t *testing.T) {
	tree := New()
	tree.Put("a", 5)
	tree.Put("apples", 6)
	tree.Put("apple", 2)
	tree.Put("app", 3)
	key, ok := tree.LongestPrefix("applesauce")
	if !ok || key != "apples" {
		t.Fatalf("longest applesauce 应为 apples：%q %v", key, ok)
	}
	if _, ok := tree.LongestPrefix("xyz"); ok {
		t.Fatal("没有任何匹配时 ok 必须为 false")
	}
	exact, ok := tree.LongestPrefix("apple")
	if !ok || exact != "apple" {
		t.Fatalf("精确命中应返回自身：%q %v", exact, ok)
	}
}

func TestEmptyTreeNoPanic(t *testing.T) {
	tree := New()
	if tree.Len() != 0 {
		t.Fatal("空树 Len 应为 0")
	}
	if len(tree.Keys()) != 0 {
		t.Fatal("空树 Keys 应为空")
	}
	if len(tree.WithPrefix("")) != 0 || len(tree.WithPrefix("x")) != 0 {
		t.Fatal("空树前缀枚举应为空")
	}
	if _, ok := tree.Get("x"); ok {
		t.Fatal("空树 Get 应失败")
	}
	if tree.Delete("x") {
		t.Fatal("空树 Delete 应返回 false")
	}
	if _, ok := tree.LongestPrefix("x"); ok {
		t.Fatal("空树 LongestPrefix 应失败")
	}
}

func TestRepeatedSequenceDeterministic(t *testing.T) {
	ops := func(tree *Tree) []string {
		tree.Put("a", 5)
		tree.Put("apples", 6)
		tree.Put("apple", 2)
		tree.Put("app", 3)
		tree.Put("banana", 1)
		tree.Put("banana", 9)
		tree.Delete("apple")
		return tree.Keys()
	}
	first := ops(New())
	second := ops(New())
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("同一操作序列结果应相同：%v vs %v", first, second)
	}
	want := []string{"a", "app", "apples", "banana"}
	if !reflect.DeepEqual(first, want) {
		t.Fatalf("样例序列终态不对：%v", first)
	}
}

func TestLenEqualsKeysLength(t *testing.T) {
	tree := New()
	for _, key := range []string{"a", "ab", "abc", "b", "ab"} {
		tree.Put(key, 1)
	}
	tree.Delete("ab")
	if tree.Len() != len(tree.Keys()) {
		t.Fatalf("Len 必须等于 Keys 长度：%d vs %d", tree.Len(), len(tree.Keys()))
	}
}

func TestEmptyKeySupported(t *testing.T) {
	tree := New()
	tree.Put("", 7)
	if value, ok := tree.Get(""); !ok || value != 7 {
		t.Fatalf("空键读写不对：%d %v", value, ok)
	}
	if tree.Len() != 1 {
		t.Fatalf("空键也应计数：%d", tree.Len())
	}
	longest, ok := tree.LongestPrefix("anything")
	if !ok || longest != "" {
		t.Fatalf("空键是任意串的前缀：%q %v", longest, ok)
	}
	tree.Delete("")
	if _, ok := tree.Get(""); ok {
		t.Fatal("删除空键后应读不到")
	}
}

// TestComparedIsSublinear 5000 键压入 + 5000 次查询，
// compared 应在键长量级（几万），线性扫描会到千万级。
func TestComparedIsSublinear(t *testing.T) {
	tree := New()
	const n = 5000
	for i := 0; i < n; i++ {
		tree.Put(keyN(i), i)
	}
	if tree.Compared() > n*20 {
		t.Fatalf("写入 compared=%d 超过键长量级", tree.Compared())
	}
	writeCompared := tree.Compared()
	for i := 0; i < n; i++ {
		if value, ok := tree.Get(keyN(i)); !ok || value != i {
			t.Fatalf("回读不对：%d %v", value, ok)
		}
	}
	readCompared := tree.Compared() - writeCompared
	if readCompared > n*20 {
		t.Fatalf("读取 compared=%d 超过键长量级", readCompared)
	}
	t.Logf("5000 键：写入 compared=%d，读取 compared=%d", writeCompared, readCompared)
}

func keyN(i int) string {
	// 与命令行 --generate 使用同一套键。
	return fmt.Sprintf("key%06d", i)
}

// differentialKeys 是差分测试用的键池，刻意包含大量前缀关系与空键。
var differentialSeeds = []string{
	"", "a", "ap", "app", "apple", "apples", "applesauce", "applet",
	"b", "ban", "band", "banana", "bananas",
	"key", "key0", "key00", "key000001", "key000002", "x", "y", "z",
}

func randomKey(rng *rand.Rand) string {
	switch rng.Intn(3) {
	case 0:
		return differentialSeeds[rng.Intn(len(differentialSeeds))]
	case 1:
		length := rng.Intn(5)
		letters := []byte("abpelnkxy012")
		b := make([]byte, length)
		for i := range b {
			b[i] = letters[rng.Intn(len(letters))]
		}
		return string(b)
	default:
		prefixes := []string{"app", "ban", "key", "a", ""}
		letters := []byte("abpelnkxy")
		b := []byte(prefixes[rng.Intn(len(prefixes))])
		for i := 0; i < rng.Intn(4); i++ {
			b = append(b, letters[rng.Intn(len(letters))])
		}
		return string(b)
	}
}

func entriesEqual(a, b []Entry) bool { return reflect.DeepEqual(a, b) }

// TestDifferentialAgainstNaive 对 1000 组随机操作序列，
// 逐条比对 radix 树与朴素切片实现的 Get/Delete/WithPrefix/LongestPrefix 结果。
func TestDifferentialAgainstNaive(t *testing.T) {
	rng := rand.New(rand.NewSource(20260927))
	const sequences, ops = 1000, 200
	for seq := 0; seq < sequences; seq++ {
		rt := New()
		nt := newNaive()
		keys := make([]string, 0, 24)
		for i := 0; i < 20; i++ {
			keys = append(keys, randomKey(rng))
		}
		prefixes := append([]string{""}, keys...)
		for step := 0; step < ops; step++ {
			key := keys[rng.Intn(len(keys))]
			switch rng.Intn(7) {
			case 0, 1, 2:
				value := rng.Intn(1000) - 500
				rt.Put(key, value)
				nt.Put(key, value)
			case 3:
				if rt.Delete(key) != nt.Delete(key) {
					t.Fatalf("seq=%d step=%d Delete(%q) 不一致", seq, step, key)
				}
			case 4:
				rv, rok := rt.Get(key)
				nv, nok := nt.Get(key)
				if rv != nv || rok != nok {
					t.Fatalf("seq=%d step=%d Get(%q) 不一致：%d/%v vs %d/%v",
						seq, step, key, rv, rok, nv, nok)
				}
			case 5:
				prefix := prefixes[rng.Intn(len(prefixes))]
				ro := rt.WithPrefix(prefix)
				no := nt.WithPrefix(prefix)
				if !entriesEqual(ro, no) {
					t.Fatalf("seq=%d step=%d WithPrefix(%q) 不一致：%v vs %v",
						seq, step, prefix, ro, no)
				}
			default:
				query := keys[rng.Intn(len(keys))] + rngStringSuffix(rng)
				rl, rok := rt.LongestPrefix(query)
				nl, nok := nt.LongestPrefix(query)
				if rl != nl || rok != nok {
					t.Fatalf("seq=%d step=%d LongestPrefix(%q) 不一致：%q/%v vs %q/%v",
						seq, step, query, rl, rok, nl, nok)
				}
			}
			// 状态不变量：Len、Keys 与全部键的 Get。
			if rt.Len() != nt.Len() {
				t.Fatalf("seq=%d step=%d Len 不一致：%d vs %d", seq, step, rt.Len(), nt.Len())
			}
			if rkeys, nkeys := rt.Keys(), nt.Keys(); !reflect.DeepEqual(rkeys, nkeys) {
				t.Fatalf("seq=%d step=%d Keys 不一致：%v vs %v", seq, step, rkeys, nkeys)
			}
			if len(rt.Keys()) != rt.Len() {
				t.Fatalf("seq=%d step=%d Len 不等于 Keys 长度", seq, step)
			}
			for _, probe := range keys {
				rv, rok := rt.Get(probe)
				nv, nok := nt.Get(probe)
				if rv != nv || rok != nok {
					t.Fatalf("seq=%d step=%d 全量探针 Get(%q) 不一致", seq, step, probe)
				}
			}
		}
	}
}

func rngStringSuffix(rng *rand.Rand) string {
	switch rng.Intn(3) {
	case 0:
		return ""
	case 1:
		return "sauce"
	default:
		letters := []byte("abpelnkxy")
		b := make([]byte, rng.Intn(4))
		for i := range b {
			b[i] = letters[rng.Intn(len(letters))]
		}
		return string(b)
	}
}
