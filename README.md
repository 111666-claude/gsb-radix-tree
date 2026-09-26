# radix-tree

字符串前缀索引：写入、读取、删除、前缀枚举、最长前缀匹配，只用 Go 标准库。

```
go test ./...
go vet ./...
go run ./cmd/radix-tree --script sample/script.txt
go run ./cmd/radix-tree --generate 5000
```

## 口径（README 为准）

- **重复键是更新**：`put banana 1` 之后再 `put banana 9`，只留一条、值取最后一次写入。
  `sample/script.txt` 的终态 `len=4`（`a/app/apples/banana`）。
- **删除要真的删掉**：`Delete` 之后 `Len()` 减一、`Keys()` 里不再出现、`Get` 读不到；
  树会在删除后压缩（无值且只有一个孩子的节点与孩子合并），始终保持 radix 树形态。
- **`WithPrefix` 按字典序**：输出与「先按前缀过滤、再按字典序排序」逐条一致。
  `prefix "app"` 输出 `app=3 apple=2 apples=6`。
- **`LongestPrefix` 返回最长匹配**：`longest applesauce` 返回 `apples`；
  树中没有任何键是查询串前缀时 `ok=false`。
- **复杂度**：`Get`/`Put`/`Delete` 为 O(键长)（子节点按首字节二分，不做全键扫描）；
  `WithPrefix` 为 O(前缀长度 + 命中条目数)；`LongestPrefix` 为 O(查询串长度)；
  `Keys` 为 O(条目数)。
- **不变量**：同一操作序列重复执行结果相同；`Len()` 恒等于 `Keys()` 的长度；
  删除后再 `Get` 必失败；空树上的所有操作都不 panic；支持空键 `""`。

## 实现

- `Tree` 是一棵压缩前缀树（radix tree）：每个节点保存入边上的压缩片段 `path`，
  孩子按首字节排序、二分定位；节点上的 `hasValue` 区分「键结束于此」与「仅是内部节点」。
- 插入在公共前缀结束处分裂节点；删除后对无值且只剩一个孩子的节点做合并压缩。
- `compared` 统计沿前缀路径逐字节比较的次数（`Keys`/`WithPrefix` 的 DFS 输出、
  以及子节点二分的下标探测不计入）。`--generate 5000`
  （压 5000 个 9 字节键再查 5000 次）的 `compared` 在 9 万上下，属键长量级；
  旧的线性切片实现为 `12507500`。

## 正确性对照

`naive_test.go` 里另有一个「切片 + 排序 + 线性扫描」的朴素参考实现，
`tree_extra_test.go` 用固定随机种子生成 1000 组操作序列，在每个操作之后逐条比对
两种实现的 `Get`、`Delete`、`Keys`、`WithPrefix` 与 `LongestPrefix`，
并检查 `Len == len(Keys)` 等不变量。另覆盖重复键更新、真删除、字典序枚举、
最长前缀、空键、空树与 `compared` 量级。

## 输出

```
get <键> <值> <ok>
del <键> <ok>
longest <查询串> <匹配键> <ok>
prefix "<前缀>" 键=值 …
keys k1,k2,…
len=… compared=…
```

不带新参数时输出格式与原先一致（仅修正了上面的行为）。

## 目录

```
tree.go                压缩 radix 树实现
naive_test.go          朴素参考实现（切片 + 排序 + 线性扫描）
tree_extra_test.go     差分、不变量与 compared 测试
tree_test.go           基础单测
cmd/radix-tree         命令行入口（--script / --generate）
sample/script.txt      含重复键、前缀枚举、最长前缀与删除
```
