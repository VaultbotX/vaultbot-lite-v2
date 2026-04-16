# Learnings & Operational Notes

## Claude Web — Workflow File Limitation

**Problem:** When Claude Code is running on the web (claude.ai/code), the GitHub OAuth token it uses does not have the `workflow` scope. GitHub rejects any `git push` that includes changes to `.github/workflows/` files without this scope, with the error:

```
refusing to allow an OAuth App to create or update workflow `...` without `workflow` scope
```

**Workaround:** When Claude needs to add or modify a workflow file:

1. Claude makes all other code changes and pushes them normally.
2. Claude pastes the full intended workflow file content directly in the chat.
3. The human manually creates or updates the workflow file on the branch with that content and pushes it.
4. Claude then verifies the pushed file looks correct once the human confirms it is up.
