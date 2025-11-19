# Git Repository Cleanup

## Problem
Git push was transferring 14+ MB of data due to large binary files being tracked.

## Files Removed
- `backend/server` (51MB) - Go binary that was accidentally committed

## .gitignore Updates
Added to ignore Go binaries anywhere in the project:
```
# Go binaries (anywhere in the project)
server
migrate
```

## Commands Run
```bash
# Remove tracked binary
git rm --cached backend/server

# Update .gitignore
# (added server and migrate to ignore list)
```

## Verification
After cleanup, largest tracked files are:
- `frontend/package-lock.json` (248K) - normal
- Spec markdown files (20-36K) - normal
- `frontend/app/favicon.ico` (28K) - normal

## Next Steps
1. Commit these changes:
   ```bash
   git add .gitignore
   git commit -m "chore: remove binary and update .gitignore"
   ```

2. Push (should be much faster now):
   ```bash
   git push origin main
   ```

## Prevention
The updated .gitignore now catches:
- `backend/bin/` - build output directory
- `__debug_bin*` - debug binaries
- `server` - server binary (anywhere)
- `migrate` - migration binary (anywhere)
- All standard Go binary extensions (`.exe`, `.dll`, `.so`, `.dylib`)

## Note
If the binary was in previous commits, the repository history still contains it. To completely remove it from history (optional, only if needed):

```bash
# WARNING: This rewrites history - coordinate with team first!
git filter-branch --force --index-filter \
  'git rm --cached --ignore-unmatch backend/server' \
  --prune-empty --tag-name-filter cat -- --all

# Force push (dangerous!)
git push origin --force --all
```

However, this is usually not necessary unless the repository size is a problem. The file won't be in future commits, which is what matters most.
