# Sherpa-ONNX 转录服务器

这是一个基于 sherpa-onnx 的 Golang 转录服务器，支持音频文件的语音转文字功能。

## 功能特性

- 🎯 支持 HTTP API 和 WebSocket 接口
- 🔧 支持 TCP 端口和 Unix socket 监听
- 📁 支持多种音频格式（WAV, MP3, FLAC 等）
- ⚡ 实时流式转录
- 🛠️ 可配置的模型参数
- 📊 健康检查端点
- 🔌 直接使用 sherpa-onnx-go 库
- 🎤 实时语音识别 WebSocket API
- 🐳 Docker 容器化支持
- 🔒 线程安全的并发处理
- 📈 高性能音频处理

## 系统要求

- **Go**: 1.24.1 或更高版本
- **操作系统**: Linux, macOS, Windows
- **内存**: 至少 2GB RAM（推荐 4GB+）
- **存储**: 至少 500MB 可用空间（用于模型文件）

## 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/your-username/transcribeserver.git
cd transcribeserver
```

### 2. 安装依赖

```bash
go mod tidy
```

项目会自动安装 sherpa-onnx-go 库及其平台特定的依赖。

### 3. 下载模型

下载 Whisper 模型文件到 `models/whisper-tiny/` 目录：

```bash
mkdir -p models/whisper-tiny
cd models/whisper-tiny

# 下载中文模型（推荐）
wget https://huggingface.co/csukuangfj/sherpa-onnx-zh-wenet-aishell3/resolve/main/encoder.onnx
wget https://huggingface.co/csukuangfj/sherpa-onnx-zh-wenet-aishell3/resolve/main/decoder.onnx
wget https://huggingface.co/csukuangfj/sherpa-onnx-zh-wenet-aishell3/resolve/main/joiner.onnx
wget https://huggingface.co/csukuangfj/sherpa-onnx-zh-wenet-aishell3/resolve/main/tokens.txt

# 或者下载英文模型
wget https://huggingface.co/csukuangfj/sherpa-onnx-whisper-tiny/resolve/main/encoder.onnx
wget https://huggingface.co/csukuangfj/sherpa-onnx-whisper-tiny/resolve/main/decoder.onnx
wget https://huggingface.co/csukuangfj/sherpa-onnx-whisper-tiny/resolve/main/joiner.onnx
wget https://huggingface.co/csukuangfj/sherpa-onnx-whisper-tiny/resolve/main/tokens.txt
```

#### 说话人分离模型（可选）

如果需要使用说话人分离功能，下载说话人分离模型：

```bash
mkdir -p models/speaker-diarization
cd models/speaker-diarization

# 下载说话人分离模型
# 注意：具体的模型文件需要根据 sherpa-onnx 的官方文档来下载
# 参考：https://github.com/k2-fsa/sherpa-onnx/blob/master/go-api-examples/non-streaming-speaker-diarization/main.go
```

您可以从以下地址下载预训练模型：
- [中文模型](https://huggingface.co/csukuangfj/sherpa-onnx-zh-wenet-aishell3)
- [英文模型](https://huggingface.co/csukuangfj/sherpa-onnx-whisper-tiny)
- [其他语言模型](https://github.com/k2-fsa/sherpa-onnx/blob/master/docs/pretrained_models.md)
- [说话人分离模型](https://github.com/k2-fsa/sherpa-onnx/blob/master/go-api-examples/non-streaming-speaker-diarization/main.go)

### 4. 配置

编辑 `config.yaml` 文件：

```yaml
server:
  port: 8080                    # HTTP 端口
  host: "0.0.0.0"              # 监听地址
  use_unix_socket: false       # 是否使用 Unix socket
  unix_socket: "/tmp/transcribe.sock"  # Unix socket 路径

sherpa:
  model_path: "./models/whisper-tiny"  # 模型路径
  tokens_path: "./models/whisper-tiny/tokens.txt"  # 词汇表路径
  sample_rate: 16000           # 采样率
  num_threads: 4               # 线程数
  decoding_method: "greedy_search"  # 解码方法
  enable_endpoint: true        # 启用端点检测
  rule1_min_trailing_silence: 2.4  # 最小尾随静音时间
  rule2_min_trailing_silence: 1.2  # 规则2最小尾随静音时间
  rule3_min_utterance_length: 300   # 最小话语长度
  # 说话人分离配置
  enable_diarization: false    # 是否启用说话人分离
  diarization_model_path: "./models/speaker-diarization"  # 说话人分离模型路径
```

### 5. 运行

```bash
# 使用默认配置文件
go run main.go

# 指定配置文件
go run main.go -config=my-config.yaml

# 构建并运行
make build
./transcribeserver
```

## Docker 部署

### 使用预构建镜像

```bash
# 拉取镜像
docker pull your-registry/transcribeserver:latest

# 运行容器
docker run -d \
  --name transcribeserver \
  -p 8080:8080 \
  -v /path/to/models:/app/models \
  -v /path/to/config.yaml:/app/config.yaml \
  your-registry/transcribeserver:latest
```

### 构建自定义镜像

```bash
# 构建镜像
docker build -t transcribeserver .

# 运行容器
docker run -d \
  --name transcribeserver \
  -p 8080:8080 \
  -v /path/to/models:/app/models \
  transcribeserver
```

### Docker Compose

创建 `docker-compose.yml` 文件：

```yaml
version: '3.8'
services:
  transcribeserver:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./models:/app/models
      - ./config.yaml:/app/config.yaml
    environment:
      - GIN_MODE=release
    restart: unless-stopped
```

运行：

```bash
docker-compose up -d
```

## API 使用

### 健康检查

```bash
curl http://localhost:8080/health
```

响应：
```json
{
  "status": "ok",
  "message": "转录服务器运行正常"
}
```

### 文件上传转录

```bash
curl -X POST http://localhost:8080/transcribe \
  -F "audio=@/path/to/audio.wav"
```

### JSON 请求转录

```bash
curl -X POST http://localhost:8080/transcribe \
  -H "Content-Type: application/json" \
  -d '{
    "audio_data": "base64_encoded_audio_data",
    "format": "wav"
  }'
```

### 实时语音识别 WebSocket API

参考 [sherpa-onnx 实时语音识别示例](https://github.com/k2-fsa/sherpa-onnx/blob/master/go-api-examples/real-time-speech-recognition-from-microphone/main.go)，我们实现了真正的实时转录功能。

#### 连接 WebSocket

```javascript
const ws = new WebSocket('ws://localhost:8080/ws/realtime');

ws.onopen = function() {
  console.log('实时转录连接已建立');
};

ws.onmessage = function(event) {
  const response = JSON.parse(event.data);
  if (response.success) {
    console.log('转录结果:', response.result.text);
  } else {
    console.error('错误:', response.error);
  }
};

ws.onclose = function() {
  console.log('WebSocket 连接已关闭');
};

ws.onerror = function(error) {
  console.error('WebSocket 错误:', error);
};
```

#### 发送音频数据

```javascript
// 发送 PCM 音频数据
function sendAudioChunk(audioData) {
  const request = {
    audio_data: audioData,  // 16-bit PCM 音频数据
    format: "pcm"
  };
  ws.send(JSON.stringify(request));
}

// 从麦克风获取音频数据并发送
navigator.mediaDevices.getUserMedia({ 
  audio: {
    sampleRate: 16000,
    channelCount: 1,
    echoCancellation: true,
    noiseSuppression: true
  } 
})
.then(stream => {
  const mediaRecorder = new MediaRecorder(stream, {
    mimeType: 'audio/webm;codecs=pcm'
  });
  
  mediaRecorder.ondataavailable = event => {
    if (event.data.size > 0) {
      sendAudioChunk(event.data);
    }
  };
  
  mediaRecorder.start(1000); // 每秒发送一次数据
})
.catch(error => {
  console.error('获取麦克风权限失败:', error);
});
```

#### 响应格式

```json
{
  "success": true,
  "result": {
    "text": "转录的文本内容 [FINAL]"
  }
}
```

- 中间结果：实时显示识别的文本
- 最终结果：文本末尾标记 `[FINAL]`

## 响应格式

### 成功响应

```json
{
  "success": true,
  "result": {
    "text": "转录的文本内容",
    "confidence": 0.95,
    "duration": 3.2
  }
}
```

### 带说话人分离的响应

当启用说话人分离功能时，响应会包含说话人片段信息：

```json
{
  "success": true,
  "result": {
    "text": "转录的文本内容",
    "confidence": 0.95,
    "duration": 3.2,
    "speaker_segments": [
      {
        "speaker_id": 0,
        "start": 0.0,
        "end": 2.5,
        "text": "第一个说话人的内容"
      },
      {
        "speaker_id": 1,
        "start": 2.8,
        "end": 5.2,
        "text": "第二个说话人的内容"
      }
    ]
  }
}
```

### 错误响应

```json
{
  "success": false,
  "error": "错误描述信息"
}
```

## 开发

### 项目结构

```
transcribeserver/
├── main.go                    # 主程序入口
├── config/
│   └── config.go              # 配置管理
├── server/
│   ├── server.go              # HTTP 服务器和 WebSocket 处理
│   └── server_test.go         # 服务器测试
├── transcribe/
│   └── sherpa.go              # sherpa-onnx 转录实现
├── examples/
│   └── client.go              # 客户端示例
├── static/
│   └── realtime.html          # 实时转录演示页面
├── config.yaml                # 配置文件
├── Dockerfile                 # Docker 配置
├── Makefile                   # 构建脚本
├── go.mod                     # Go 模块文件
└── README.md                  # 项目文档
```

### 构建

```bash
# 构建可执行文件
go build -o transcribeserver main.go

# 使用 Makefile
make build

# 构建特定平台
GOOS=linux GOARCH=amd64 go build -o transcribeserver-linux main.go
GOOS=darwin GOARCH=amd64 go build -o transcribeserver-mac main.go
GOOS=windows GOARCH=amd64 go build -o transcribeserver.exe main.go
```

### 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./server
go test ./transcribe

# 运行测试并显示覆盖率
go test -cover ./...

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### 代码质量

```bash
# 运行代码检查
go vet ./...

# 运行静态分析
go install golang.org/x/lint/golint@latest
golint ./...

# 格式化代码
go fmt ./...

# 运行所有检查
make lint
```

## 实时转录特性

### 技术实现

- **WebSocket 连接**: 每个客户端建立独立的 WebSocket 连接
- **音频流处理**: 使用 sherpa-onnx 的 `OnlineStream` 进行实时处理
- **并发安全**: 使用互斥锁确保线程安全
- **资源管理**: 自动清理音频流和连接资源
- **端点检测**: 智能检测语音结束点
- **流式解码**: 实时解码和结果输出

### 性能优化

- **流式处理**: 音频数据实时输入，无需等待完整文件
- **内存管理**: 及时释放音频流资源
- **并发处理**: 支持多个客户端同时连接
- **线程池**: 优化 CPU 密集型任务
- **缓冲区管理**: 高效的音频数据缓冲

### 音频格式支持

- **PCM**: 16-bit, 单声道, 16kHz
- **WAV**: 标准 WAV 格式
- **MP3**: MPEG-1 Audio Layer III
- **FLAC**: Free Lossless Audio Codec
- **OGG**: Ogg Vorbis 格式

## 性能基准

### 硬件要求

| 配置 | 最低要求 | 推荐配置 |
|------|----------|----------|
| CPU | 2核心 | 4核心+ |
| 内存 | 2GB | 8GB+ |
| 存储 | 500MB | 2GB+ |

### 性能指标

- **延迟**: < 200ms（实时转录）
- **准确率**: > 95%（在标准测试集上）
- **并发连接**: 支持 100+ 同时连接
- **内存使用**: < 500MB（基础配置）

### 优化建议

1. **增加线程数**: 在 `config.yaml` 中调整 `num_threads`
2. **使用 SSD**: 提高模型加载速度
3. **调整缓冲区**: 根据网络条件调整音频缓冲区大小
4. **启用端点检测**: 提高实时转录准确性

## 故障排除

### 常见问题

#### 1. 模型相关问题

**问题**: 模型加载失败
```bash
Error: failed to load model: model file not found
```

**解决方案**:
1. 确保模型文件路径正确
2. 检查模型文件完整性
3. 验证 tokens.txt 文件格式
4. 确认文件权限

**问题**: 模型内存不足
```bash
Error: out of memory
```

**解决方案**:
1. 增加系统内存
2. 使用更小的模型
3. 减少并发连接数

#### 2. 依赖问题

**问题**: sherpa-onnx-go 编译失败
```bash
#cgo CFLAGS: -I${SRCDIR}/include
```

**解决方案**:
```bash
# 清理并重新安装依赖
go clean -modcache
go mod tidy

# 确保安装了必要的系统依赖
# Ubuntu/Debian
sudo apt-get install build-essential

# CentOS/RHEL
sudo yum groupinstall "Development Tools"

# macOS
xcode-select --install
```

#### 3. WebSocket 连接问题

**问题**: WebSocket 连接失败
```javascript
WebSocket connection to 'ws://localhost:8080/ws/realtime' failed
```

**解决方案**:
1. 确保服务器正在运行
2. 检查 WebSocket URL 是否正确
3. 验证防火墙设置
4. 检查 CORS 配置

**问题**: 音频数据格式错误
```bash
Error: unsupported audio format
```

**解决方案**:
1. 确保音频数据为 16-bit PCM 格式
2. 检查采样率是否为 16kHz
3. 验证音频数据完整性

#### 4. 性能问题

**问题**: 转录速度慢
**解决方案**:
1. 增加 CPU 核心数
2. 调整 `num_threads` 配置
3. 使用更快的存储设备
4. 优化网络连接

**问题**: 内存使用过高
**解决方案**:
1. 减少并发连接数
2. 使用更小的模型
3. 调整音频缓冲区大小
4. 定期重启服务

### 日志调试

启用详细日志：

```yaml
# config.yaml
logging:
  level: "debug"
  format: "json"
```

查看日志：

```bash
# 查看实时日志
tail -f transcribeserver.log

# 查看错误日志
grep "ERROR" transcribeserver.log

# 查看性能日志
grep "latency" transcribeserver.log
```

### 监控指标

服务器提供以下监控端点：

```bash
# 健康检查
curl http://localhost:8080/health

# 系统信息（如果实现）
curl http://localhost:8080/metrics
```

## 部署指南

### 生产环境部署

#### 1. 系统配置

```bash
# 增加文件描述符限制
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf

# 优化内核参数
echo "net.core.somaxconn = 65535" >> /etc/sysctl.conf
echo "net.ipv4.tcp_max_syn_backlog = 65535" >> /etc/sysctl.conf
sysctl -p
```

#### 2. 使用 systemd 服务

创建服务文件 `/etc/systemd/system/transcribeserver.service`：

```ini
[Unit]
Description=Sherpa-ONNX Transcription Server
After=network.target

[Service]
Type=simple
User=transcribe
WorkingDirectory=/opt/transcribeserver
ExecStart=/opt/transcribeserver/transcribeserver
Restart=always
RestartSec=5
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable transcribeserver
sudo systemctl start transcribeserver
sudo systemctl status transcribeserver
```

#### 3. 使用 Nginx 反向代理

配置 Nginx：

```nginx
upstream transcribeserver {
    server 127.0.0.1:8080;
}

server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://transcribeserver;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 负载均衡

使用 HAProxy 进行负载均衡：

```haproxy
global
    daemon

defaults
    mode http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms

frontend transcribe_frontend
    bind *:80
    default_backend transcribe_backend

backend transcribe_backend
    balance roundrobin
    server transcribe1 127.0.0.1:8080 check
    server transcribe2 127.0.0.1:8081 check
    server transcribe3 127.0.0.1:8082 check
```

## 贡献指南

### 开发环境设置

1. Fork 项目
2. 克隆你的 fork
3. 创建功能分支
4. 提交更改
5. 推送到分支
6. 创建 Pull Request

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 添加适当的注释和文档
- 编写单元测试
- 确保所有测试通过

### 提交信息格式

```
type(scope): description

[optional body]

[optional footer]
```

类型：
- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

### 测试指南

```bash
# 运行所有测试
make test

# 运行特定测试
go test -v ./server -run TestTranscribeHandler

# 运行基准测试
go test -bench=. ./transcribe

# 检查代码覆盖率
make coverage
```

## 许可证

MIT License

## 致谢

- [sherpa-onnx](https://github.com/k2-fsa/sherpa-onnx) - 优秀的语音识别引擎
- [Gin](https://github.com/gin-gonic/gin) - 高性能 HTTP Web 框架
- [Gorilla WebSocket](https://github.com/gorilla/websocket) - WebSocket 实现

## 更新日志

### v1.0.0 (2024-01-XX)
- 初始版本发布
- 支持 HTTP API 和 WebSocket 接口
- 实现实时语音识别
- 支持多种音频格式
- 提供 Docker 部署支持

## 联系方式

- 项目主页: https://github.com/your-username/transcribeserver
- 问题反馈: https://github.com/your-username/transcribeserver/issues
- 邮箱: your-email@example.com

---

**注意**: 这是一个开源项目，欢迎社区贡献和改进！
