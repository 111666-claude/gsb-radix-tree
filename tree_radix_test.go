package radixtree

import (
	"math/rand"
	"sort"
	"testing"
)

func TestDuplicatePutIsUpdate(t *testing.T) {
	tree := New()
	tree.Put("banana", 1)
	tree.Put("banana", 9)
	if tree.Len() != 1 {
		t.Fatalf("重复键只该留一条，Len=%d", tree.Len())
	}
	value, ok := tree.Get("banana")
	if !ok || value != 9 {
		t.Fatalf("值应取最后一次写入：%d %v", value, ok)
	}
	keys := tree.Keys()
	if len(keys) != 1 || keys[0] != "banana" {
		t.Fatalf("Keys 里只该有一条 banana：%v", keys)
	}
}

func TestDeleteReallyRemoves(t *testing.T) {
	tree := New()
	tree.Put("apple", 2)
	if !tree.Delete("apple") {
		t.Fatal("删除存在的键应返回 true")
	}
	if tree.Len() != 0 {
		t.Fatalf("删除后 Len 应减一：%d", tree.Len())
	}
	if len(tree.Keys()) != 0 {
		t.Fatalf("删除后 Keys 不该再出现：%v", tree.Keys())
	}
	if _, ok := tree.Get("apple"); ok {
		t.Fatal("删除后 Get 必失败")
	}
	if tree.Delete("apple") {
		t.Fatal("再次删除应返回 false")
	}
	// 删完还能重新写。
	tree.Put("apple", 5)
	if value, ok := tree.Get("apple"); !ok || value != 5 {
		t.Fatalf("删除后重写不对：%d %v", value, ok)
	}
}

func TestWithPrefixIsLexicographic(t *testing.T) {
	tree := New()
	tree.Put("apples", 6)
	tree.Put("apple", 2)
	tree.Put("app", 3)
	tree.Put("a", 5)
	tree.Put("banana", 1)
	want := []Entry{{Key: "app", Value: 3}, {Key: "apple", Value: 2}, {Key: "apples", Value: 6}}
	got := tree.WithPrefix("app")
	if len(got) != len(want) {
		t.Fatalf("WithPrefix 条数不对：%v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("WithPrefix 应先过滤再按字典序：%v", got)
		}
	}
}

func TestLongestPrefixMatch(t *testing.T) {
	tree := New()
	tree.Put("a", 1)
	tree.Put("apples", 6)
	tree.Put("apple", 2)
	tree.Put("app", 3)

	key, ok := tree.LongestPrefix("applesauce")
	if !ok || key != "apples" {
		t.Fatalf("最长匹配应是 apples：%q %v", key, ok)
	}
	key, ok = tree.LongestPrefix("app")
	if !ok || key != "app" {
		t.Fatalf("查询串自身是键时应返回自身：%q %v", key, ok)
	}
	if _, ok := tree.LongestPrefix("banana"); ok {
		t.Fatal("没有匹配时 ok 应为 false")
	}
	if _, ok := tree.LongestPrefix(""); ok {
		t.Fatal("空查询且无空键时 ok 应为 false")
	}

	// 空键是所有查询串的前缀。
	tree.Put("", 0)
	if key, ok := tree.LongestPrefix("zzz"); !ok || key != "" {
		t.Fatalf("空键应能命中：%q %v", key, ok)
	}
}

func TestEmptyTreeDoesNotPanic(t *testing.T) {
	tree := New()
	if tree.Len() != 0 || len(tree.Keys()) != 0 {
		t.Fatal("空树长度应为 0")
	}
	if _, ok := tree.Get("x"); ok {
		t.Fatal("空树 Get 应失败")
	}
	if tree.Delete("x") {
		t.Fatal("空树 Delete 应返回 false")
	}
	if out := tree.WithPrefix("x"); len(out) != 0 {
		t.Fatalf("空树 WithPrefix 应为空：%v", out)
	}
	if _, ok := tree.LongestPrefix("x"); ok {
		t.Fatal("空树 LongestPrefix 应失败")
	}
}

func TestRepeatedSequenceIsDeterministic(t *testing.T) {
	keys := []string{"a", "app", "apple", "apples", "banana", "band", "cat"}
	makeSequence := func(seed int64) []op {
		rng := rand.New(rand.NewSource(seed))
		ops := make([]op, 200)
		for i := range ops {
			kind := rng.Intn(5)
			ops[i] = op{kind: kind, key: keys[rng.Intn(len(keys))], value: rng.Intn(100)}
		}
		return ops
	}

	run := func(ops []op) ([]Entry, []string) {
		tree := New()
		var prefixes []Entry
		for _, operation := range ops {
			switch operation.kind {
			case 0:
				tree.Put(operation.key, operation.value)
			case 1:
				tree.Delete(operation.key)
			case 2:
				tree.Get(operation.key)
			case 3:
				prefixes = tree.WithPrefix(operation.key)
			case 4:
				tree.LongestPrefix(operation.key)
			}
		}
		return prefixes, tree.Keys()
	}

	ops := makeSequence(42)
	p1, k1 := run(ops)
	p2, k2 := run(ops)
	if len(p1) != len(p2) {
		t.Fatal("重复执行 WithPrefix 结果不一致")
	}
	for i := range p1 {
		if p1[i] != p2[i] {
			t.Fatal("重复执行 WithPrefix 结果不一致")
		}
	}
	if len(k1) != len(k2) {
		t.Fatal("重复执行 Keys 长度不一致")
	}
	for i := range k1 {
		if k1[i] != k2[i] {
			t.Fatal("重复执行 Keys 内容不一致")
		}
	}
}

func TestDeleteAllThenReuse(t *testing.T) {
	tree := New()
	keys := []string{"a", "app", "apple", "apples", "banana", "band"}
	for i, key := range keys {
		tree.Put(key, i)
	}
	for _, key := range keys {
		if !tree.Delete(key) {
			t.Fatalf("删除 %s 应成功", key)
		}
		if tree.Len() != len(tree.Keys()) {
			t.Fatal("任何时刻 Len 都应等于 Keys 长度")
		}
	}
	if tree.Len() != 0 || len(tree.Keys()) != 0 {
		t.Fatalf("删空后应为空树：len=%d keys=%v", tree.Len(), tree.Keys())
	}
	for _, key := range keys {
		if _, ok := tree.Get(key); ok {
			t.Fatalf("删空后 Get(%s) 必须失败", key)
		}
	}
	tree.Put("apple", 9)
	if got := tree.Keys(); len(got) != 1 || got[0] != "apple" {
		t.Fatalf("删空后重写不对：%v", got)
	}
}

// TestGenerateComparedIsKeyLength 复现 --generate 5000 的负载：
// 旧的线性实现 compared=12507500；走前缀结构后必须在键长量级（几万以内）。
func TestGenerateComparedIsKeyLength(t *testing.T) {
	const n = 5000
	tree := New()
	for index := 0; index < n; index++ {
		tree.Put(generateKey(index), index)
	}
	for index := 0; index < n; index++ {
		tree.Get(generateKey(index))
	}
	compared := tree.Compared()
	if tree.Len() != n {
		t.Fatalf("Len 应为 %d，实际 %d", n, tree.Len())
	}
	if compared >= 100000 {
		t.Fatalf("compared 应在键长量级（几万以内），实际 %d", compared)
	}
	// 与线性扫描的 O(n^2) 拉开数量级。
	if compared > n*n/100 {
		t.Fatalf("compared 仍像线性扫描：%d", compared)
	}
}

func generateKey(index int) string {
	const digits = "0123456789"
	key := []byte("key000000")
	for position := len(key) - 1; position >= 3; position-- {
		key[position] = digits[index%10]
		index /= 10
	}
	return string(key)
}

func TestKeysAlwaysSortedAndMatchesLen(t *testing.T) {
	tree := New()
	rng := rand.New(rand.NewSource(7))
	keys := []string{"a", "ap", "app", "apple", "apples", "banana", "band", "cat", "x", "xy", "xyz"}
	for step := 0; step < 300; step++ {
		key := keys[rng.Intn(len(keys))]
		switch rng.Intn(3) {
		case 0:
			tree.Put(key, rng.Intn(10))
		case 1:
			tree.Delete(key)
		case 2:
			tree.Get(key)
		}
		all := tree.Keys()
		if tree.Len() != len(all) {
			t.Fatalf("第 %d 步：Len=%d 但 len(Keys)=%d", step, tree.Len(), len(all))
		}
		if !sort.StringsAreSorted(all) {
			t.Fatalf("第 %d 步：Keys 未按字典序：%v", step, all)
		}
	}
}
