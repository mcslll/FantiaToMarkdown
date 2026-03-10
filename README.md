## FantiaToMarkdown

Fantia (fantia.jp) 爬虫，用于下载 Fantia Fanclub 的作品投稿并保存为 Markdown 文件

**！！！该软件不能直接帮你免费爬取订阅后才能查看的内容！！！**

### 准备

使用浏览器插件 [Cookie Master](https://chromewebstore.google.com/detail/cookie-master/jahkihogapggenanjnlfdcbgmldngnfl) 导出 Fantia cookie，点击 `copy`

将复制到的 JSON 文本粘贴进与在 RELEASE 中下载的可执行文件同级（或 git clone 的项目根目录）的 `cookies.json` 即可

### 参数说明

| 长参数       | 短参数 | 说明                                        | 默认值                          |
| ------------ | ------ | ------------------------------------------- | ------------------------------- |
| `--host`     | -      | 主站域名                                    | `fantia.jp`                     |
| `--dir`      | `-d`   | 数据存储目录                                | 程序所在目录下的 `data` 文件夹  |
| `--cookie`   | `-c`   | cookies.json 文件路径                       | 程序所在目录下的 `cookies.json` |
| `--proxy`    | `-p`   | 代理服务器地址 (如 `http://127.0.0.1:7890`) | (空)                            |
| `--debug`    | `-D`   | 启用调试日志                                | `false`                         |
| `--log`      | `-l`   | 开启本地日志文件保存 (`crawler.log`)        | `false`                         |

#### fanclub 子命令参数

| 参数     | 短参数 | 说明                     | 必填 |
| -------- | ------ | ------------------------ | ---- |
| `--id`   | `-i`   | Fanclub ID               | 是   |
| `--tag`  | `-t`   | 根据标签过滤帖子         | 否   |

### 构建

如果你不需要对源码进行开发，请直接在 Release 页面下载编译好的版本

- go build

### 帮助

```shell
.\FantiaToMarkdown.exe -h
```

### 使用示例

本程序为命令行程序，需要在 `cmd`, `powershell` 或 `bash` 等 shell 中输入参数调用

#### 下载指定 Fanclub 的所有作品投稿

```shell
.\FantiaToMarkdown.exe fanclub -i "12345"
```

#### 按标签过滤并使用代理

```shell
.\FantiaToMarkdown.exe -p "http://127.0.0.1:7890" fanclub -i "12345" -t "xxx"
```

### 特点

1. **自动命名目录**：程序会自动获取 Fanclub 的名称并将其作为文件夹名称（如果获取失败则使用 ID）
2. **文件名规范**：文件名格式为 `[作品ID]_[日期]_[标题].md`，便于排序和增量更新
3. **增量下载**：程序会自动检测本地已存在的文件，如果发现该作品已下载过，将自动跳过
4. **图片备份**：自动下载作品中的所有图片并保存到本地 `.assets` 目录，且图片名包含 `postID` 防止重名冲突
5. **代理支持**：支持通过代理访问，并具备严格的错误检查防止真实 IP 泄露

### 更新日志

#### v1.1.0 (Current)

1. **进度显示**：使用 `post=当前/总数` 格式，直观展示整体进度。
2. **参数简写**：补全 `-i`, `-p`, `-D`, `-l`, `-d`, `-c`, `-t` 等短别名。
3. **下载优化**: 增加下载retry操作，限制并发数量

#### v1.0.0 (Initial Release)

1. **极速下载**：基于 Go 协程实现多图并发下载
2. **智能目录**：存储目录优先采用作者/Fanclub 名称
3. **稳定传输**：限制图片下载并发数并增加 3 次自动重试机制
4. **代理增强**：支持通过 `--proxy` 设置代理且能自动补全协议
5. **增量备份**：自动查重并跳过已下载的作品
6. **格式规范**：作品统一转换为 Markdown 并本地化所有图片

### TODO

- [ ] **集成 Backnumber API**：研究通过 JSON API 获取特定方案和月份的作品列表

### 参考

- [Fantia API Integration](https://deepwiki.com/suzumiyahifumi/Fantia-Downloader-tampermonkey/6-fantia-api-integration)

---

基于 [AfdianToMarkdown](https://github.com/PhiFever/AfdianToMarkdown) 的设计思路开发
