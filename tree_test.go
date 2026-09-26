package radixtree

import "testing"

func TestPutAndGet(t *testing.T) {
	tree := New()
	tree.Put("a", 1)
	value, ok := tree.Get("a")
	if !ok || value != 1 {
		t.Fatalf("Get 不对：%d %v", value, ok)
	}
}

func TestGetMissing(t *testing.T) {
	if _, ok := New().Get("x"); ok {
		t.Fatal("空树不该读到")
	}
}

func TestLenCountsEntries(t *testing.T) {
	tree := New()
	tree.Put("a", 1)
	tree.Put("b", 2)
	if tree.Len() != 2 {
		t.Fatalf("Len 不对：%d", tree.Len())
	}
}

func TestKeysAreSorted(t *testing.T) {
	tree := New()
	tree.Put("b", 2)
	tree.Put("a", 1)
	keys := tree.Keys()
	if len(keys) != 2 || keys[0] != "a" || keys[1] != "b" {
		t.Fatalf("Keys 应该按字典序：%v", keys)
	}
}

func TestWithPrefixFilters(t *testing.T) {
	tree := New()
	tree.Put("app", 1)
	tree.Put("banana", 2)
	out := tree.WithPrefix("app")
	if len(out) != 1 || out[0].Key != "app" {
		t.Fatalf("前缀过滤不对：%v", out)
	}
}

func TestDeleteReportsFound(t *testing.T) {
	tree := New()
	tree.Put("a", 1)
	if !tree.Delete("a") {
		t.Fatal("删除存在的键应该返回 true")
	}
	if tree.Delete("ghost") {
		t.Fatal("删除不存在的键应该返回 false")
	}
}

func TestComparedIsCounted(t *testing.T) {
	tree := New()
	tree.Put("a", 1)
	tree.Get("a")
	if tree.Compared() == 0 {
		t.Fatal("扫描次数应该被统计")
	}
}
