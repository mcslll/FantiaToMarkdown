## FantiaToMarkdown

Fantia (fantia.jp) 爬虫，用于下载 Fantia Fanclub 的作品投稿并保存为 Markdown 文件

**！！！该软件不能直接帮你免费爬取订阅后才能查看的内容！！！**

### 准备

使用浏览器插件 [Cookie Master](https://chromewebstore.google.com/detail/cookie-master/jahkihogapggenanjnlfdcbgmldngnfl) 导出 Fantia cookie，点击 `copy`

将复制到的 JSON 文本粘贴进与在 RELEASE 中下载的可执行文件同级（或 git clone 的项目根目录）的 `cookies.json` 即可

### 全局参数

| 参数       | 说明                                        | 默认值                          |
| ---------- | ------------------------------------------- | ------------------------------- |
| `--host`   | 主站域名                                    | `fantia.jp`                     |
| `--dir`    | 数据存储目录                                | 程序所在目录下的 `data` 文件夹  |
| `--cookie` | cookies.json 文件路径                       | 程序所在目录下的 `cookies.json` |
| `--proxy`  | 代理服务器地址 (如 `http://127.0.0.1:7890`) | (空)                            |
| `--debug`  | 启用调试日志                                | `false`                         |

### 构建

如果你不需要对源码进行开发，请直接在 Release 页面下载编译好的版本

- go build

### 帮助

```shell
.\FantiaToMarkdown.exe -h
```

### 使用

本程序为命令行程序，需要在 `cmd`, `powershell` 或 `bash` 等 shell 中输入参数调用

#### 下载指定 Fanclub 的所有作品投稿

注：下文提到的 `fanclub_id` 为作者主页 URL 的最后一部分，如 `https://fantia.jp/fanclubs/12345` 中的 `12345`

```shell
.\FantiaToMarkdown.exe fanclub --id "fanclub_id"
```

#### 按标签过滤下载

如果你只想下载特定标签的作品投稿：

```shell
.\FantiaToMarkdown.exe fanclub --id "fanclub_id" --tag "标签名"
```

### 特点

1. **自动命名目录**：程序会自动获取 Fanclub 的名称并将其作为文件夹名称（如果获取失败则使用 ID）
2. **文件名规范**：文件名格式为 `[作品ID]_[日期]_[标题].md`，便于排序和增量更新
3. **增量下载**：程序会自动检测本地已存在的文件，如果发现该作品已下载过，将自动跳过，节省流量和时间
4. **图片备份**：自动下载作品中的所有图片（包括封面和内容中的图片）并保存到本地
5. **代理支持**：支持通过 `--proxy` 参数设置网络代理，解决部分地区的访问问题

### 更新日志

#### v1.0.0 (Initial Release)

1. **极速下载**：基于 Go 协程实现多图并发下载
2. **智能目录**：存储目录优先采用作者/Fanclub 名称
3. **代理增强**：支持通过 `--proxy` 设置代理且能自动补全协议
4. **增量备份**：自动查重并跳过已下载的作品
5. **格式规范**：作品统一转换为 Markdown 并本地化所有图片

### TODO

- [ ] **集成 Backnumber API**：研究通过 JSON API 获取特定方案和月份的作品列表
  - **接口**：`https://fantia.jp/api/v1/fanclub/backnumbers/monthly_contents/plan/{planId}/month/{month}`
  - **验证**：需从页面提取 `csrf-token` 并放入请求头的 `x-csrf-token`
  - **优势**：可直接获取结构化数据，避免解析 HTML 列表，进一步提升稳定性

### 参考

- [Fantia API Integration](https://deepwiki.com/suzumiyahifumi/Fantia-Downloader-tampermonkey/6-fantia-api-integration)

---

基于 [AfdianToMarkdown](https://github.com/PhiFever/AfdianToMarkdown) 的设计思路开发
