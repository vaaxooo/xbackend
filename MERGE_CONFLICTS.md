# Why merge requests show conflicts

If the merge request says the branch has conflicts, it usually means the target branch changed after this work started. The two branches now touch the same lines in files like `internal/modules/users/app/*` or the domain layer and Git cannot auto-merge.

## How to resolve
1. Fetch the latest changes from the target branch (e.g., `main`).
2. Rebase or merge those changes into this branch and resolve the reported files.
3. Run `go test ./...` to ensure the rebased code still works.
4. Push the updated branch so the merge request becomes conflict-free.

> Tip: during rebase, prefer keeping the newer domain invariants (`Email`, `PasswordHash`, `UserID`) and application-layer calls that rely on them.
