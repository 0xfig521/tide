# Release & Deploy

## 版本号规范

采用语义化版本 `vMAJOR.MINOR.PATCH`：

- `v0.1.x` — bug fix
- `v0.2.0` — new feature (backward compatible)
- `v1.0.0` — stable

## 发版流程

### 1. 确保代码就绪

```bash
git checkout main
git pull
go test ./...        # 测试通过
go vet ./...         # 静态检查
```

### 2. 打 tag 并推送

```bash
git tag v0.2.0 -m "v0.2.0 — <简短描述>"
git push origin v0.2.0
```

推送 tag 后，GitHub Actions 自动触发：

| 步骤 | 说明 |
|------|------|
| 构建 | 交叉编译 `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64` |
| 打包 | tar.gz 归档 + checksums.txt |
| Release | 创建 GitHub Release，上传二进制 |
| Brew | 自动更新 `0xfig521/homebrew-tap` formula |

### 3. 验证

```bash
# 二进制下载
curl -fsSL https://raw.githubusercontent.com/0xfig521/tide/main/install.sh | bash

# Homebrew
brew update && brew upgrade tide
tide --version
```

## 分发渠道

| 渠道 | 用户命令 | 自动？ |
|------|----------|--------|
| curl 脚本 | `curl .../install.sh \| bash` | ✅ |
| Homebrew | `brew install 0xfig521/tap/tide` | ✅ |
| Go install | `go install github.com/0xfig521/tide@latest` | ✅ 源码编译 |

## CI/CD 配置

### GitHub Actions（`.github/workflows/release.yml`）

触发条件：`v*` tag push

依赖 Secret：

| Secret | 用途 |
|--------|------|
| `GITHUB_TOKEN` | 自动提供，创建 Release |
| `TAP_GITHUB_TOKEN` | 需手动配置，更新 homebrew-tap |

### GoReleaser（`.goreleaser.yml`）

- 构建目标：`linux/darwin` × `amd64/arm64`
- CGO 关闭，纯静态二进制
- 版本号通过 ldflags 注入：`-X github.com/0xfig521/tide/cmd.version={{.Version}}`

## 创建 TAP_GITHUB_TOKEN

如果 Secret 过期，重新生成：

1. https://github.com/settings/tokens → Generate new token (classic)
2. 勾选 `repo` 权限
3. https://github.com/0xfig521/tide/settings/secrets/actions → `TAP_GITHUB_TOKEN`

## 回滚

删除 GitHub Release 和 tag：

```bash
# 删除线上 release
gh release delete v0.2.0 --repo 0xfig521/tide --yes

# 删除 tag
git tag -d v0.2.0
git push origin :refs/tags/v0.2.0

# Homebrew tap 需手动回退 formula
cd /path/to/homebrew-tap
git revert HEAD
git push
```
