# Contributing to Ruum

Thank you for your interest in contributing to Ruum! 🎉

## Commit Convention

We use [Conventional Commits](https://www.conventionalcommits.org/) for automatic versioning:

### Commit Types

- `feat:` - New feature (triggers **minor** version bump: 1.0.0 → 1.1.0)
- `fix:` - Bug fix (triggers **patch** version bump: 1.0.0 → 1.0.1)
- `perf:` - Performance improvement (triggers **patch** version bump)
- `refactor:` - Code refactoring (triggers **patch** version bump)
- `docs:` - Documentation changes (triggers **patch** version bump)
- `chore:` - Maintenance tasks (no version bump)
- `test:` - Test changes (no version bump)
- `ci:` - CI/CD changes (no version bump)

### Breaking Changes

Add `!` or `BREAKING CHANGE:` for major version bump (1.0.0 → 2.0.0):

```
feat!: change authentication API

BREAKING CHANGE: Authentication now requires explicit configuration
```

### Examples

```bash
# Feature (1.0.0 → 1.1.0)
git commit -m "feat: add WebSocket support"

# Bug fix (1.0.0 → 1.0.1)
git commit -m "fix: resolve memory leak in container"

# Performance (1.0.0 → 1.0.1)
git commit -m "perf: optimize routing performance"

# Breaking change (1.0.0 → 2.0.0)
git commit -m "feat!: redesign module system"

# Documentation (1.0.0 → 1.0.1)
git commit -m "docs: update installation guide"

# No release
git commit -m "chore: update dependencies"
```

## Development Workflow

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes following our coding standards
4. Write tests for new functionality
5. **Run pre-commit checks:** `make pre-commit` or `make ci`
6. Commit with conventional commit message
7. Push and create a Pull Request

### Quality Gates

Before your PR can be merged, it must pass:

- ✅ **All tests** must pass (with race detection)
- ✅ **Test coverage** must be ≥70%
- ✅ **Linter** must pass with no errors
- ✅ **Security checks** must pass (govulncheck + gosec)
- ✅ **Build** must succeed for all packages
- ✅ **Commit messages** must follow conventional commits
- ✅ **PR title** must follow conventional commits format

Use `make ci` to run all checks locally before pushing!

## Pull Request Process

1. Ensure all tests pass
2. Update documentation if needed
3. Follow the PR template
4. Wait for review from maintainers
5. Address review feedback
6. Once approved, maintainers will merge to `main`
7. **Automatic release will be triggered** 🚀

## Automatic Versioning

When your PR is merged to `main`:

1. ✅ **Quality Gates** run (tests, linting, security, coverage)
2. ⏸️  **Release blocked** if any check fails
3. 🏷️ **Version calculated** based on commit types (if checks pass)
4. 📝 **CHANGELOG.md** updated automatically
5. 🚀 **GitHub release** created
6. 📦 **Go modules** published to proxy

**Important:** Release only happens if ALL quality gates pass! This ensures no broken versions are published.

## Questions?

Feel free to:
- Open an issue
- Start a discussion
- Ask in pull requests

Thank you for contributing! 🙏

