# radix-tree

字符串前缀索引：写入、读取、删除、前缀枚举、最长前缀匹配，只用 Go 标准库。

```
go test ./...
go run ./cmd/radix-tree --script sample/script.txt
go run ./cmd/radix-tree --generate 5000
```

## 口径（README 为准）

- **重复键是更新**：`put banana 1` 之后再 `put banana 9`，只留一条、值取最后一次写入。
  `sample/script.txt` 现在留下两条 `banana` 且 `len=6`，按口径应该是 `len=5`。
- **删除要真的删掉**：`Delete` 之后 `Len()` 减一、`Keys()` 里不再出现、`Get` 读不到。
  现在 `del apple` 只是把值清零，`keys` 里还有 `apple`。
- **`WithPrefix` 按字典序**：输出要与「先按前缀过滤、再按字典序排序」逐条一致。
  现在返回的是插入顺序：`prefix "app"` 给 `apples=6 apple=2 app=3`，按口径应该是 `app=3 apple=2 apples=6`。
- **`LongestPrefix` 返回最长匹配**：`longest applesauce` 现在给 `a`（插入顺序里第一个匹配的），
  按口径必须是 `apples`。
- **复杂度**：读取/写入/删除是 O(键长)，前缀枚举与最长前缀匹配不许线性扫描全部键。
  `--generate 5000`（压 5000 个键再查 5000 次）现在 `compared=12507500`，按口径应该在
  键长量级（几万次以内）。
- 不变量：同一操作序列重复执行结果相同；`Len()` 等于 `Keys()` 的长度；删除后再 `Get` 必失败。

## 现在的行为

- `Tree` 内部就是一根 `[]Entry` 切片：`Get`/`Delete`/`WithPrefix`/`LongestPrefix` 全是线性扫描，
  `compared` 就是扫描次数。
- `Put` 不查重，重复键会插成两条，`Get` 返回的是**先插入**的那条（旧值）。
- `Delete` 只把 `Value` 置 0，长度不变、键还在。
- `WithPrefix` 不排序（按插入顺序），`LongestPrefix` 返回第一个匹配的键而不是最长的。

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
