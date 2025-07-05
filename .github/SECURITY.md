# Security Policy

## Supported Versions

Use this section to tell people about which versions of your project are currently being supported with security updates.

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0.0 | :x:                |

## Reporting a Vulnerability

We take the security of Ngebut seriously. If you believe you've found a security vulnerability in our project, please follow these steps to report it:

### Where to Report

Please **DO NOT** report security vulnerabilities through public GitHub issues.

Instead, please report them privately through GitHub's security advisory feature:
1. Go to the repository's "Security" tab
2. Select "Report a vulnerability"
3. Fill out the form with details about the vulnerability

Alternatively, you can reach out to the maintainer (Achmad Irianto Eka Putra / @ryanbekhen) directly through GitHub.

### What to Include

When reporting a vulnerability, please include as much information as possible:

- Type of issue (e.g., buffer overflow, SQL injection, cross-site scripting, etc.)
- Full paths of source file(s) related to the vulnerability
- The location of the affected source code (tag/branch/commit or direct URL)
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit the issue

### Response Process

After you have submitted a vulnerability report, you can expect:

1. Confirmation of receipt within 48 hours
2. An initial assessment of the report within 7 days
3. Updates on our progress as we work to address the vulnerability

### Disclosure Policy

- The vulnerability will be addressed as quickly as possible
- A CVE (Common Vulnerabilities and Exposures) identifier will be requested if appropriate
- Once the vulnerability is fixed, we will release a security advisory
- Credit will be given to the reporter (unless anonymity is requested)

## Security Best Practices for Users

- Always use the latest version of Ngebut
- Regularly check for updates and security advisories
- Follow secure coding practices when implementing applications using Ngebut
- Implement proper input validation and output encoding
- Use HTTPS for all production deployments
- Set appropriate security headers
- Implement proper authentication and authorization mechanisms

## Security-Related Configuration

Ngebut provides several security-related features and middleware:

- Basic Authentication middleware
- CORS middleware
- Session management with secure defaults

Please refer to the documentation for each middleware to ensure you're implementing them securely.

## Contact

If you have any questions regarding this security policy, please contact the maintainer through GitHub [@ryanbekhen](https://github.com/ryanbekhen) or use the GitHub security advisory feature as described above.
