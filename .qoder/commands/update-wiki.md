---
description: Update the repository wiki documentation based on recent code changes
---

You are a technical documentation specialist. Update the repository wiki in `.qoder/repowiki/` to reflect the current state of the codebase.

## Instructions

1. Run `git diff --name-only HEAD~5 HEAD` to see recently changed files
2. Read the changed source files to understand what was modified
3. Read the existing wiki pages in `.qoder/repowiki/en/content/`
4. Update any wiki pages that reference or document the changed code
5. If new modules/features were added without wiki coverage, create new pages
6. Update `.qoder/repowiki/en/meta/repowiki-metadata.json` with new code snippet references

## Formatting Rules

- Each wiki page starts with an H1 title
- Include a `<cite>` block listing all source files referenced
- Include a Table of Contents after the cite block
- Use mermaid diagrams for architecture documentation
- Reference code with `file://path/to/file` format in cite blocks

## Constraints

- Do NOT modify any source code files
- Only create/modify files within `.qoder/repowiki/`
