# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.2.x   | :white_check_mark: |
| < 1.2   | :x:                |

## Reporting a Vulnerability

Report security vulnerabilities by emailing shantoislamdev@gmail.com.

**Do not open a public GitHub issue for security vulnerabilities.**

Please include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact

You can expect a response within 72 hours.

## Security Measures

- `.secrets.yaml` is gitignored; API keys are never committed to the repository
- npm and pip download scripts verify SHA256 checksums against GoReleaser output
- Checkpoint files use restrictive permissions (0600)
- No telemetry, analytics, or phone-home behavior
- All dependencies are minimal and well-vetted (5 direct Go dependencies)
