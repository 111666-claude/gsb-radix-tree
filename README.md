# radix-tree

字符串前缀索引：写入、读取、删除、前缀枚举、最长前缀匹配，只用 Go 标准库。

```
go test ./...
go run ./cmd/radix-tree --script sample/script.txt
go run ./cmd/radix-tree --generate 5000
```

## 口径（README 为准）

- **重复键是更新**：`put banana 1` 之后再 `put banana 9`，只留一条、值取最后一次写入。
  `sample/script.txt` 里两条 `banana` 合并为一条，最终 `len=4`（删 `apple` 之后）。
- **删除要真的删掉**：`Delete` 之后 `Len()` 减一、`Keys()` 里不再出现、`Get` 读不到；
  删除不存在的键返回 `false`。删空后自动剪枝合并，树仍是压缩形态。
- **`WithPrefix` 按字典序**：输出与「先按前缀过滤、再按字典序排序」逐条一致。
  `prefix "app"` 给 `app=3 apple=2 apples=6`。
- **`LongestPrefix` 返回最长匹配**：沿查询串下行，记录沿途命中的终端键，取最长者；
  `longest applesauce` 给 `apples`；没有任何键是查询串前缀时 `ok=false`。
- **复杂度**：内部是压缩基数树（compressed trie / radix tree）。读取/写入/删除 O(键长)；
  前缀枚举只走前缀对应子树，复杂度约为 O(前缀长 + 命中条目数)；最长前缀匹配 O(查询串长)，
  均不线性扫描全部键。`--generate 5000`（压 5000 个键再查 5000 次）的 `compared`
  从 12507500 降到 42500（`compared` 计下行经过的边数），在键长量级。
- 不变量：同一操作序列重复执行结果相同；`Len()` 等于 `Keys()` 的长度；删除后再 `Get` 必失败。
  空键 `""` 与空树上的所有操作也正常工作，不会 panic。

## 实现

- `Tree` 内部是压缩基数树：每条边挂一段公共字符串，节点用 `map[byte]` 选下一条边，
  终端节点存值；`Put` 遇到分叉在中间分裂节点，`Delete` 后自底向上剪枝并把
  「非终端且只有一个孩子」的节点与孩子合并。
- `WithPrefix` 走到前缀位置后按首字节有序深度优先遍历，天然按字典序产出。
- `Compared()` 统计下行过程中经过的边数（对应旧线性实现里“与多少条记录比较过”），
  用于验收不发生全量扫描。
- `naive_test.go` 里另有一份「切片 + 排序 + 线性扫描」的朴素参考实现，
  对样例与 1000 组随机操作序列逐条比对 `Get`、`Keys`、`WithPrefix`、`LongestPrefix`。

## 计划新增的能力

```
--generate N    压 N 个键再查 N 次，打印 len 与 compared
--script 里的 longest/keys
```

## 输出

```
get <键> <值> <ok>
longest <查询串> <匹配键> <ok>
prefix "<前缀>" 键=值 …
keys k1,k2,…
len=… compared=…
```

## 目录

```
tree.go              前缀索引与操作
cmd/radix-tree       命令行入口
tree_test.go         go test 用例
sample/script.txt    含重复键、前缀枚举、最长前缀与删除
```
