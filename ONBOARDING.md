# Scribble Team Onboarding Guide

Welcome to the Scribble project! This guide explains how our team of 2 developers will work together efficiently on this Discord LeetCode Activity.

---

## Team Structure

- **Senior Developer**: Project scaffolding, architecture, code reviews, unblocking issues
- **Developer 1**: Frontend focus (React, UI components, Discord SDK integration)
- **Developer 2**: Backend focus (Go server, code execution, Kubernetes)

Both developers can work in **parallel**, but we follow a structured workflow to avoid conflicts and ensure code quality.

---

## The Workflow: Pull → Claim → Lock → Work → Push → Mark Complete

This is critical for parallel development. Every developer must follow this exact sequence.

**⚠️ CRITICAL:** Do NOT skip the "Lock" step. This prevents both developers from claiming the same task.

### Step 1: Pull Latest Changes

Before starting work on an issue, pull the latest code from main:

```bash
cd /home/nahtao97/projects/scribble
git fetch origin
git pull origin main
bd sync --from-main
```

This ensures you have the latest work from the other developer.

### Step 2: Claim Your Issue (Lock It Immediately!)

This is different from your old workflow - you MUST lock the claim globally:

```bash
# See what's available
bd ready

# Claim the issue in beads
bd update scribble-ms2 --status=in_progress

# CRITICAL: Push the claim lock immediately (before creating branch!)
git add .beads/
git commit -m "chore: claim scribble-ms2"
git push origin main
```

**Why this order matters:** When you push `.beads/` to main, the other developer's `bd ready` will show the task is already claimed. Without this push, both of you could claim the same task.

### Step 3: Work on Your Issue

After claiming and locking:

```bash
# Create a feature branch with the issue ID
git checkout -b dev1/scribble-ms2-init-nodejs-server

# Do your work (code, test locally, commit regularly)
git add .
git commit -m "feat(nodejs): initialize express server with health endpoint"
```

**Important:**
- Use feature branches with pattern: `dev1/scribble-XXX-short-description` or `dev2/...`
- Make atomic commits (one logical change per commit)
- Push small, focused changes
- Test everything locally before pushing

### Step 4: Push Your Changes

When your work is ready (tested locally):

```bash
# Push your feature branch
git push origin dev1/scribble-ms2-init-nodejs-server

# Create a pull request
gh pr create --title "Initialize Node.js Express server" \
  --body "Closes scribble-ms2"
```

**Wait for the other developer to pull your changes** - they won't have your code until they run `git pull`.

### Step 5: Mark Issue Complete in Beads

Only after your code is pushed and merged:

```bash
# Verify your changes are merged to main
git log --oneline -5

# Mark issue as complete in beads
bd close scribble-ms2 --reason="Code pushed and tested. Ready for integration."
```

This unlocks dependent issues in the workflow.

---

## Daily Workflow Example

**Developer 1 (Frontend) Working on Issue scribble-spl:**

```bash
# 9:00 AM - Start work
git fetch origin && git pull origin main
bd sync --from-main
bd ready

# CLAIM and LOCK
bd update scribble-spl --status=in_progress
git add .beads/ && git commit -m "chore: claim scribble-spl" && git push origin main

# Create branch and work
git checkout -b dev1/scribble-spl-init-react-vite
# ... create files, test locally ...

# 12:00 PM - Push changes
git push origin dev1/scribble-spl-init-react-vite
gh pr create --title "Initialize React + Vite frontend" --body "Closes scribble-spl"

# Wait for review from Senior Dev
# 12:30 PM - PR approved and merged to main

# Mark complete
git pull origin main  # Verify merge locally
bd close scribble-spl --reason="React + Vite setup complete, tested locally"
```

**Developer 2 (Backend) Working on Issue scribble-5n7 (in parallel):**

```bash
# 9:00 AM - Start work
git fetch origin && git pull origin main
bd sync --from-main
bd ready

# CLAIM and LOCK
bd update scribble-5n7 --status=in_progress
git add .beads/ && git commit -m "chore: claim scribble-5n7" && git push origin main

# Create branch and work
git checkout -b dev2/scribble-5n7-init-go-backend
# ... create files, test locally ...

# 1:00 PM - Push changes
git push origin dev2/scribble-5n7-init-go-backend
gh pr create --title "Initialize Go backend server" --body "Closes scribble-5n7"

# Wait for review from Senior Dev
# 1:30 PM - PR approved and merged to main

# Mark complete
git pull origin main
bd close scribble-5n7 --reason="Go server setup complete, tested"
```

**Next Morning:**

```bash
# Developer 1 starts next issue
git fetch origin && git pull origin main  # <-- Gets Developer 2's changes!
bd ready  # See what's unblocked after closing previous issue
```

---

## Git Workflow Rules

### Branch Naming (CRITICAL for conflict prevention)

**Format:** `dev{1|2}/scribble-{ISSUE_ID}-{short-description}`

**Developer 1 (Frontend):**
```
dev1/scribble-ms2-init-nodejs-server
dev1/scribble-spl-init-react-vite
dev1/scribble-eg8-problem-ui
```

**Developer 2 (Backend):**
```
dev2/scribble-5n7-init-go-backend
dev2/scribble-kb9-postgres-schema
dev2/scribble-efh-execute-endpoint
```

**Why separate branches by developer?**
- If you both edit the same file, your separate branches mean you review changes together in a merge request
- Prevents accidentally overwriting each other's work
- Makes it clear who's responsible for each feature

### Commit Messages

Use Conventional Commits format:

```
type(scope): subject

body (optional)
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `refactor`: Code restructuring (no feature change)
- `docs`: Documentation only
- `test`: Test additions/changes
- `chore`: Dependencies, configs, setup

**Examples:**
```
feat(nodejs): initialize express server with cors and dotenv
feat(react): setup vite with tailwind and shadcn/ui
feat(go): initialize gin server with health check endpoint
fix(frontend): correct discord sdk initialization flow
docs(readme): update architecture diagram
```

### Pull Requests

Every change should go through a PR (code review):

```bash
gh pr create \
  --title "Initialize Node.js Express server" \
  --body "Closes scribble-ms2

## What changed
- Created services/nodejs-proxy/ with Express.js setup
- Added health check endpoint at /api/health
- Configured TypeScript and CORS

## How to test
\`\`\`bash
cd services/nodejs-proxy
npm install
npm run dev
# Visit http://localhost:3000/api/health
\`\`\`

## Checklist
- [x] Code follows project style
- [x] Tested locally
- [x] No console errors
- [x] Ready for integration
"
```

---

## Beads Issue Workflow

### Find Work

```bash
# See what's ready to work on (no blockers)
bd ready

# Pick an issue and claim it
bd update scribble-ms2 --status=in_progress

# View full issue details
bd show scribble-ms2
```

### Track Progress

```bash
# See issues you're working on
bd list --status=in_progress

# See what's blocking other developers
bd blocked

# Check overall project progress
bd stats
```

### Mark Complete

```bash
# Close issue after code is merged
bd close scribble-ms2 --reason="Code tested and merged to main"

# This automatically unlocks dependent issues!
bd ready  # Now shows new available work
```

---

## Syncing Between Developers

### Before You Start Work

**ALWAYS do this first:**

```bash
git fetch origin
git pull origin main
bd sync --from-main  # Sync beads with main branch
```

This ensures you have:
- Latest code from the other developer
- Latest issue status updates

### After You Finish Work

**ALWAYS do this last:**

```bash
# 1. Push your code
git push origin your-branch
gh pr create ...

# 2. Wait for approval and merge (senior dev reviews)

# 3. Pull the merged changes
git pull origin main

# 4. Sync beads with main
bd sync --from-main

# 5. Mark issue complete
bd close scribble-xxx --reason="Merged and tested"
```

### Conflict Resolution

If you get a git conflict:

```bash
git pull origin main  # This will show merge conflicts

# Edit the conflicted files, keep the parts you need
# Then resolve:
git add .
git commit -m "fix: resolve merge conflicts with main"
git push origin your-branch

# Tell the other developer and senior dev about the conflict
# They can review your conflict resolution
```

---

## Development Environment Setup

### Prerequisites

Each developer needs:
- Docker Desktop with Kubernetes enabled
- Node.js 20+
- Go 1.21+
- kubectl configured for your home K8s cluster
- GitHub CLI (`gh`)
- git configured with your name and email

### Local Development Setup (Do This Once)

```bash
# Clone the repository
git clone https://github.com/yourusername/scribble.git
cd scribble

# Copy environment template
cp .env.example .env

# Edit .env with your settings (Discord credentials, etc.)
nano .env

# Start local services with Docker Compose
docker-compose up -d

# Verify everything is running
docker-compose ps
```

### Before Each Work Session

```bash
# Pull latest changes
git fetch origin && git pull origin main
bd sync --from-main

# Start local services
docker-compose up -d

# Check logs
docker-compose logs -f
```

---

## Communication & Coordination

### Daily Stand-up (15 min)

**Time**: 9:00 AM daily

**What to share:**
- What you're working on (issue ID)
- What you finished yesterday
- Any blockers you need help with
- Any changes that might affect the other developer

**Example:**
```
Developer 1: "Working on scribble-spl (React setup), should be done by noon.
             Will push changes so you can integrate."

Developer 2: "Working on scribble-5n7 (Go server), no blockers.
             Ready to work on database next once you finish."

Senior: "Both look good, I'll review PRs as they come in."
```

### Slack/Chat Channel

Post when:
- You push a PR (notify other developer to pull changes)
- You encounter a blocker
- You find an issue with another developer's code
- You finish an epic (unlock new work for them)

### Code Review

Senior developer reviews all PRs:
- **Target time**: 30 minutes
- **Before marking complete**: Get approval from senior dev
- **If changes requested**: Fix in same PR, push again, ping senior dev

---

## Common Scenarios

### Scenario 1: Developer 1 Finishes First, Unblocks Developer 2

```
Dev 1: Finishes scribble-spl (React setup)
Dev 1: git push → PR → Senior reviews → Merged
Dev 1: bd close scribble-spl

bd ready now shows scribble-rel (Discord SDK integration)
which depends on scribble-spl

Dev 2: git pull origin main  # Gets Dev 1's React setup
Dev 2: bd ready  # Sees scribble-rel is ready
Dev 2: bd update scribble-rel --status=in_progress
```

### Scenario 2: Merge Conflict

```
Dev 1: Pushes changes to services/nodejs-proxy/src/server.ts
Dev 2: Also modifying the same file locally

Dev 2: git pull origin main
# Git detects conflict in services/nodejs-proxy/src/server.ts

Dev 2: Edits the conflicted file, resolves manually
Dev 2: git add services/nodejs-proxy/src/server.ts
Dev 2: git commit -m "fix: resolve merge conflicts"
Dev 2: git push origin your-branch

Senior Dev: Reviews PR, approves the conflict resolution
```

### Scenario 3: Senior Dev Needs to Scaffold Code

```
Senior Dev: "I'm scaffolding Epic 2 database files,
           you can pull those changes and build on top"

Senior Dev: Creates database/schema.sql, database/seed.sql
Senior Dev: git push origin main

Dev 2: git pull origin main
Dev 2: Now sees database structure, can start
       implementing Go handlers against it
```

---

## Troubleshooting

### "I can't see the other developer's changes"

```bash
# Pull latest
git fetch origin && git pull origin main

# Sync beads
bd sync --from-main

# Restart local services
docker-compose restart
```

### "I'm getting merge conflicts constantly"

- Pull more frequently (every 2 hours)
- Communicate with the other developer about which files you're modifying
- Keep branches short-lived (merge within 2-4 hours)

### "An issue is marked blocked but I think it's ready"

```bash
# Check what's blocking it
bd show scribble-xxx

# If the blocking issue is actually done, ask senior dev to
# manually unblock it:
bd update scribble-xxx --status=open
```

### "I pushed code but the other developer doesn't see it"

Remind them to:
```bash
git fetch origin && git pull origin main
```

Files are only on their machine after pulling.

---

## Weekly Checklist

Every Friday:

- [ ] All issues you closed have code merged to main
- [ ] Run `bd stats` and note progress
- [ ] Pull latest changes: `git fetch origin && git pull origin main`
- [ ] Sync beads: `bd sync --from-main`
- [ ] Next week's work is visible with `bd ready`
- [ ] No outstanding PRs waiting for review
- [ ] No merge conflicts in progress

---

## Success Criteria

You're working effectively together when:

✅ Both developers have something to work on every day (no blocking)
✅ Code merges happen at least once per day per developer
✅ Issues complete and move to next phase smoothly
✅ No surprises - both developers know what the other is working on
✅ Pull requests reviewed and merged within 30 minutes
✅ No accidental git conflicts or lost work

---

## Senior Developer's Role

As the senior, I will:

1. **Review all PRs** within 30 minutes
2. **Scaffold complex components** (database schemas, architecture setup)
3. **Unblock you** if dependencies aren't working
4. **Monitor progress** with weekly status checks
5. **Guide decisions** on architecture or implementation approach
6. **Help with conflicts** (git or technical)

**You can always ping me if stuck** - that's what I'm here for.

---

## Git Help & Resources

**New to Git?** Read these in order:
1. `/GIT_CHEATSHEET.md` - Quick one-page reference (print this!)
2. `/GIT_WORKFLOW.md` - Complete step-by-step guide for all scenarios
3. `/QUICKSTART.md` - Fast setup guide

**Key Git Rules:**
- Always use `dev1/` or `dev2/` branch prefix (prevents conflicts!)
- Commit frequently with message format: `type(scope): description`
- Test before pushing: `npm run lint` or `go test`
- Wait for code review - never merge your own code
- Pull every morning: `git fetch origin && git pull origin main`

---

## Questions?

If anything is unclear:
1. **Check the docs** - Read GIT_WORKFLOW.md or GIT_CHEATSHEET.md first
2. **Ask in standup** - Daily team meeting at 9 AM
3. **Post in Slack** - For quick questions
4. **Email senior dev** - For architecture/complex issues

**Git emergency?** Don't panic! Senior developer can help fix any git issue.

Let's build something great together! 🚀
