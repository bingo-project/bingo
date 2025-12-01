# Google Search Console 提交指南

## 步骤 1：访问 Google Search Console

打开：https://search.google.com/search-console

## 步骤 2：添加资源

1. 如果已经添加过 `bingoctl.dev`，直接跳到步骤 3
2. 如果没有，点击「添加资源」
3. 选择「网域」方式
4. 输入：`bingoctl.dev`
5. 按照指示完成 DNS 验证

## 步骤 3：提交 Sitemap

1. 在左侧菜单选择「站点地图」
2. 在「添加新的站点地图」输入框中输入：`sitemap.xml`
3. 点击「提交」
4. 等待 Google 抓取（通常需要几天时间）

## 步骤 4：请求编入索引（重要！）

对以下重要页面使用「网址检查」工具手动请求索引：

### 中文页面（优先）
1. `https://bingoctl.dev/`
2. `https://bingoctl.dev/guide/what-is-bingo`
3. `https://bingoctl.dev/guide/getting-started`
4. `https://bingoctl.dev/essentials/architecture`
5. `https://bingoctl.dev/guide/using-bingo`

### 英文页面
1. `https://bingoctl.dev/en/`
2. `https://bingoctl.dev/en/guide/what-is-bingo`
3. `https://bingoctl.dev/en/guide/getting-started`
4. `https://bingoctl.dev/en/essentials/architecture`

### 如何请求索引：
1. 在顶部搜索框输入完整 URL
2. 点击「测试实际 URL」
3. 等待测试完成
4. 点击「请求编入索引」
5. 等待确认

**注意**：每个页面请求索引后需要等待几分钟，每天有配额限制（约 10-20 个）

## 步骤 5：监控收录状态

### 检查索引状态
1. 在左侧菜单选择「覆盖率」
2. 查看「有效」、「错误」、「警告」的页面数
3. 确保主要页面都在「有效」列表中

### 搜索测试
定期使用以下命令测试：
```
site:bingoctl.dev
```

### 关键词测试
等待 1-2 周后，尝试搜索：
- `bingo go 框架`
- `bingoctl`
- `go 微服务脚手架 bingo`
- `golang backend scaffold bingo`

## 常见问题

### Q: 提交后多久能被索引？
A: 通常 1-7 天，快的话 1-2 天就能看到部分页面被索引

### Q: 为什么有些页面没被索引？
A: 检查以下几点：
1. robots.txt 是否允许抓取
2. sitemap.xml 是否包含该页面
3. 页面是否有 noindex 标签
4. 内容质量是否足够（太短或重复内容可能不被索引）

### Q: 如何加快索引速度？
A:
1. 手动请求索引（最有效）
2. 获取外部链接（GitHub、技术博客等）
3. 保持内容更新
4. 提高内容质量

## 自动化脚本（可选）

如果需要批量提交所有 URL，可以使用 Google Indexing API：
https://developers.google.com/search/apis/indexing-api/v3/quickstart

但对于文档网站，手动提交重要页面 + sitemap 通常已经足够。
