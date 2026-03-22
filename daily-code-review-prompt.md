# Daily Code Review Prompt

Conduct a daily code review of pull requests on the `kubernetes/kubernetes` repository where `aaron-prindle` is either an assignee or a requested reviewer.

## Environment Notes

- The MCP GitHub tools are scoped to `aaprindle/kubernetes` (NOT `aaron-prindle/kubernetes`).
- The git proxy only allows access to `aaprindle/kubernetes`. To fetch from `kubernetes/kubernetes` or PR author forks, use WebFetch on the `.diff` URL (following redirects to `patch-diff.githubusercontent.com`).
- Use `mcp__github__search_pull_requests` (which works cross-repo) to find PRs on `kubernetes/kubernetes`.
- Use `mcp__github__pull_request_read`, `mcp__github__create_pull_request`, `mcp__github__push_files`, `mcp__github__create_or_update_file`, etc. only on `aaprindle/kubernetes`.

## PR Selection

1. **Fetch open PRs** targeting `aaron-prindle` using the GitHub search API (only PRs updated in the last 14 days):
   - `search_pull_requests` with query `is:open repo:kubernetes/kubernetes assignee:aaron-prindle updated:>={14_DAYS_AGO_YYYY-MM-DD}`
   - `search_pull_requests` with query `is:open repo:kubernetes/kubernetes reviewer:aaron-prindle updated:>={14_DAYS_AGO_YYYY-MM-DD}`

   Replace `{14_DAYS_AGO_YYYY-MM-DD}` with the actual date 14 days before today (e.g., if today is 2026-03-22, use `2026-03-08`).

2. **Deduplicate and prioritize:**
   - First priority: PRs authored by `lalitc375` or `yongruilin` (sorted most recent first)
   - Second priority: all other PRs (sorted most recent first)

3. **Check for already-reviewed PRs** by searching for PRs on `aaprindle/kubernetes` with `[Review]` in the title. Skip any upstream PR that already has a corresponding review PR.

4. **Select the top 3 unreviewed PRs** from the queue.

## Review Process

For each selected PR:

### Step 1: Get the upstream PR diff and context

Since MCP read tools are restricted to `aaprindle/kubernetes`, fetch diffs via WebFetch:

```
WebFetch: https://github.com/kubernetes/kubernetes/pull/<number>/files
  prompt: "Extract all file diffs with full paths, added/removed lines, and complete changes"
```

Also use the search results from Step 1 for PR metadata (title, body, author, labels).

### Step 2: Create a mirror PR on the `aaprindle/kubernetes` fork

**Goal:** Push the actual changed files (not just a marker) so the fork PR shows real code diffs.

1. **Read the base files** from the local repo at `/home/user/kubernetes/` (this is a clone of `aaprindle/kubernetes` which is a fork of `kubernetes/kubernetes`, so it has the base versions of files).

2. **Apply the diffs** from the WebFetch results to produce the modified file content.

3. **Create a branch** on the fork:
   ```
   mcp__github__create_branch(owner: "aaprindle", repo: "kubernetes", branch: "review/<upstream-pr-number>")
   ```

4. **Push the modified files** to the branch:
   - For small files (< ~15KB): use `mcp__github__push_files` with all files in one call
   - For large files: use `mcp__github__create_or_update_file` individually (requires the SHA of the existing file, obtainable via `git rev-parse origin/<branch>:<path>` after `git fetch origin <branch>`)
   - Commit message: the upstream PR title

5. **If any files could not be pushed** (e.g., too large to reconstruct from diffs, or WebFetch truncated the content), add a comment to the PR listing the missing files and explaining why. Link to the upstream PR's files tab for the full diff.

6. **Create the PR:**
   ```
   mcp__github__create_pull_request(
     owner: "aaprindle", repo: "kubernetes",
     head: "review/<upstream-pr-number>",
     base: "master",
     title: "[Review] upstream#<number>: <original-title>",
     body: "Mirror of https://github.com/kubernetes/kubernetes/pull/<number> for code review purposes.\n\n**Author:** <author>\n**Type:** <labels>\n**Area:** <sig labels>"
   )
   ```

### Step 3: Submit a GitHub PR review with inline comments

Use the pending review workflow to create proper line-level inline comments:

1. **Create a pending review** (no `event` parameter):
   ```
   mcp__github__pull_request_review_write(method: "create", owner: "aaprindle", repo: "kubernetes", pullNumber: <fork-pr-number>)
   ```

2. **Add inline comments** for each finding using `mcp__github__add_comment_to_pending_review`:
   - `path`: relative file path in the repo
   - `line`: the line number in the diff (for the last line of the range)
   - `startLine`: (optional) for multi-line comments, the first line
   - `side`: `RIGHT` for new code, `LEFT` for removed code
   - `subjectType`: `LINE`
   - `body`: the review feedback

   Aim for 3-6 inline comments per PR covering:
   - Bugs or correctness issues
   - Style or naming concerns
   - Performance considerations
   - Missing edge cases in tests
   - Security implications
   - Suggestions for improvement

3. **Submit the review** with a summary body:
   ```
   mcp__github__pull_request_review_write(
     method: "submit_pending",
     owner: "aaprindle", repo: "kubernetes",
     pullNumber: <fork-pr-number>,
     event: "COMMENT",
     body: "<overall summary>"
   )
   ```

   The summary body should include:
   - Overall design and approach assessment
   - Higher-level architectural concerns or questions
   - Missing tests or documentation
   - Overall recommendation (approve / request changes / comment only)

## Output

After completing all reviews, print a summary table:

| # | Upstream PR | Fork Review PR | Author | Title | Inline Comments |
|---|------------|----------------|--------|-------|----------------|
| 1 | [link]     | [link]         | ...    | ...   | N comments      |

And a skipped table:

| Upstream PR | Author | Title | Reason |
|------------|--------|-------|--------|
| [link]     | ...    | ...   | Already reviewed / Outside top 3 |
