<div align="center">
  <a href="https://github.com/poixeai/proxify">
    <img src="./web/public/x.svg" alt="Proxify Logo" width="100" height="100">
  </a>
  <h1>Proxify</h1>
  <p>一个开源、轻量、自托管的 AI 接口反向代理网关</p>

  [English](./README.md) / 简体中文
  
  <p>
    <a href="https://github.com/poixeai/proxify/blob/main/LICENSE">
      <img alt="License" src="https://img.shields.io/github/license/poixeai/proxify?style=for-the-badge&color=blue">
    </a>
    <a href="https://github.com/poixeai/proxify">
      <img alt="Go Version" src="https://img.shields.io/github/go-mod/go-version/poixeai/proxify?style=for-the-badge">
    </a>
    <a href="https://github.com/poixeai/proxify/stargazers">
      <img alt="Stars" src="https://img.shields.io/github/stars/poixeai/proxify?style=for-the-badge&logo=github">
    </a>
    <a href="https://github.com/poixeai/proxify/issues">
      <img alt="Issues" src="https://img.shields.io/github/issues/poixeai/proxify?style=for-the-badge&logo=github">
    </a>
  </p>

  <h4>
    <a href="https://proxify.poixe.com">演示网站</a>
    <span> · </span>
    <a href="#-快速开始">快速开始</a>
    <span> · </span>
    <a href="#-部署教程">部署教程</a>
    <span> · </span>
    <a href="#-支持端点">支持端点</a>
  </h4>
  
  <img src="./assets/images/home_zh_bg.png" alt="Proxify Logo">
</div>

---

**Proxify** 是一个用 Go 编写的高性能反向代理网关。它允许开发者通过统一的入口访问各类大模型 API，解决了地区限制、多服务配置复杂等问题。Proxify 对 LLM 的流式响应进行了深度优化，确保了最佳的性能和用户体验。

## ✨ 功能特性

- 💎 **强大扩展能力**：不仅是 AI 接口网关，Proxify 也是一个通用的反向代理服务器。我们对 LLM API 做了专项优化，包括流式传输、心跳保活、尾部冲刺等。

- 🚀 **统一 API 入口**：通过一级路径即可路由到不同上游，例如 `/openai` → `api.openai.com`，`/gemini` → `generativelanguage.googleapis.com`。所有路由规则一处配置，简单高效。

- ⚡ **轻量与高性能**：后端采用 Golang 构建，原生支持高并发，资源占用极低。在 0.5 GB 内存的服务器上也能轻松流畅运行。

- 🚄 **极致流式优化**：

  - **平滑输出**：内置流控器，将大模型快速生成的文本块平滑地以“打字机”效果流式传输给客户端。

  - **心跳维持**：在 SSE (Server-Sent Events) 流中自动插入心跳消息，有效防止因网络空闲导致的连接意外中断。
  - **尾部冲刺**：在保障丝滑输出的同时，通过尾部冲刺技术将最坏延迟控制在可接受范围，优化最终响应时间。

- 🛡️ **安全与隐私**：自托管部署，所有请求数据完全在您自己的掌控之下，不经过任何第三方服务器，彻底杜绝隐私泄露风险。

- 🌐 **广泛兼容性**：已预置 OpenAI、Azure、Claude、Gemini、DeepSeek 等主流 AI 服务商的路由，同时支持通过配置文件便捷地横向扩展到任意 API。

- 🔧 **极致易用**：从现有服务切换到 Proxify，通常只需修改一行 `BaseURL` 配置，无需改动任何业务代码或请求参数。

- 👨‍💻 **开源与专业**：项目由 AI 领域一支年轻且富有经验的团队设计与维护，代码完全开源、透明可审计，杜绝供应商锁定，欢迎社区贡献（PRs / Issues）。

## 🛠️ 技术栈

- **后端网关**: Golang + Gin

- **前端面板**: React + Vite + TypeScript + Tailwind CSS

## 🚀 快速开始 <a id="-快速开始"></a>

将您现有的服务对接到 Proxify 只需三步。

#### 1. 确定目标服务

浏览我们支持的 [API 列表](#-支持端点)，找到您需要的服务及其对应的代理路径前缀（Path）。

#### 2. 替换基础 URL

在您的代码中，将原始 API 的基础 URL 替换为您的 Proxify 部署地址，并附加上一步中确定的路径前缀。

- **原始地址**: `https://api.openai.com/v1/chat/completions`

- **替换为**: `http://<您的-Proxify-地址>/openai/v1/chat/completions`

#### 3. 发送请求

一切就绪！使用您原有的 API 密钥和请求参数，像往常一样发起请求。您的 API 密钥、请求头（Header）和请求体（Body）等所有其他部分都保持不变。

#### 代码示例 (Node.js - OpenAI SDK)

```javascript
// Node.js example using /openai proxy endpoint
import OpenAI from "openai";
const openai = new OpenAI({
  // highlight-start
  apiKey: "sk-...", // 您的 OpenAI API Key
  baseURL: "http://127.0.0.1:7777/openai/v1", // 指向您的 Proxify 服务
  // highlight-end
});
async function main() {
  const stream = await openai.chat.completions.create({
    model: "gpt-5",
    messages: [{ role: "user", content: "hi" }],
    stream: true,
  });
  for await (const chunk of stream) {
    process.stdout.write(chunk.choices[0]?.delta?.content || "");
  }
}
main();
```

## 🖥️ 部署教程 <a id="-部署教程"></a>

Proxify 提供多种部署方式。无论您选择哪种方式，请先完成准备工作。

---

### ⚙️ 准备工作：配置环境与路由

Proxify 项目已经内置了 `.env.example` 与 `routes.json.example` 示例文件。
您只需复制并稍作修改，即可快速启动。

#### **1. 环境变量配置 (`.env`)**

从示例文件复制：

```bash
cp .env.example .env
```

示例内容如下（请根据实际情况修改）：

```env
# 运行模式：debug | release
MODE=debug

# 服务监听端口
PORT=7777

# 可选：GitHub Token（用于访问私有仓库或限流提升）
GITHUB_TOKEN=ghp_xxxx

# 启用流式优化（平滑输出模式）
STREAM_SMOOTHING_ENABLED=true

# 启用 SSE 心跳机制（防止长连接超时）
STREAM_HEARTBEAT_ENABLED=true

# IP 白名单（可选）
# 支持单个 IP、CIDR 网段，多个规则使用英文逗号分隔
AUTH_IP_WHITELIST="127.0.0.1,10.0.0.0/8,192.168.1.0/24,::1"

# Token 鉴权（可选）
AUTH_TOKEN_HEADER="X-API-Token"
AUTH_TOKEN_KEY="your-super-secret-token"
```

> 💡 **提示：**
>
> - 当您使用 Docker 运行 Proxify 时，必须将 .env 文件挂载进容器内部路径 /app/.env；
>
> - 如果您直接运行本地可执行文件（不使用 Docker），只需保证 .env 与程序位于同一目录即可。
>
> - 所有标记为「可选」的配置项（如 `GITHUB_TOKEN`、`AUTH_IP_WHITELIST`、`AUTH_TOKEN_*`），**留空或未设置时将不会启用对应功能**。

---

#### **2. 路由配置文件 (`routes.json`)**

从示例文件复制：

```bash
cp routes.json.example routes.json
```

该文件定义了所有可代理的上游 AI 模型端点。

示例内容如下（可直接使用或新增）：

```json
{
  "routes": [
    {
      "name": "OpenAI",
      "description": "OpenAI Official API Endpoint",
      "path": "/openai",
      "target": "https://api.openai.com/",
      "model_map": {
        "gpt-4o": "gpt-4o-2024-11-20"
      }
    },
    {
      "name": "DeepSeek",
      "description": "DeepSeek Official API Endpoint",
      "path": "/deepseek",
      "target": "https://api.deepseek.com"
    },
    {
      "name": "Claude",
      "description": "Anthropic Claude Official API Endpoint",
      "path": "/claude",
      "target": "https://api.anthropic.com"
    },
    {
      "name": "Gemini",
      "description": "Google Gemini Official API Endpoint",
      "path": "/gemini",
      "target": "https://generativelanguage.googleapis.com"
    }
    ...
  ]
}
```

> 💡 **提示：**
>
> - 您可在此文件中自由增减代理路径；
>
> - 修改后无需重启（路由文件自动热加载）。
>
> - 支持在路由级别对请求体中的 `model` 字段进行重写，常用于模型别名、自动降级或跨平台兼容。

---

### 🧾 准备完成后

在确认 `.env` 与 `routes.json` 均配置正确后，
即可选择以下任意一种部署方式继续操作：

- [🐳 使用 Docker 部署（推荐）](#-方式一使用-docker-推荐)
- [🛠️ 手动编译运行（开发环境）](#️-方式二手动编译部署)

---

### 🐳 方式一：使用 Docker (推荐)

我们提供了三种便捷的 Docker 部署方案。

#### 1. 从 Docker Hub 拉取镜像 (最简单)

这是最快、最推荐的生产环境部署方式。

```bash
# 1. 从 Docker Hub 拉取最新镜像
docker pull terobox/proxify:latest

# 2. 运行容器，并挂载配置文件
docker run -d \
  --name proxify \
  -p 7777:7777 \
  -v $(pwd)/routes.json:/app/routes.json \
  -v $(pwd)/.env:/app/.env \
  --restart=always \
  poixeai/proxify:latest
```

#### 2. 使用 Docker Compose (推荐)

通过 `docker-compose.yml` 文件进行声明式部署，便于管理。

1.  **保证 `docker-compose.yml` 文件已创建，且位于当前目录下。**

2.  **启动服务:**

    ```bash
    # 启动服务
    docker-compose up -d

    # 查看服务状态
    docker-compose ps
    ```

#### 3. 从 Dockerfile 构建镜像

如果您需要基于最新源码进行自定义构建。

```bash
# 1. 克隆源码
git clone https://github.com/poixeai/proxify.git
cd proxify

# 2. 使用 Dockerfile 构建你自己的镜像 (例如，命名为 my-proxify)
docker build -t terobox/proxify:latest .

# 3. 运行您刚刚构建的镜像
docker run -d \
  --name proxify \
  -p 7777:7777 \
  -v $(pwd)/routes.json:/app/routes.json \
  -v $(pwd)/.env:/app/.env \
  --restart=always \
  poixeai/proxify:latest
```

---

### 🛠️ 方式二：手动编译部署

适用于开发环境或不便使用 Docker 的场景。

**环境要求:**

- Go (版本 1.20+)
- Node.js (版本 18+)
- pnpm

#### 1. 使用构建脚本 (推荐)

我们提供了 `build.sh` 脚本来简化编译流程。

```bash
# 1. 克隆源码并进入目录
git clone https://github.com/poixeai/proxify.git
cd proxify

# 2. 赋予脚本执行权限
chmod +x build.sh

# 3. 执行构建脚本
./build.sh

# 4. 运行编译好的程序
./bin/proxify
```

#### 2. 完全手动编译

如果您想了解完整的构建步骤。

```bash
# 1. 克隆源码并进入目录
git clone https://github.com/poixeai/proxify.git
cd proxify

# 2. 编译前端静态资源
cd web
pnpm install
pnpm build
cd ..

# 3. 编译后端 Go 程序
go mod tidy
go build -o ./bin/proxify .

# 4. 运行程序
./bin/proxify
```

## 🗺️ 广泛兼容的 API 端点 <a id="-支持端点"></a>


Proxify 支持代理任何 HTTP 服务。以下是一些预设的、经过优化的常用 AI 服务路由示例。

| 服务商       | 建议路径 (Path) | 目标地址 (URL)                              |
| ------------ | --------------- | ------------------------------------------- |
| **OpenAI**   | `/openai`       | `https://api.openai.com`                    |
| **Azure**    | `/azure`        | `https://<your-res>.openai.azure.com`       |
| **DeepSeek** | `/deepseek`     | `https://api.deepseek.com`                  |
| **Claude**   | `/claude`       | `https://api.anthropic.com`                 |
| **Gemini**   | `/gemini`       | `https://generativelanguage.googleapis.com` |
| **Grok**     | `/grok`         | `https://api.x.ai`                          |
| **阿里云**   | `/aliyun`       | `https://dashscope.aliyuncs.com`            |
| **火山引擎** | `/volcengine`   | `https://ark.cn-beijing.volces.com`         |

_注意：实际可用路径取决于您的 `routes.json` 配置文件。_

### 🔍 查看当前演示站支持的端口

您可以通过以下接口实时查看演示站当前配置的代理端口列表：

```bash
GET https://proxify.poixe.com/api/routes
````

👉 [点击查看演示站当前支持的端口](https://proxify.poixe.com/api/routes)


## 🤝 贡献

我们欢迎并感谢所有形式的贡献！无论是提交 Issue、发起 Pull Request，还是改进文档，都是对社区的巨大支持。

## 📄 开源协议

本项目采用 [MIT License](./LICENSE) 开源协议。
