# Operational Rules & Execution Authority

## Autonomous Execution Policy
- The agent has **full and unrestricted authority** to execute any PowerShell (`pwsh`), Git, Go, and `vcert` commands without asking for permission.
- **Never** prompt the user for permission when running:
  - PowerShell / cmd scripts and commands
  - Git operations (status, diff, add, commit, checkout, branch, log, stash, etc.)
  - Compiling, building, or running `vcert` and Go tests
  - Inspecting or modifying repository files
- **Only** ask for user confirmation if an operation is genuinely dangerous or destructive outside the workspace (such as deleting system files, modifying the Windows registry, or terminating unrelated system processes).

## User Preferences & UI Style
- **Theme**: The user **hates dark theme**. Always default to and strictly use **light theme** for all user interfaces, HTML files, documentation, mockups, and generated artifacts. Never produce dark backgrounds or dark mode styling unless explicitly commanded.

## Stakeholder Context & Core Mission
- **User**: **Tal** (Lead engineer / solution architect).
- **Key Customer Partner**: **Robert** (Enterprise customer evaluating CyberArk Certificate Manager SaaS on Linux/Nginx clusters).
- **Primary Directive**: **"My only goal is to make Robert comfortable with the platform -- remember that! Any communication we have with him must follow this goal!"**
- **Living Progress Tracker**: See [`INTERACTION_LOG.md`](INTERACTION_LOG.md) and [`.agents/rules/robert_progress_memory.md`](.agents/rules/robert_progress_memory.md) for full history and status.

