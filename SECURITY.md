# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |

## Reporting a Vulnerability

We take the security of junit-to-checklist seriously. If you discover a security vulnerability, please follow these steps:

### How to Report

1. **Do NOT** open a public GitHub issue for security vulnerabilities
2. Email security concerns to the project maintainers
3. Include the following information:
   - Description of the vulnerability
   - Steps to reproduce the issue
   - Potential impact
   - Suggested fix (if available)

### What to Expect

- **Acknowledgment**: We will acknowledge receipt of your vulnerability report within 48 hours
- **Updates**: We will provide updates on the status of your report within 7 days
- **Resolution**: We will work to patch confirmed vulnerabilities promptly
- **Credit**: If you wish, we will credit you in the CHANGELOG when the fix is released

### Security Considerations

This tool processes XML files and generates markdown output. Key security considerations:

- **Input Validation**: The tool parses XML from JUnit reports - ensure reports come from trusted sources
- **File System Access**: The tool reads from and writes to the local file system
- **Dependencies**: Regularly update dependencies to patch known vulnerabilities
- **No Network Access**: This tool does not make network requests or send data externally

### Best Practices for Users

1. Only process JUnit XML reports from trusted CI/CD pipelines
2. Keep dependencies up to date (`npm audit` and `npm update`)
3. Review generated output before sharing sensitive test failure information
4. Use the tool in secure environments where file system access is controlled

## Known Security Limitations

- This tool does not sanitize or validate XML input beyond basic parsing
- Generated markdown files may contain sensitive information from test failures
- The tool has file system write access in the directory where it runs

If you have questions about security practices for this project, please reach out to the maintainers.
