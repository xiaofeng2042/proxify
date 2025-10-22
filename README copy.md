<div align="center">
  <a href="https://github.com/poixeai/proxify">
    <img src="https://proxify.poixe.com/x.svg" alt="Proxify Logo" width="100" height="100">
  </a>
  <h1>Proxify</h1>
  <p>An open-source, lightweight, and self-hosted reverse proxy gateway for AI APIs</p>

  <p>
    <a href="https://github.com/poixeai/proxify" target="_blank">
      <img src="https://img.shields.io/badge/View_on_GitHub-181717?style=for-the-badge&logo=github&logoColor=white" alt="GitHub Link">
    </a>
  </p>

  <img src="https://s3.poixe.com/apple/home_en_bg.png" alt="Proxify Logo">
</div>

---

**Proxify** is a high-performance reverse proxy gateway written in Go.
It allows developers to access various large model APIs through a unified entry point — solving problems such as regional restrictions and multi-service configuration complexity.
Proxify is deeply optimized for LLM streaming responses, ensuring the best performance and smooth user experience.

## ✨ Features

- 💎 **Powerful Extensibility**:
  More than just an AI gateway — Proxify is a universal reverse proxy server with special optimizations for LLM APIs, including stream smoothing, heartbeat keepalive, and tail acceleration.

- 🚀 **Unified API Entry**:
  Route to multiple upstreams through a single-level path — e.g., `/openai` → `api.openai.com`, `/gemini` → `generativelanguage.googleapis.com`.
  All routes are defined in one configuration file for simplicity and efficiency.

- ⚡ **Lightweight & High Performance**:
  Built with Golang and natively supports high concurrency with minimal memory usage. Runs smoothly on servers with as little as 0.5 GB RAM.

- 🚄 **Stream Optimization**:

  - **Smooth Output**: Built-in flow controller ensures a "typing effect" by streaming model responses smoothly.
  - **Heartbeat Keepalive**: Automatically inserts heartbeat messages into SSE (Server-Sent Events) streams to prevent idle timeouts.
  - **Tail Acceleration**: Keeps latency under control by accelerating the final part of the response.

- 🛡️ **Security & Privacy**:
  Fully self-hosted — all requests and data remain under your control. No third-party servers are involved, ensuring zero privacy risk.

- 🌐 **Broad Compatibility**:
  Preconfigured routes for major AI providers like OpenAI, Azure, Claude, Gemini, and DeepSeek. Easily extendable to any HTTP API via configuration.

- 🔧 **Easy Integration**:
  Switch from your existing API service to Proxify simply by updating the `BaseURL` — no code changes or request parameter modifications required.

- 👨‍💻 **Open Source & Professional**:
  Designed and maintained by a young and experienced AI engineering team. 100% open-source, auditable, and community-driven (PRs and Issues are welcome).

## 🛠️ Tech Stack

- **Backend Gateway**: Golang + Gin
- **Frontend Dashboard**: React + Vite + TypeScript + Tailwind CSS

## 🚀 Quick Start <a id="-quick-start"></a>

Integrating your existing services with Proxify only takes three steps.

#### 1. Identify Target Service

Browse the [Supported API list](https://proxify.poixe.com/api/routes) and find the proxy path prefix (Path) for the desired service.

#### 2. Replace the Base URL

Replace the original API base URL in your code with your Proxify deployment address, appending the route prefix.

- **Original**:
  `https://api.openai.com/v1/chat/completions`
- **Replaced with**:
  `http://<your-proxify-domain>/openai/v1/chat/completions`

#### 3. Send Requests

Done! Use your existing API key and parameters as usual.
Your headers and request body remain unchanged.

#### Example (Node.js - OpenAI SDK)

```javascript
import OpenAI from "openai";

const openai = new OpenAI({
  apiKey: "sk-...", // your OpenAI API key
  baseURL: "http://127.0.0.1:7777/openai/v1", // your Proxify address
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

## 🖥️ Deployment Guide <a id="-deployment-guide"></a>

Proxify offers multiple deployment options.
Before starting, make sure you’ve completed the setup steps below.

---

### ⚙️ Step 1: Configure Environment & Routes

Proxify includes `.env.example` and `routes.json.example`.
Copy and adjust them to your needs.

#### **1. Environment Variables (`.env`)**

```bash
cp .env.example .env
```

Example:

```env
# Mode: debug | release
MODE=debug

# Server port
PORT=7777

# Optional GitHub token
GITHUB_TOKEN=ghp_xxxx

# Stream optimization
STREAM_SMOOTHING_ENABLED=true
STREAM_HEARTBEAT_ENABLED=true
```

> 💡 **Tips:**
>
> - For Docker, mount `.env` into `/app/.env` inside the container.
> - For local binary, keep `.env` in the same directory as the executable.

---

#### **2. Route Configuration (`routes.json`)**

```bash
cp routes.json.example routes.json
```

Example:

```json
{
  "routes": [
    {
      "name": "OpenAI",
      "description": "OpenAI Official API Endpoint",
      "path": "/openai",
      "target": "https://api.openai.com/"
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
  ]
}
```

> 💡 Routes can be modified freely — changes are automatically hot-reloaded without restarting the service.

---

### 🐳 Option 1: Deploy with Docker (Recommended)

We provide three convenient Docker deployment methods.

#### 1. Pull from Docker Hub (Simplest)

This is the fastest and most recommended way to deploy in production.

```bash
# 1. Pull the latest image from Docker Hub
docker pull poixeai/proxify:latest

# 2. Run the container and mount configuration files
docker run -d \
  --name proxify \
  -p 7777:7777 \
  -v $(pwd)/routes.json:/app/routes.json \
  -v $(pwd)/.env:/app/.env \
  --restart=always \
  poixeai/proxify:latest
```

#### 2. Use Docker Compose (Recommended)

Manage your service declaratively via a `docker-compose.yml` file for better maintainability.

1. **Ensure the `docker-compose.yml` file exists in your current directory.**

2. **Start the service:**

   ```bash
   # Start the service
   docker-compose up -d

   # Check service status
   docker-compose ps
   ```

#### 3. Build from Dockerfile

If you want to build your own image based on the latest source code.

```bash
# 1. Clone the repository
git clone https://github.com/poixeai/proxify.git
cd proxify

# 2. Build your own image (for example, name it my-proxify)
docker build -t poixeai/proxify:latest .

# 3. Run the image you just built
docker run -d \
  --name proxify \
  -p 7777:7777 \
  -v $(pwd)/routes.json:/app/routes.json \
  -v $(pwd)/.env:/app/.env \
  --restart=always \
  poixeai/proxify:latest
```

---

### 🛠️ Option 2: Manual Build and Run

For development environments or when Docker is not preferred.

**Requirements:**

- Go (version 1.20+)
- Node.js (version 18+)
- pnpm

#### 1. Use the Build Script (Recommended)

We provide a `build.sh` script to simplify the compilation process.

```bash
# 1. Clone the repository and enter the directory
git clone https://github.com/poixeai/proxify.git
cd proxify

# 2. Grant execution permission to the script
chmod +x build.sh

# 3. Run the build script
./build.sh

# 4. Run the compiled binary
./bin/proxify
```

#### 2. Fully Manual Build

If you prefer to understand the full build process.

```bash
# 1. Clone the repository and enter the directory
git clone https://github.com/poixeai/proxify.git
cd proxify

# 2. Build frontend static assets
cd web
pnpm install
pnpm build
cd ..

# 3. Build backend Go application
go mod tidy
go build -o ./bin/proxify .

# 4. Run the binary
./bin/proxify
```

---

## 🗺️ Supported Endpoints <a id="-supported-endpoints"></a>

Proxify can proxy **any HTTP service**.
Below are the preconfigured and optimized AI API routes:

| Provider       | Path          | Target URL                                  |
| -------------- | ------------- | ------------------------------------------- |
| **OpenAI**     | `/openai`     | `https://api.openai.com`                    |
| **Azure**      | `/azure`      | `https://<your-res>.openai.azure.com`       |
| **DeepSeek**   | `/deepseek`   | `https://api.deepseek.com`                  |
| **Claude**     | `/claude`     | `https://api.anthropic.com`                 |
| **Gemini**     | `/gemini`     | `https://generativelanguage.googleapis.com` |
| **Grok**       | `/grok`       | `https://api.x.ai`                          |
| **Aliyun**     | `/aliyun`     | `https://dashscope.aliyuncs.com`            |
| **VolcEngine** | `/volcengine` | `https://ark.cn-beijing.volces.com`         |

_⚠️ Actual available routes depend on your `routes.json` configuration._

### 🔍 View Live Demo Routes

```bash
GET https://proxify.poixe.com/api/routes
```

👉 [View Current Demo Routes](https://proxify.poixe.com/api/routes)

---

## 🤝 Contributing

We welcome all contributions — whether it’s filing an issue, submitting a PR, or improving documentation.
Your support helps the community grow.

## 📄 License

This project is licensed under the [MIT License](https://github.com/poixeai/proxify/blob/main/LICENSE).
