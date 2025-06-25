# GORM Auto Explain 插件

[English README](./README.md)

一个 GORM 插件，根据查询时长和阈值自动执行 `EXPLAIN`，通过回调函数返回详细的查询性能分析结果。

## 功能

- 自动对超过阈值的查询执行 `EXPLAIN`。
- 提供详细查询分析：
  - 查询执行时间
  - 查询成本评分
  - 是否使用临时表
  - 原始 `EXPLAIN` 输出
  - 原始 SQL 查询
- 支持自定义回调函数处理 `EXPLAIN` 结果。
- 可在迁移或特定操作时开关插件。
- 轻量级，易于集成到现有 GORM 项目。

## 安装

```bash
go get github.com/etng/gorm_auto_explain
```

## 使用

### 初始化

1. 导入插件。
2. 初始化 GORM 和插件，并设置回调函数。
3. 可在迁移时禁用插件。

```go
package main

import (
    "log"
    "time"
    autoExplain "github.com/etng/gorm_auto_explain"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

type User struct {
    ID        uint
    Name      string
    JoinedAt  time.Time
}

func main() {
    db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
    if err != nil {
        log.Fatal("数据库连接失败:", err)
    }

    // 迁移时禁用插件
    autoExplain.Toggle(false)
    if err := db.AutoMigrate(&User{}); err != nil {
        log.Fatal("迁移失败:", err)
    }
    autoExplain.Toggle(true)

    // 初始化插件并设置回调
    db.Use(autoExplain.InitPlugin(0).OnExplain(func(result autoExplain.Result) {
        if result.Message != "" {
            log.Println("提示:", result.Message)
        }
        if result.QueryCost > 100 {
            log.Printf("发现高成本 SQL，建议优化，评分=%.2f，耗时=%s", result.QueryCost, result.Duration)
        }
        log.Printf("分析详情: 查询: %s\n成本: %.2f\n耗时: %s\n%s", result.Query, result.QueryCost, result.Duration, result.Raw)
    }))

    // 示例查询
    var users []User
    db.Where("name = ?", "test").Find(&users)
}
```

### 结果结构

`Result` 结构体包含以下字段：

```go
type Result struct {
    Duration            time.Duration `json:"duration"`            // 查询执行时间
    Message             string        `json:"message"`             // 提示或警告
    QueryCost           float64       `json:"query_cost"`          // 查询成本评分
    UsingTemporaryTable bool          `json:"using_temporary_table"` // 是否使用临时表
    Raw                 string        `json:"raw"`                 // 原始 EXPLAIN 输出
    Query               string        `json:"query"`               // 原始 SQL 查询
}
```

### 回调函数

`OnExplainCb` 回调用于自定义处理 `EXPLAIN` 结果：

```go
type OnExplainCb func(result Result)
```

### 示例

查看 `example/main.go` 获取完整示例。

```go
// 来自 example/main.go 的摘录
db.Use(autoExplain.InitPlugin(0).OnExplain(func(result autoExplain.Result) {
    log.Printf("分析详情: 查询: %s\n成本: %.2f\n耗时: %s\n%s", result.Query, result.QueryCost, result.Duration, result.Raw)
}))
```

## 配置

- **阈值**：设置触发 `EXPLAIN` 的查询时长阈值（纳秒）。设为 `0` 表示所有查询。
  ```go
  autoExplain.InitPlugin(1000000) // 1ms 阈值
  ```
- **开关**：动态启用或禁用插件。
  ```go
  autoExplain.Toggle(false) // 禁用
  autoExplain.Toggle(true)  // 启用
  ```

## 注意事项

- 确保数据库支持 `EXPLAIN`（如 MySQL、PostgreSQL）。
- 迁移时默认禁用插件以避免不必要的分析。
- 高查询成本（>100）或使用临时表可能提示优化空间。

## 贡献

欢迎贡献！请在 [GitHub](https://github.com/etng/gorm_auto_explain) 提交问题或拉取请求。

## 许可证

MIT 许可证。详见 [LICENSE](LICENSE)。