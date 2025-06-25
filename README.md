# GORM Auto Explain Plugin

[中文版介绍](./README_ZH.md)

A GORM plugin that automatically executes `EXPLAIN` on SQL queries based on their duration and a configurable threshold, providing detailed query performance insights via a callback function.

## Features

- Automatically runs `EXPLAIN` on queries exceeding a duration threshold.
- Provides detailed query analysis, including:
  - Query duration
  - Query cost score
  - Use of temporary tables
  - Raw `EXPLAIN` output
  - Original query
- Customizable callback function to handle `EXPLAIN` results.
- Toggleable to enable/disable during migrations or specific operations.
- Lightweight and easy to integrate with existing GORM projects.

## Installation

```bash
go get github.com/etng/gorm_auto_explain
```

## Usage

### Initialization

1. Import the plugin.
2. Initialize GORM and the plugin with a callback function.
3. Optionally toggle the plugin to disable during migrations.

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
        log.Fatal("failed to connect database:", err)
    }

    // Disable plugin during migration
    autoExplain.Toggle(false)
    if err := db.AutoMigrate(&User{}); err != nil {
        log.Fatal("migration failed:", err)
    }
    autoExplain.Toggle(true)

    // Initialize plugin with callback
    db.Use(autoExplain.InitPlugin(0).OnExplain(func(result autoExplain.Result) {
        if result.Message != "" {
            log.Println("notice:", result.Message)
        }
        if result.QueryCost > 100 {
            log.Printf("heavy sql you should try to optimize found, score=%.2f, duration=%s", result.QueryCost, result.Duration)
        }
        log.Printf("explain detail: query: %s\ncost: %.2f\n duration:%s\n%s", result.Query, result.QueryCost, result.Duration, result.Raw)
    }))

    // Example query
    var users []User
    db.Where("name = ?", "test").Find(&users)
}
```

### Result Structure

The `Result` struct contains the following fields:

```go
type Result struct {
    Duration            time.Duration `json:"duration"`            // Query execution time
    Message             string        `json:"message"`             // Notices or warnings
    QueryCost           float64       `json:"query_cost"`          // Query cost score
    UsingTemporaryTable bool          `json:"using_temporary_table"` // Indicates temporary table usage
    Raw                 string        `json:"raw"`                 // Raw EXPLAIN output
    Query               string        `json:"query"`               // Original SQL query
}
```

### Callback Function

The `OnExplainCb` callback allows custom handling of `EXPLAIN` results:

```go
type OnExplainCb func(result Result)
```

### Example

See `example/main.go` for a complete example demonstrating plugin setup and usage.

```go
// Excerpt from example/main.go
db.Use(autoExplain.InitPlugin(0).OnExplain(func(result autoExplain.Result) {
    log.Printf("explain detail: query: %s\ncost: %.2f\n duration:%s\n%s", result.Query, result.QueryCost, result.Duration, result.Raw)
}))
```

## Configuration

- **Threshold**: Set the duration threshold (in nanoseconds) for triggering `EXPLAIN`. Use `0` for all queries.
  ```go
  autoExplain.InitPlugin(1000000) // 1ms threshold
  ```
- **Toggle**: Enable or disable the plugin dynamically.
  ```go
  autoExplain.Toggle(false) // Disable
  autoExplain.Toggle(true)  // Enable
  ```

## Notes

- Ensure your database supports `EXPLAIN` (e.g., MySQL, PostgreSQL).
- The plugin is disabled by default during migrations to avoid unnecessary analysis.
- High query costs (>100) or temporary table usage may indicate optimization opportunities.

## Contributing

Contributions are welcome! Please submit issues or pull requests on [GitHub](https://github.com/etng/gorm_auto_explain).

## License

MIT License. See [LICENSE](LICENSE) for details.