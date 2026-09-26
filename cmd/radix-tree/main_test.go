package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func runCLI(t *testing.T, args ...string) string {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if code := Run(args, &stdout, &stderr); code != 0 {
		t.Fatalf("退出码 %d：%s", code, stderr.String())
	}
	return stdout.String()
}

// TestScriptGolden 锁定样例脚本的输出口径：最长前缀、字典序枚举、
// 重复键合并与真删除。
func TestScriptGolden(t *testing.T) {
	script := filepath.Join("..", "..", "sample", "script.txt")
	if _, err := os.Stat(script); err != nil {
		t.Skip("找不到样例脚本")
	}
	out := runCLI(t, "--script", script)
	wantLines := []string{
		`longest applesauce apples true`,
		`prefix "app" app=3 apple=2 apples=6`,
		`del apple true`,
		`keys a,app,apples,banana`,
		`len=4 compared=`,
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != len(wantLines) {
		t.Fatalf("输出行数不对：%q", out)
	}
	for i, want := range wantLines {
		if !strings.HasPrefix(lines[i], want) {
			t.Fatalf("第 %d 行应前缀 %q，实际 %q", i+1, want, lines[i])
		}
	}
}

// TestGenerateSublinear --generate 5000 的 compared 必须在键长量级
// （线性扫描旧实现是 12507500）。
func TestGenerateSublinear(t *testing.T) {
	out := runCLI(t, "--generate", "5000")
	fields := strings.Fields(strings.TrimSpace(out))
	var compared int
	for _, field := range fields {
		if strings.HasPrefix(field, "compared=") {
			n, err := strconv.Atoi(strings.TrimPrefix(field, "compared="))
			if err != nil {
				t.Fatal(err)
			}
			compared = n
		}
	}
	if compared == 0 {
		t.Fatalf("没读到 compared：%q", out)
	}
	if compared >= 5000*20 {
		t.Fatalf("compared=%d 超出键长量级", compared)
	}
}

func TestNoArgsPrintsUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run(nil, &stdout, &stderr); code != 2 {
		t.Fatalf("无参数退出码应为 2，实际 %d", code)
	}
}
