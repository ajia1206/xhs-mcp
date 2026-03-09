# MCP And Skill Restructure Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Restructure the repository so the Xiaohongshu MCP source lives under `mcp/` and the external skill project is copied into `skill/`.

**Architecture:** Keep the existing Go module intact by moving the whole current MCP source tree as-is into `mcp/`. Copy the standalone skill repository into `skill/` without its `.git` metadata, then update in-repo docs and helper scripts to point at the new locations. Do not touch the separate data-analysis and visualization assets.

**Tech Stack:** Go, shell scripts, Markdown documentation, git working tree operations

---

### Task 1: Establish the target layout

**Files:**
- Create: `docs/plans/2026-03-09-mcp-skill-restructure.md`
- Modify: `README.md`

**Step 1: Define the target directories**

Target layout:

```text
.
├── mcp/
│   ├── go.mod
│   ├── main.go
│   └── ...
├── skill/
│   ├── README.md
│   ├── SKILL.md
│   └── LICENSE
```

**Step 2: Keep non-target assets in place**

Leave `scripts/`, `web/`, `data/`, and `artifacts/` where they are.

### Task 2: Move the MCP project

**Files:**
- Move: `build/xiaohongshu-mcp` -> `mcp`
- Modify: `scripts/xhs-ready.sh`
- Modify: `README.md`

**Step 1: Move the directory**

Run:

```bash
mv build/xiaohongshu-mcp mcp
```

**Step 2: Update helper paths**

Change build and runtime references from `build/xiaohongshu-mcp` to `mcp`.

### Task 3: Copy the external skill project

**Files:**
- Create: `skill/README.md`
- Create: `skill/SKILL.md`
- Create: `skill/LICENSE`
- Create: `skill/.gitignore`

**Step 1: Copy the repo contents without nested git metadata**

Run:

```bash
mkdir -p skill
rsync -a --exclude '.git' ../xiaohongshu-mcp-skill/ skill/
```

**Step 2: Normalize in-repo references if needed**

Update copied docs only if they still point to now-stale paths.

### Task 4: Adjust ignore rules and documentation

**Files:**
- Modify: `.gitignore`
- Modify: `README.md`
- Modify: `scripts/xhs-ready.sh`

**Step 1: Ignore moved runtime artifacts**

Add ignore entries for `mcp` local runtime artifacts where relevant.

**Step 2: Refresh the root README**

Describe the repository as containing the `mcp/` project and the `skill/` project.

### Task 5: Verify the restructure

**Files:**
- Verify: `mcp/go.mod`
- Verify: `skill/SKILL.md`

**Step 1: Run filesystem checks**

Run:

```bash
test -f mcp/go.mod
test -f skill/SKILL.md
```

**Step 2: Run a lightweight Go validation**

Run:

```bash
cd mcp && go test ./...
```

Expected:
- The Go module resolves from `mcp/`
- No import path regressions are introduced by the directory move

**Step 3: Inspect git status**

Run:

```bash
git status --short
```

Expected:
- The move is tracked as renames/additions
- No unrelated files were modified beyond the intended docs/script updates
