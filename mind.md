# 角色：资深 Go 后端与数据工程师
# 任务：**构建科学数据管线** (Genesis Research Pipeline) 骨架

## 1. 项目概述
我正在构建一个名为 `genesis-research-pipeline` 的 Go 后端项目。目标是为 ArXiv API 的科学文献建立自动化数据管线，展示以下工程能力：
- 数据采集与解析 (Ingestion & Parsing)
- 结构化数据存储 (Storage)
- 系统化评测基准 (Benchmarking)

## 2. 技术栈
- **语言**: Go (Golang)
- **数据库**: PostgreSQL (通过 Docker 部署)
- **容器化**: Docker & Docker Compose

## 3. 第一阶段：工程骨架任务清单

### A. 定义数据模型
在 `internal/model/paper.go` 中创建一个 `Paper` 结构体，包含以下字段：
- `ID` (string): ArXiv 的唯一标识符。
- `Title` (string): 论文标题。
- `Abstract` (string): 论文摘要全文。
- `Authors` ([]string): 作者姓名列表。
- `Categories` ([]string): 学术分类标签（如 cs.AI, cond-mat）。
- `UpdatedAt` (time.Time): 最后更新时间戳。

### B. 项目目录结构
生成标准的 Go 项目布局：
- `cmd/`: 应用程序入口。
- `internal/`: 私有业务逻辑（模型、解析器、存储）。
- `deployments/`: Docker 配置相关。

### C. 基础设施 (Docker)
生成 `deployments/docker-compose.yml` 文件，包含：
- 一个 **PostgreSQL** 服务。
- 必要的环境变量（用户名、密码、数据库名）。
- 健康检查 (Healthcheck)，确保数据库就绪。

### D. 接口定义
在 `internal/parser/provider.go` 中定义一个 `Provider` 接口：
- 方法签名: `FetchPapers(query string, limit int) ([]model.Paper, error)`

## 4. 开发要求
- 遵循 Go 语言