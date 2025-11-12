# Release Guide

This document explains how to create a new release for garmin-to-ido.

## Creating a Release

The project uses GitHub Actions to automatically build and publish releases. When you push a version tag, the workflow will:

1. Build binaries for multiple platforms (Linux, macOS, Windows)
2. Generate checksums for verification
3. Create a GitHub Release with all binaries attached
4. Auto-generate release notes from commits

## Steps to Create a Release

### 1. Make sure your code is ready
Ensure all changes are committed and pushed to the `main` branch:

```bash
git add .
git commit -m "Your commit message"
git push origin main
```

### 2. Create and push a version tag

```bash
# Create a new tag (use semantic versioning: v1.0.0, v1.2.3, etc.)
git tag -a v1.0.0 -m "Release version 1.0.0"

# Push the tag to GitHub
git push origin v1.0.0
```

### 3. Wait for the workflow to complete

- Go to your repository on GitHub
- Click on the "Actions" tab
- You should see the "Release" workflow running
- Wait for it to complete (usually takes 2-3 minutes)

### 4. Check your release

- Go to the "Releases" section of your repository
- You should see your new release with:
  - Release notes auto-generated from commits
  - Binary files for different platforms
  - A `checksums.txt` file for verification

## Available Binaries

Each release includes binaries for:

- **Linux AMD64** (`garmin-to-ido-linux-amd64`)
- **Linux ARM64** (`garmin-to-ido-linux-arm64`)
- **macOS Intel** (`garmin-to-ido-darwin-amd64`)
- **macOS Apple Silicon** (`garmin-to-ido-darwin-arm64`)
- **Windows AMD64** (`garmin-to-ido-windows-amd64.exe`)

## Semantic Versioning

Use [Semantic Versioning](https://semver.org/) for version numbers:

- **MAJOR** version (v1.0.0 → v2.0.0): Breaking changes
- **MINOR** version (v1.0.0 → v1.1.0): New features, backward compatible
- **PATCH** version (v1.0.0 → v1.0.1): Bug fixes, backward compatible

## Example: Creating v1.0.0

```bash
# Ensure everything is committed
git status

# Create and push the tag
git tag -a v1.0.0 -m "Initial release"
git push origin v1.0.0

# Wait for the GitHub Action to complete
# Visit: https://github.com/YOUR_USERNAME/garmin-to-ido/releases
```

## Deleting a Release (if needed)

If you need to delete a release:

```bash
# Delete the tag locally
git tag -d v1.0.0

# Delete the tag remotely
git push origin :refs/tags/v1.0.0

# Manually delete the release from GitHub UI (Releases → Delete)
```

## Troubleshooting

**Q: The workflow didn't trigger**
- Make sure your tag follows the pattern `v*.*.*` (e.g., v1.0.0, not 1.0.0)
- Check that you pushed the tag: `git push origin v1.0.0`

**Q: The build failed**
- Check the Actions tab for error messages
- Ensure all dependencies are properly listed in `go.mod`

**Q: Binary doesn't work on user's machine**
- Users may need to install Python and `garminconnect` package separately
- Consider documenting this in the release notes
