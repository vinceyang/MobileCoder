# Git 提交规范

## 规则

1. **所有代码必须先提交到 main 分支**
2. 创建新分支进行开发
3. 开发完成后合并到 main
4. 再创建 tag 用于发布

## 工作流程

```bash
# 1. 确保在 main 分支
git checkout main

# 2. 拉取最新代码
git pull origin main

# 3. 创建开发分支（可选）
git checkout -b feature/xxx

# 4. 开发并提交...

# 5. 切换回 main 并合并
git checkout main
git merge feature/xxx

# 6. 推送到远程
git push origin main

# 7. 创建 tag
git tag -a v1.0 -m "Release v1.0"
git push origin v1.0
```

## 注意事项

- **禁止直接推送到 tag**
- **必须先 main 后 tag**
