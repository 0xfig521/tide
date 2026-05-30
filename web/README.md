# Tide Landing Page

Tide CLI 的官方网站，托管于 Cloudflare Pages。

- **域名**: https://tide.0xfig.xyz
- **技术栈**: React 19 + Vite 8 + TypeScript 6 + Tailwind CSS 4 + GSAP
- **包管理**: bun

## 本地开发

```bash
cd web
bun install
bun run dev
```

## 构建预览

```bash
bun run build     # 输出到 dist/
bun run preview   # 本地预览构建产物
```

## 部署

### 自动部署（推荐）

推送代码到 `main` 分支，且变更路径在 `web/**` 时，GitHub Action 自动触发：

```bash
git add web/
git commit -m "..."
git push origin main
```

可在 GitHub Actions 页面手动触发 **Deploy Web** workflow（`workflow_dispatch`）。

### 手动部署

```bash
# 本地构建后通过 wrangler 部署
cd web
bun install --frozen-lockfile
bun run build
bunx wrangler pages deploy dist --project-name=tide
```

### 所需 GitHub Secrets

| Secret | 说明 |
|---|---|
| `CLOUDFLARE_API_TOKEN` | Cloudflare API Token，权限为 Cloudflare Pages: Edit |
| `CLOUDFLARE_ACCOUNT_ID` | Cloudflare 账号 ID |

### 域名绑定

域名 `tide.0xfig.xyz` 在 Cloudflare Dashboard 中绑定：
Workers & Pages → tide → Custom domains → 添加 `tide.0xfig.xyz`

SPA 客户端路由由 `public/_redirects`（`/* /index.html 200`）处理。

### CI 流程

定义在 `.github/workflows/deploy-web.yml`：

1. `actions/checkout@v4`
2. `oven-sh/setup-bun@v2` 安装最新 bun
3. `bun install --frozen-lockfile`
4. `bun run build`（tsc + vite build）
5. `bunx wrangler pages deploy dist --project-name=tide`
