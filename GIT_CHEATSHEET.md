# Git Cheat Sheet - Scribble Project

**Print this page or bookmark it!**

---

## 🚀 Start Your Day

```bash
cd /home/nahtao97/projects/scribble
git fetch origin && git pull origin main
bd sync --from-main
```

---

## 📌 Create Your Feature Branch

**Developer 1:**
```bash
bd update scribble-XXX --status=in_progress
git checkout -b dev1/scribble-XXX-short-name
```

**Developer 2:**
```bash
bd update scribble-XXX --status=in_progress
git checkout -b dev2/scribble-XXX-short-name
```

---

## 💾 Save Your Work (Commit)

```bash
# See what changed
git status

# Add all your changes
git add .

# Create a commit
git commit -m "type(scope): description"

# Examples:
# git commit -m "feat(frontend): add header component"
# git commit -m "fix(backend): correct database query"
# git commit -m "docs(readme): update setup instructions"
```

**Commit Types:**
- `feat` - New feature
- `fix` - Bug fix
- `refactor` - Code cleanup
- `docs` - Documentation
- `test` - Tests
- `chore` - Setup/config

---

## ✅ Test Your Code

**Frontend:**
```bash
cd frontend
npm run lint
npm run type-check
npm run dev
```

**Backend:**
```bash
cd services/go-backend
go test ./...
go run cmd/server/main.go
```

---

## 🎯 Push Your Changes

```bash
git push origin dev1/scribble-XXX-short-name
```

---

## 📤 Create a Merge Request

```bash
gh pr create \
  --title "Your feature title" \
  --body "Closes scribble-XXX

## What changed
- What you did

## How to test
- How to verify it works

## Checklist
- [x] Tested locally
- [x] No console errors
- [x] Code follows style
"
```

---

## ✨ After Your Code is Merged

```bash
git checkout main
git pull origin main
bd close scribble-XXX --reason="Merged and tested"
git branch -d dev1/scribble-XXX-short-name
```

---

## 🔍 Check Status Anytime

```bash
git status           # What changed?
git branch           # What branch am I on?
git log --oneline    # What commits did I make?
```

---

## 🚨 If Something Goes Wrong

### "I committed but need to change my message"
```bash
git commit --amend -m "new message"
git push origin -f dev1/feature  # Force push only to YOUR branch!
```

### "I want to undo my last commit"
```bash
git reset --soft HEAD~1  # Keep changes
git reset --hard HEAD~1  # Delete changes
```

### "I want to see what I changed"
```bash
git diff                           # Unstaged changes
git diff --staged                  # Staged changes
git diff main...HEAD               # Your changes vs main
```

### "I have uncommitted changes and need to switch branches"
```bash
git stash              # Save changes temporarily
git checkout other-branch
git checkout your-branch
git stash pop          # Restore changes
```

### "Merge conflict! What do I do?"
```bash
# 1. Open the conflicted file in editor
# 2. Remove the <<<<<<, =======, >>>>>>> markers
# 3. Keep the code you want
git add .
git commit -m "fix: resolve merge conflict"
git push origin your-branch
```

---

## 📋 Quick Branch Guide

| Goal | Command |
|------|---------|
| See all branches | `git branch -a` |
| Create new branch | `git checkout -b dev1/feature-name` |
| Switch to branch | `git checkout branch-name` |
| Delete branch | `git branch -d branch-name` |
| Rename branch | `git branch -m old-name new-name` |

---

## 🔄 Pull vs Fetch vs Merge

| Command | What it does |
|---------|------------|
| `git fetch` | Download updates (no changes to files) |
| `git pull` | Download AND apply updates |
| `git merge` | Combine branches (usually done in merge request) |

---

## 📝 Commit Message Examples

✅ **GOOD:**
```
feat(frontend): add user profile page
fix(database): correct query for user submissions
refactor(backend): simplify code execution handler
docs(readme): update installation instructions
test(frontend): add unit tests for auth component
```

❌ **BAD:**
```
changes
Fixed stuff
WIP
URGENT!!!
updated code
misc fixes
```

---

## ⚡ Most Common Workflow

```bash
# 1. Start day
git fetch origin && git pull origin main

# 2. Pick issue
bd ready

# 3. Create branch
git checkout -b dev1/scribble-123-feature

# 4. Code!
# ... write code ...

# 5. Commit frequently
git add .
git commit -m "feat(scope): what you did"

# 6. Test
npm run lint && npm run dev  # or go test

# 7. Push
git push origin dev1/scribble-123-feature

# 8. Create merge request
gh pr create --title "..." --body "..."

# 9. Wait for approval...

# 10. After merged
git checkout main
git pull origin main
bd close scribble-123
```

---

## 🆘 When You're Stuck

1. **Check this cheat sheet** ← You're reading it!
2. **Run `git status`** - It usually tells you what to do
3. **Ask in Slack** - Quick help from team
4. **Daily standup** - Discuss with everyone
5. **Senior dev** - For complex issues

---

## 🎯 Remember These 3 Rules

1. **Always pull before starting:** `git pull origin main`
2. **Create YOUR branch:** `git checkout -b devX/...`
3. **Test before pushing:** `npm run lint` or `go test`

---

**Questions?** Read `/GIT_WORKFLOW.md` for full guide
**Confused?** Ask in Slack or standup
**Git emergency?** Senior developer can help fix anything!
