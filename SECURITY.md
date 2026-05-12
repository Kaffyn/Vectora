# Security Policy

## Supported Versions

Security fixes are applied to the actively maintained branches and releases of Vectora.
When a vulnerability is confirmed, the fix should be backported to all supported branches if applicable.

## Reporting a Vulnerability

If you discover a security issue, do not open a public issue.

Please report it through the repository's private security channel or the maintainer contact path used by the project.
If no private channel is available, create a minimal public report that avoids technical exploit details and request a private follow-up.

Include as much of the following as possible:

- Affected component or path
- Severity estimate
- Reproduction steps
- Expected versus actual behavior
- Any relevant logs, screenshots, or proof of concept

## What to Avoid

- Public disclosure before a fix is ready
- Sharing secrets, tokens, or live credentials
- Running destructive tests against production data

## Disclosure Process

Once a report is received, the maintainers should:

1. Acknowledge receipt.
2. Validate the issue.
3. Prepare a fix.
4. Coordinate disclosure timing with the reporter.
5. Publish the fix and advisory after remediation.

## Security Expectations

All contributions should follow secure coding practices:

- Validate and sanitize input.
- Keep secrets out of logs and source control.
- Prefer least-privilege access.
- Use reviewed dependencies.
- Review changes that affect authentication, authorization, persistence, or external integrations with extra care.
