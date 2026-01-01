# Genesis Research Pipeline

[English](#english) | [中文](#中文)

---

## English

A data pipeline for collecting and storing scientific papers from ArXiv, built with Go.

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Genesis Pipeline                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌──────────┐      ┌──────────┐      ┌──────────────────────┐  │
│   │  ArXiv   │      │  Parser  │      │      Storage         │  │
│   │   API    │─────>│  Layer   │─────>│   (PostgreSQL)       │  │
│   │          │ XML  │          │ Go   │                      │  │
│   └──────────┘      └──────────┘      └──────────────────────┘  │
│                           │                      │               │
│                           v                      v               │
│                    ┌──────────┐          ┌──────────┐           │
│                    │Validation│          │ REST API │           │
│                    │  Layer   │          │ (Query)  │           │
│                    └──────────┘          └──────────┘           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Features

- **Data Ingestion**: Fetch papers from ArXiv API with configurable queries
- **Quality Filtering**: Two-level filtering system for high-quality papers
  - Level 1: Hard gate (acceptance signals, DOI, strong evidence)
  - Level 2: Scoring (0-100) based on evaluation keywords, code links, etc.
- **Time Filtering**: Filter papers by recency (configurable max age in days)
- **Incremental Updates**: Track sync history with new/updated paper counts
- **Data Validation**: Validate paper metadata quality
- **PostgreSQL Storage**: Persist papers with search support
- **REST API**: Query stored papers via HTTP endpoints
- **Benchmarking**: Performance metrics and data quality reports

### Quick Start

```bash
# Clone the repository
git clone https://github.com/1psychoQAQ/genesis-pipeline.git
cd genesis-pipeline

# Start PostgreSQL
docker-compose -f deployments/docker-compose.yml up -d

# Run the pipeline (with quality + time filtering)
go run cmd/pipeline/main.go -query "machine learning" -limit 100 -max-age 90 -min-score 60

# Run the API server
go run cmd/api/main.go -port 8088

# Run benchmarks
go run cmd/benchmark/main.go -limit 100
```

### Pipeline Options

| Flag | Default | Description |
|------|---------|-------------|
| `-query` | "machine learning" | Search query for ArXiv |
| `-limit` | 10 | Number of papers to fetch |
| `-min-score` | 60 | Minimum quality score (0-100) |
| `-max-age` | 365 | Maximum paper age in days (0 = no limit) |
| `-skip-db` | false | Skip database operations |
| `-skip-filter` | false | Skip quality filtering |

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/papers` | List papers (with pagination) |
| GET | `/api/papers/:id` | Get paper by ID |
| GET | `/api/papers/search?q=` | Search papers |
| GET | `/api/stats` | Pipeline statistics |
| POST | `/api/sync` | Trigger paper sync |
| GET | `/health` | Health check |

### Project Structure

```
genesis-pipeline/
├── cmd/
│   ├── pipeline/       # CLI for data ingestion
│   ├── api/            # REST API server
│   └── benchmark/      # Performance benchmarks
├── internal/
│   ├── model/          # Data models
│   ├── parser/         # ArXiv API client
│   ├── filter/         # Quality filtering & scoring
│   ├── storage/        # PostgreSQL repository
│   ├── validation/     # Data quality checks
│   ├── benchmark/      # Benchmark utilities
│   └── api/            # HTTP handlers
└── deployments/        # Docker configurations
```

### Tech Stack

- **Language**: Go 1.21+
- **Database**: PostgreSQL 16
- **Driver**: pgx/v5
- **Container**: Docker & Docker Compose

---

## 中文

一个用 Go 构建的科学论文数据管道，用于从 ArXiv 采集和存储论文数据。

### 架构

```
┌─────────────────────────────────────────────────────────────────┐
│                     Genesis Pipeline                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌──────────┐      ┌──────────┐      ┌──────────────────────┐  │
│   │  ArXiv   │      │  解析层   │      │      存储层          │  │
│   │   API    │─────>│ (Parser) │─────>│   (PostgreSQL)       │  │
│   │          │ XML  │          │ Go   │                      │  │
│   └──────────┘      └──────────┘      └──────────────────────┘  │
│                           │                      │               │
│                           v                      v               │
│                    ┌──────────┐          ┌──────────┐           │
│                    │ 数据验证  │          │ REST API │           │
│                    │  Layer   │          │ (查询接口) │           │
│                    └──────────┘          └──────────┘           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 功能特性

- **数据采集**: 从 ArXiv API 获取论文，支持自定义查询条件
- **质量过滤**: 双层过滤系统，确保高质量论文
  - Level 1: 硬过滤（接收信号、DOI、强实证）
  - Level 2: 打分 (0-100)，基于评估关键词、代码链接等
- **时效过滤**: 按发布时间过滤（可配置最大天数）
- **增量更新**: 追踪同步历史，统计新增/更新论文数量
- **数据验证**: 验证论文元数据质量
- **PostgreSQL 存储**: 持久化论文数据，支持搜索
- **REST API**: 通过 HTTP 接口查询已存储的论文
- **性能基准测试**: 性能指标和数据质量报告

### 快速开始

```bash
# 克隆仓库
git clone https://github.com/1psychoQAQ/genesis-pipeline.git
cd genesis-pipeline

# 启动 PostgreSQL
docker-compose -f deployments/docker-compose.yml up -d

# 运行数据管道（带质量+时效过滤）
go run cmd/pipeline/main.go -query "machine learning" -limit 100 -max-age 90 -min-score 60

# 启动 API 服务
go run cmd/api/main.go -port 8088

# 运行性能测试
go run cmd/benchmark/main.go -limit 100
```

### 管道参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-query` | "machine learning" | ArXiv 搜索查询词 |
| `-limit` | 10 | 获取论文数量 |
| `-min-score` | 60 | 最低质量分数 (0-100) |
| `-max-age` | 365 | 最大论文天数 (0 = 不限制) |
| `-skip-db` | false | 跳过数据库操作 |
| `-skip-filter` | false | 跳过质量过滤 |

### API 接口

| 方法 | 端点 | 描述 |
|------|------|------|
| GET | `/api/papers` | 论文列表（支持分页） |
| GET | `/api/papers/:id` | 根据 ID 获取论文 |
| GET | `/api/papers/search?q=` | 搜索论文 |
| GET | `/api/stats` | 管道统计信息 |
| POST | `/api/sync` | 触发论文同步 |
| GET | `/health` | 健康检查 |

### 项目结构

```
genesis-pipeline/
├── cmd/
│   ├── pipeline/       # 数据采集 CLI
│   ├── api/            # REST API 服务
│   └── benchmark/      # 性能基准测试
├── internal/
│   ├── model/          # 数据模型
│   ├── parser/         # ArXiv API 客户端
│   ├── filter/         # 质量过滤与打分
│   ├── storage/        # PostgreSQL 存储层
│   ├── validation/     # 数据质量验证
│   ├── benchmark/      # 基准测试工具
│   └── api/            # HTTP 处理器
└── deployments/        # Docker 配置
```

### 数据模型

```go
type Paper struct {
    ID         string    // ArXiv 唯一标识符
    Title      string    // 论文标题
    Abstract   string    // 摘要全文
    Authors    []string  // 作者列表
    Categories []string  // 学术分类 (如 cs.AI)
    UpdatedAt  time.Time // 最后更新时间
}
```

### 技术栈

- **语言**: Go 1.21+
- **数据库**: PostgreSQL 16
- **驱动**: pgx/v5
- **容器化**: Docker & Docker Compose

### 开发

```bash
# 运行测试
go test ./...

# 运行性能测试
go test -bench=. ./...

# 构建所有二进制文件
go build ./...
```

---

## License

MIT License
