# Known Issues

## IME/focus flake
E2E scripts may fail when IME/focus interferes.

Mitigation:
- Use headless/API paths
- Add bounded retries
- Keep GUI scripts out of CI
