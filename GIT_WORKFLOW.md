# Git Workflow Guide - Scribble Team

This guide explains **exactly how to use git** for the Scribble project. Follow these procedures step-by-step. They're designed for developers new to git.

---

## Table of Contents

1. [Before You Start (First Time Only)](#before-you-start-first-time-only)
2. [Daily Start Procedure](#daily-start-procedure)
3. [Creating and Committing Code](#creating-and-committing-code)
4. [Creating a Branch for Your Feature](#creating-a-branch-for-your-feature)
5. [Pushing and Creating a Merge Request](#pushing-and-creating-a-merge-request)
6. [Handling Merge Conflicts](#handling-merge-conflicts)
7. [Common Commands Reference](#common-commands-reference)
8. [Git Rules for Scribble](#git-rules-for-scribble)

---

## Before You Start (First Time Only)

### Step 1: Configure Your Git Identity

Run these commands once to tell git who you are:

```bash
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
```

**Example:**
```bash
git config --global user.name "Alice Developer"
git config --global user.email "alice@example.com"
```

### Step 2: Verify Configuration

```bash
git config --global --list
```

You should see your name and email in the output.

### Step 3: Clone the Repository

```bash
git clone https://github.com/yourusername/scribble.git
cd scribble
```

Done! You now have the code locally.

---

## Daily Start Procedure

**DO THIS EVERY MORNING before you start work:**

### Step 1: Navigate to Project Directory

```bash
cd /home/nahtao97/projects/scribble
```

### Step 2: Update Your Local Code

Pull the latest changes from the main branch:

```bash
git fetch origin
git pull origin main
```

**What this does:**
- `git fetch origin` - Downloads latest changes from the server
- `git pull origin main` - Applies those changes to your local code

### Step 3: Sync Beads Issues

```bash
bd sync --from-main
```

This updates the issue tracker with the latest status.

### Step 4: Start Your Work

You're now ready to start! Check what's available:

```bash
bd ready
```

---

## Creating and Committing Code

### Step 1: Make Your Changes

Edit your files, write your code, add components, whatever the issue requires.

### Step 2: Check What Changed

Before committing, see what files you changed:

```bash
git status
```

**Output will look like:**
```
On branch main

Changes not staged for commit:
  modified:   frontend/src/App.tsx
  modified:   frontend/src/components/Header.tsx
  new file:   frontend/src/lib/api.ts

Untracked files:
  frontend/src/styles/new-styles.css
```

### Step 3: Review Your Changes (Optional but Recommended)

See exactly what changed in a file:

```bash
git diff frontend/src/App.tsx
```

This shows additions (+) and deletions (-).

### Step 4: Stage Your Changes

**Stage files you want to commit:**

```bash
# Stage specific file
git add frontend/src/App.tsx

# Stage all changes
git add .

# Stage multiple specific files
git add frontend/src/App.tsx frontend/src/lib/api.ts
```

**Verify what's staged:**

```bash
git status
```

You should see "Changes to be committed" in green.

### Step 5: Create a Commit

Commit is a "snapshot" of your changes with a message explaining what you did.

```bash
git commit -m "feat(frontend): add user authentication component"
```

**Format: `type(scope): description`**

**Allowed types:**
- `feat` - New feature
- `fix` - Bug fix
- `refactor` - Code cleanup (no feature change)
- `docs` - Documentation only
- `test` - Test additions
- `chore` - Dependencies, setup, config

**Scope:** What part changed (frontend, backend, database, etc.)

**Description:** What you did (short, lowercase, no period)

**Good Examples:**
```bash
git commit -m "feat(nodejs): initialize express server with cors"
git commit -m "feat(database): add submission history table"
git commit -m "fix(frontend): correct discord sdk initialization"
git commit -m "docs(readme): update architecture section"
```

**Bad Examples:**
```bash
git commit -m "changes"                    # Too vague
git commit -m "Fixed stuff"                # Capitalized, vague
git commit -m "URGENT FIX"                 # Shouty, not descriptive
git commit -m "WIP"                        # Work in progress, don't commit
```

### Step 6: Verify Your Commit

```bash
git log --oneline -5
```

You should see your new commit at the top.

---

## Claiming a Task (CRITICAL - No Race Conditions!)

**⚠️ CRITICAL: Follow this procedure exactly to prevent both developers claiming the same task.**

When two developers work on the same codebase, there's a risk of both claiming the same task simultaneously and not realizing until they commit. We prevent this with an immediate claim lock.

### The Safe Claiming Procedure (Do This FIRST)

**Step 1: Check What's Available**

```bash
# Make sure you're on main and up-to-date
git fetch origin
git pull origin main
bd sync --from-main

# See what's ready to work on
bd ready
```

**Step 2: CLAIM the Task (This updates `.beads/`)**

```bash
# Mark the issue as in_progress in your local beads tracker
bd update scribble-spl --status=in_progress
```

**Step 3: LOCK the Claim Immediately (Push `.beads/` to main BEFORE creating your branch)**

This is the critical step! You must push the `.beads/` changes to main immediately:

```bash
# Stage only the .beads/ directory changes
git add .beads/

# Commit the claim lock
git commit -m "chore: claim scribble-spl"

# Push immediately to lock it globally
git push origin main
```

**Why this order?** When you push this change to main, the `.beads/` tracker on the server is updated. Now when the OTHER developer runs `bd ready`, they will see `scribble-spl` is already `in_progress`, preventing them from claiming it.

**Step 4: Verify the Lock (Optional but Recommended)**

On the other developer's machine:
```bash
git pull origin main    # Gets your claim
bd sync --from-main     # Updates beads from main
bd ready                # Shows scribble-spl is no longer available
```

### Example: Avoiding a Race Condition

**The Problem (without claim lock):**
```
9:00 AM Dev 1: bd update scribble-spl --status=in_progress  (local only)
9:00 AM Dev 2: bd update scribble-spl --status=in_progress  (local only)

Both claimed the same task!

9:05 AM Dev 1: git push → PR → Gets merged
9:05 AM Dev 2: git push → PR → Wait... Dev 1 already finished this!
```

**The Solution (with claim lock):**
```
9:00 AM Dev 1: bd update scribble-spl --status=in_progress
9:00 AM Dev 1: git add .beads/ && git commit && git push origin main
              ↓ Claim is locked globally

9:00 AM Dev 2: bd ready
               Shows: scribble-spl is ALREADY in_progress (dev 1 claimed it)
9:00 AM Dev 2: Picks different task: scribble-eg8
```

---

## Creating a Branch for Your Feature

**IMPORTANT:** Only create a branch AFTER you've claimed and locked the task. This keeps `main` clean and allows the other developer to work simultaneously.

### Step 1: Create a Branch with Correct Naming (AFTER claiming)

**Branch name format:** `dev{1|2}/scribble-{ISSUE-ID}-short-description`

**For Developer 1 (Frontend):**
```bash
git checkout -b dev1/scribble-spl-init-react-vite
```

**For Developer 2 (Backend):**
```bash
git checkout -b dev2/scribble-5n7-init-go-backend
```

**Naming rules:**
- Use your developer number: `dev1` or `dev2`
- Include the issue ID: `scribble-spl`
- Use lowercase and hyphens (no spaces)
- Keep it short and descriptive

**Verify you're on the right branch:**

```bash
git branch
```

You should see your branch with a `*` next to it:
```
* dev1/scribble-spl-init-react-vite
  main
```

### Step 2: Make Your Changes and Commit

Do your work, then commit regularly:

```bash
# Make changes to files

# Commit frequently (not just at the end!)
git add .
git commit -m "feat(react): setup vite build configuration"

# Make more changes
git add .
git commit -m "feat(react): add tailwind css configuration"

# More changes
git add .
git commit -m "feat(react): create main app component"
```

**Why commit frequently?**
- Each commit is a checkpoint you can go back to
- Easier to review changes
- If something breaks, easier to find what caused it

---

## Pushing and Creating a Merge Request

### Step 1: Commit All Your Work

Make sure everything is committed:

```bash
git status
```

Should show: `nothing to commit, working tree clean`

If not, commit remaining changes:

```bash
git add .
git commit -m "your message"
```

### Step 2: Test Your Code Locally

**For Frontend:**
```bash
cd frontend
npm run lint        # Check for errors
npm run type-check  # Check types
npm run dev         # Test in browser
```

**For Backend:**
```bash
cd services/go-backend
go test ./...       # Run tests
go run cmd/server/main.go  # Test locally
```

**Make sure:**
- ✅ No console errors
- ✅ No warnings
- ✅ Feature works as expected
- ✅ No breaking changes to other code

### Step 3: Push Your Branch to Server

```bash
git push origin dev1/scribble-spl-init-react-vite
```

Replace `dev1/scribble-spl-init-react-vite` with your branch name.

**Output will show:**
```
Enumerating objects: 15, done.
Counting objects: 100% (15/15), done.
Writing objects: 100% (15/15), 340 bytes, done.
...
To github.com:yourusername/scribble.git
 * [new branch]      dev1/scribble-spl-init-react-vite -> dev1/scribble-spl-init-react-vite
```

### Step 4: Create a Merge Request

Use GitHub CLI (gh) to create a pull request:

```bash
gh pr create \
  --title "Initialize React + Vite frontend" \
  --body "Closes scribble-spl

## What changed
- Created Vite project with React
- Added Tailwind CSS configuration
- Setup TypeScript

## How to test
\`\`\`bash
cd frontend
npm install
npm run dev
# Visit http://localhost:5173
\`\`\`

## Checklist
- [x] Code follows project style
- [x] No console errors
- [x] Tested locally
- [x] Ready for review
"
```

**Key parts of the merge request:**
- **Title:** What you did (short)
- **Body:** Detailed explanation
- **"Closes scribble-spl":** Links to the issue
- **Test instructions:** How to verify it works
- **Checklist:** Confirm you tested

### Step 5: Wait for Review

The senior developer will review your code within 30 minutes. They might:
- ✅ Approve it
- 💬 Ask for changes
- ❓ Ask questions

Check GitHub or get a notification.

### Step 6: If Changes Requested

**Do not delete the branch!** Make changes in the same branch:

```bash
# Make the requested changes
git add .
git commit -m "fix: address code review feedback"

# Push the new commit
git push origin dev1/scribble-spl-init-react-vite
```

The merge request automatically updates with your new commit.

### Step 7: Once Approved - Merge to Main

The senior dev will merge your code to main. After they do:

```bash
# Go back to main
git checkout main

# Update your local main with the merged code
git pull origin main

# Delete your feature branch (it's done!)
git branch -d dev1/scribble-spl-init-react-vite
```

### Step 8: Mark Issue Complete

Now that your code is in main:

```bash
# Verify your code is in main
git log --oneline -5

# Mark the issue complete
bd close scribble-spl --reason="Code tested and merged to main"
```

**This unlocks dependent issues for the other developer!**

---

## Handling Merge Conflicts

**Merge conflicts happen when you and the other developer edit the same file.**

### Scenario: You Pull and Get a Conflict

```bash
git pull origin main
# Output shows:
# CONFLICT (content): Merge conflict in frontend/src/App.tsx
```

### Step 1: See What Conflicts

```bash
git status
```

Shows files with conflicts in red.

### Step 2: Open the Conflicted File

Open `frontend/src/App.tsx` in your editor. You'll see:

```typescript
<<<<<<< HEAD
  // Your code here
  function MyComponent() {
    return <div>My version</div>
  }
=======
  // Other developer's code
  function MyComponent() {
    return <div>Their version</div>
  }
>>>>>>> origin/main
```

**Markers:**
- `<<<<<<< HEAD` - Your code
- `=======` - Divider
- `>>>>>>> origin/main` - Their code

### Step 3: Fix the Conflict

Decide which code to keep:

**Option A: Keep your code**
```typescript
function MyComponent() {
  return <div>My version</div>
}
```

**Option B: Keep their code**
```typescript
function MyComponent() {
  return <div>Their version</div>
}
```

**Option C: Keep both (merge them)**
```typescript
function MyComponent() {
  return (
    <div>
      <div>My version</div>
      <div>Their version</div>
    </div>
  )
}
```

Remove the `<<<<<<`, `=======`, `>>>>>>>` markers entirely.

### Step 4: Complete the Merge

```bash
git add frontend/src/App.tsx
git commit -m "fix: resolve merge conflict in app component"
git push origin main
```

### Step 5: Tell the Team

Post in Slack/chat: "Resolved merge conflict in App.tsx - please pull latest"

---

## Common Commands Reference

### Check Status
```bash
git status              # See what changed
git log --oneline -10   # See last 10 commits
git branch -a           # See all branches
```

### Create/Switch Branches
```bash
git checkout -b dev1/feature-name    # Create new branch
git checkout main                     # Switch to main
git branch -d dev1/old-feature        # Delete branch
```

### Stage and Commit
```bash
git add .                             # Stage all changes
git add specific-file.ts              # Stage one file
git commit -m "message"               # Commit staged changes
git diff                              # See unstaged changes
```

### Push and Pull
```bash
git push origin dev1/feature-name     # Push your branch
git pull origin main                  # Pull from main
git fetch origin                      # Download updates (no merge)
```

### Undo Changes
```bash
git restore file.ts                   # Undo changes to a file
git reset HEAD file.ts                # Unstage a file
git revert HEAD                       # Undo last commit (creates new commit)
```

### See History
```bash
git log --oneline                     # See all commits
git log -5                            # See last 5 commits
git show commit-hash                  # See what changed in a commit
```

---

## Git Rules for Scribble

### Rule 1: Use Your Developer Branch

**Developer 1:** Always work in `dev1/*` branches
**Developer 2:** Always work in `dev2/*` branches

```bash
# ✅ CORRECT
git checkout -b dev1/scribble-spl-init-react-vite
git checkout -b dev2/scribble-5n7-init-go-backend

# ❌ WRONG
git checkout -b my-cool-feature
git checkout -b scribble-spl-init-react  # Missing dev1/dev2
```

### Rule 2: Always Pull Before Starting

Never start work without pulling latest:

```bash
# ALWAYS do this first
git fetch origin
git pull origin main
bd sync --from-main
```

### Rule 3: Commit Frequently

Don't wait until you're done to commit everything at once:

```bash
# ❌ BAD - One huge commit
git add .
git commit -m "big changes"

# ✅ GOOD - Multiple logical commits
git add frontend/src/App.tsx
git commit -m "feat(react): initialize app component"

git add frontend/src/components/Header.tsx
git commit -m "feat(react): add header component"

git add frontend/tailwind.config.js
git commit -m "feat(react): configure tailwind css"
```

### Rule 4: Commit Messages Follow Format

```
type(scope): description

- type: feat, fix, refactor, docs, test, chore
- scope: what part (frontend, backend, database, etc.)
- description: lowercase, no period, short
```

**ALWAYS:**
```bash
git commit -m "feat(frontend): initialize react app"
```

**NEVER:**
```bash
git commit -m "updates"
git commit -m "Fixed code"
git commit -m "WIP"
git commit -m "IMPORTANT CHANGES!!!"
```

### Rule 5: Test Before Pushing

**Frontend:**
```bash
cd frontend
npm run lint          # No errors
npm run type-check    # TypeScript OK
npm run dev           # Works in browser
```

**Backend:**
```bash
cd services/go-backend
go test ./...         # Tests pass
go run cmd/server/main.go  # Starts without errors
```

### Rule 6: One Issue Per Branch

Don't mix multiple issues in one branch:

```bash
# ✅ GOOD - One issue, one branch
git checkout -b dev1/scribble-spl-init-react-vite
# Work on scribble-spl only

# ❌ BAD - Multiple issues in one branch
git checkout -b dev1/multiple-stuff
# Work on scribble-spl AND scribble-rel AND scribble-eg8
```

### Rule 7: Wait for Code Review

Never merge your own code. Always wait for the senior dev:

```bash
# ✅ Create PR and wait
gh pr create --title "..." --body "..."
# Wait for review

# ❌ DON'T do this
git push origin dev1/feature
git checkout main
git merge dev1/feature  # WRONG!
```

### Rule 8: Delete Branch After Merge

After merging, clean up:

```bash
# After merge is complete
git checkout main
git pull origin main
git branch -d dev1/scribble-spl-init-react-vite  # Delete local
git push origin -d dev1/scribble-spl-init-react-vite  # Delete remote
```

---

## Typical Day Workflow

### Morning
```bash
# 1. Start work
cd /home/nahtao97/projects/scribble
git fetch origin && git pull origin main
bd sync --from-main

# 2. Check what's ready
bd ready

# 3. CLAIM an issue
bd update scribble-XXX --status=in_progress

# 4. LOCK the claim immediately (push .beads/ to main)
git add .beads/
git commit -m "chore: claim scribble-XXX"
git push origin main

# 5. NOW create your branch
git checkout -b dev1/scribble-XXX-short-description

# 6. Start development
npm run dev  # or go run ...
```

### During Day (Every 30-60 min)
```bash
# Make changes and commit
git add .
git commit -m "feat(scope): what you did"

# Test your code
npm run lint  # frontend
go test ./...  # backend
```

### End of Day
```bash
# Final commit
git add .
git commit -m "feat(scope): final changes"

# Push to server
git push origin dev1/scribble-XXX-short-description

# Create merge request (if ready for review)
gh pr create --title "..." --body "..."

# Or if still working, just leave it pushed
```

### When Ready for Review
```bash
# 1. Test everything works
npm run dev
# Visit http://localhost:5173 and test thoroughly

# 2. Push if not already pushed
git push origin dev1/scribble-XXX-short-description

# 3. Create merge request
gh pr create --title "Feature name" --body "Closes scribble-XXX"

# 4. Wait for review (30 min typically)

# 5. After approval and merge
git checkout main
git pull origin main
bd close scribble-XXX --reason="Merged and tested"

# 6. Next day, pick next issue
bd ready
```

---

## Troubleshooting

### "I committed to the wrong branch"

```bash
# Undo the last commit (but keep changes)
git reset --soft HEAD~1

# Or just start over:
git reset --hard origin/main
```

### "I want to see what's different from main"

```bash
git diff main...HEAD
```

Shows what you added compared to main.

### "How do I undo my last commit?"

```bash
# Undo last commit, keep changes
git reset --soft HEAD~1

# Or completely undo it
git reset --hard HEAD~1
```

### "I'm on the wrong branch"

```bash
# Create a new branch with your commits
git checkout -b correct-branch-name

# Go back to wrong branch and delete it
git checkout wrong-branch
git reset --hard origin/main
git checkout -
git branch -d wrong-branch
```

### "I have uncommitted changes and want to switch branches"

```bash
# Save your changes temporarily
git stash

# Switch branches
git checkout other-branch

# Come back and restore
git checkout your-branch
git stash pop
```

### "I want to see my commit history"

```bash
# Last 10 commits
git log --oneline -10

# Pretty graph view
git log --graph --oneline --decorate --all

# See who changed a file
git log --oneline frontend/src/App.tsx
```

---

## Questions?

If you're stuck:

1. **Check this guide** - Most questions answered here
2. **Ask in standup** - Daily at 9 AM
3. **Slack the team** - Quick question
4. **Senior dev** - Complex git issues

---

## Key Takeaways

✅ **DO THIS:**
- Pull before starting: `git pull origin main`
- Use dev1/dev2 branches: `git checkout -b dev1/feature`
- Commit frequently: `git commit -m "feat(scope): msg"`
- Test before pushing: `npm run dev` or `go test`
- Wait for review: Create MR and be patient
- Communicate: Post in Slack when pushing

❌ **DON'T DO THIS:**
- Work on main branch
- Commit without testing
- One huge commit with everything
- Commit messages like "changes" or "WIP"
- Merge your own code
- Delete .git folder
- Force push (`git push -f`)

---

**Remember:** Git is your safety net. Commit frequently, push regularly, and ask for help when confused. That's what the team is here for!

---

**Last Updated:** December 31, 2025
**For:** Scribble Team (Developers 1 & 2)
**Questions?** Ask Senior Developer
