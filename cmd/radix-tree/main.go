// Command radix-tree 按脚本驱动前缀索引。
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"example.com/radixtree"
)

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("radix-tree", flag.ContinueOnError)
	flags.SetOutput(stderr)
	script := flags.String("script", "", "脚本：put k v / get k / del k / prefix p / longest k / stats")
	generate := flags.Int("generate", 0, "压入 N 个键并做 N 次查询，打印扫描次数")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	tree := radixtree.New()
	if *generate > 0 {
		for index := 0; index < *generate; index++ {
			tree.Put(fmt.Sprintf("key%06d", index), index)
		}
		for index := 0; index < *generate; index++ {
			tree.Get(fmt.Sprintf("key%06d", index))
		}
		fmt.Fprintf(stdout, "len=%d compared=%d\n", tree.Len(), tree.Compared())
		return 0
	}
	if *script == "" {
		fmt.Fprintln(stderr, "用法：radix-tree --script sample/script.txt | --generate 5000")
		return 2
	}
	raw, err := os.ReadFile(*script)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	scanner := bufio.NewScanner(strings.NewReader(string(raw)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		switch fields[0] {
		case "put":
			value, _ := strconv.Atoi(fields[2])
			tree.Put(fields[1], value)
		case "get":
			value, ok := tree.Get(fields[1])
			fmt.Fprintf(stdout, "get %s %d %v\n", fields[1], value, ok)
		case "del":
			fmt.Fprintf(stdout, "del %s %v\n", fields[1], tree.Delete(fields[1]))
		case "prefix":
			out := tree.WithPrefix(fields[1])
			fmt.Fprintf(stdout, "prefix %q", fields[1])
			for _, entry := range out {
				fmt.Fprintf(stdout, " %s=%d", entry.Key, entry.Value)
			}
			fmt.Fprintln(stdout)
		case "longest":
			key, ok := tree.LongestPrefix(fields[1])
			fmt.Fprintf(stdout, "longest %s %s %v\n", fields[1], key, ok)
		case "keys":
			fmt.Fprintf(stdout, "keys %s\n", strings.Join(tree.Keys(), ","))
		case "stats":
			fmt.Fprintf(stdout, "len=%d compared=%d\n", tree.Len(), tree.Compared())
		default:
			fmt.Fprintf(stderr, "不认识的命令：%s\n", fields[0])
			return 2
		}
	}
	return 0
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
