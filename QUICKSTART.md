# Scribble - Quick Start for Developers

Welcome to the Scribble team! Here's everything you need to get started.

## For Developer 1 (Frontend)

### Initial Setup (First Time Only)
```bash
cd /home/nahtao97/projects/scribble
cp .env.example .env
nano .env  # Add your DISCORD_CLIENT_ID
docker-compose up -d
```

### Start Your First Issue
```bash
bd ready                           # See what's available
bd show scribble-spl               # View "Initialize React + Vite frontend"

# IMPORTANT: Claim the task FIRST
bd update scribble-spl --status=in_progress

# LOCK the claim immediately (prevent race conditions)
git add .beads/
git commit -m "chore: claim scribble-spl"
git push origin main

# NOW create your branch
git checkout -b dev1/scribble-spl-init-react-vite

# Start dev server
cd frontend
npm install
npm run dev
```

**You're in:** `frontend/` folder
- `src/App.tsx` - Main app component (already has Discord SDK setup!)
- `src/components/` - Your UI components go here
- `src/pages/` - Page-level components
- `package.json` - Already configured with React, Vite, Tailwind

### Daily Workflow
```bash
# Before starting
git fetch origin && git pull origin main
bd sync --from-main
bd ready

# Claim a task (locks it globally)
bd update scribble-XXX --status=in_progress
git add .beads/ && git commit -m "chore: claim scribble-XXX" && git push origin main

# Create your branch and work
git checkout -b dev1/scribble-XXX-description
npm run dev

# When done
git push origin dev1/scribble-XXX-description
gh pr create --title "..." --body "..."
# Wait for review, then:
git pull origin main
bd close scribble-XXX --reason="Merged"
```

**Need git help?** Read `/GIT_WORKFLOW.md` for step-by-step guide or `/GIT_CHEATSHEET.md` for quick reference

---

## For Developer 2 (Backend)

### Initial Setup (First Time Only)
```bash
cd /home/nahtao97/projects/scribble
cp .env.example .env
nano .env  # Update database credentials if needed
docker-compose up -d
```

### Start Your First Issue
```bash
bd ready                           # See what's available
bd show scribble-5n7               # View "Initialize Go backend server"

# IMPORTANT: Claim the task FIRST
bd update scribble-5n7 --status=in_progress

# LOCK the claim immediately (prevent race conditions)
git add .beads/
git commit -m "chore: claim scribble-5n7"
git push origin main

# NOW create your branch
git checkout -b dev2/scribble-5n7-init-go-backend

# Start dev server
cd services/go-backend
go mod download
go run cmd/server/main.go
```

**You're in:** `services/go-backend/` folder
- `cmd/server/main.go` - HTTP server entry point (already has basic setup!)
- `internal/handlers/` - HTTP request handlers
- `internal/services/` - Business logic
- `internal/db/` - Database code
- `internal/k8s/` - Kubernetes integration

### Database Access
```bash
# Connect to PostgreSQL
psql postgres://postgres:dev_password@localhost:5432/scribble_dev

# View schema
\dt
\dv  # views

# Or check our SQL files
cat database/schema.sql
cat database/seed.sql
```

### Daily Workflow
```bash
# Before starting
git fetch origin && git pull origin main
bd sync --from-main
bd ready

# Claim a task (locks it globally)
bd update scribble-XXX --status=in_progress
git add .beads/ && git commit -m "chore: claim scribble-XXX" && git push origin main

# Create your branch and work
git checkout -b dev2/scribble-XXX-description
go run cmd/server/main.go

# When done
git push origin dev2/scribble-XXX-description
gh pr create --title "..." --body "..."
# Wait for review, then:
git pull origin main
bd close scribble-XXX --reason="Merged"
```

---

## Check Service Status

```bash
# See all running services
docker-compose ps

# View logs (all services)
docker-compose logs -f

# View logs (specific service)
docker-compose logs -f frontend          # Vite dev server
docker-compose logs -f nodejs-proxy      # Node.js on port 3000
docker-compose logs -f go-backend        # Go backend on port 8080
docker-compose logs -f postgres          # Database on port 5432

# Restart a service if it crashes
docker-compose restart go-backend
```

---

## Your First Epic

### Issues You'll Work On (In Order)

**Frontend (Developer 1):**
1. scribble-ms2 - Initialize Node.js proxy (short support)
2. scribble-spl - Initialize React + Vite ✓
3. scribble-rel - Discord SDK integration ✓
4. scribble-eg8 - Problem UI with Monaco Editor
5. scribble-55t - Hardcoded submission flow
6. scribble-ey4 - Deploy to Docker & test

**Backend (Developer 2):**
1. scribble-5n7 - Initialize Go server ✓
2. scribble-kb9 - PostgreSQL schema ✓
3. scribble-at1 - Seed database ✓
4. scribble-j80 - Problem API endpoints
5. (And more, but start here!)

### Goal: Complete Epic 1 POC

**By End of Week 1:**
- Discord Activity loads in Discord
- User authenticates with Discord
- Can see hardcoded problem
- Can submit "code" and see fake result
- **Success:** Full proof of concept working end-to-end

---

## Quick Reference

### Check What's Blocking You
```bash
bd blocked         # See issues you're waiting on
bd show scribble-XXX  # See dependencies
```

### When You're Stuck
1. Read the issue description: `bd show scribble-XXX`
2. Check DEVELOPMENT.md in the repo
3. Ask in standup or Slack
4. Reach out to senior dev directly

### Code Review
- PR reviews happen within 30 minutes
- Keep PRs focused (one feature per branch)
- Tests should pass before submitting

### Before Committing
```bash
# Frontend
npm run lint
npm run type-check

# Backend
go test ./...
go fmt ./...
```

---

## Important Files You'll Use

**Documentation:**
- `/README.md` - Project overview
- `/ONBOARDING.md` - Detailed workflow & git guide
- `/DEVELOPMENT.md` - Full dev setup & troubleshooting
- `/QUICKSTART.md` - This file!

**Configuration:**
- `.env.example` - Environment template (copy to `.env`)
- `docker-compose.yml` - Local dev stack
- `.gitignore` - What gets committed

**Issue Tracker:**
- `bd ready` - Tasks ready to work
- `bd list` - All tasks
- `bd stats` - Progress overview

---

## Git Cheat Sheet

```bash
# Before you start work
git pull origin main
bd sync --from-main

# Create your branch
git checkout -b scribble-XXX-short-description

# Make commits (keep them small!)
git commit -m "feat(scope): what you did"

# Push when ready for review
git push origin scribble-XXX-short-description
gh pr create --title "Feature name" --body "Closes scribble-XXX"

# After PR is merged
git pull origin main
bd close scribble-XXX --reason="Merged and tested"
```

---

## Slack/Chat Notifications

Post when:
- ✅ You finish and push a PR (tell other dev to pull)
- 🚨 You're blocked on something
- ✨ You find something cool to share
- 🎉 You finish an issue

---

## Questions?

Ask in order:
1. DEVELOPMENT.md - 80% of questions answered here
2. Standup meeting - Quickest answer
3. Slack/chat - Async help from senior dev
4. Senior dev directly - For architecture questions

---

## This Week's Goals

**Developer 1 (Frontend):**
- [ ] Day 1: React setup working locally
- [ ] Day 1-2: Discord SDK initialization
- [ ] Day 2-3: Problem UI with Monaco Editor
- [ ] Day 3-4: Mock submission flow
- [ ] Day 4-5: Docker & Discord testing
- [ ] EOW: Epic 1 POC complete ✓

**Developer 2 (Backend):**
- [ ] Day 1: Go server running
- [ ] Day 1-2: PostgreSQL connected
- [ ] Day 2-3: Database queries working
- [ ] Day 3-4: Basic API endpoints
- [ ] Day 4-5: Docker testing
- [ ] EOW: Ready for Dev 1 integration ✓

**Both:**
- Daily standup at 9:00 AM
- Pull before you start work
- Push code by end of day
- Mark issues complete when merged

---

## You've Got This! 🚀

The scaffold is ready. Your issue tracker is set up. The team knows what to do.

Time to build something awesome!

---

**Last Updated:** December 31, 2025
**Status:** Ready for Development
**Contact:** Senior Developer / Slack
