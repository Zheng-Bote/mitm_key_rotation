# Changelog

All notable changes to the `mitm_key_rotation` tool will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.0] - 2026-09-07

### Added
- **Initial Release**: Created the `mitm_key_rotation` tool to facilitate secure Envelope Encryption key rotation.
- **Idempotent Execution**: Implemented logic to test if a DEK is already wrapped with the new `MASTER_KEY` to prevent errors during repeated runs.
- **Documentation**: Added English `README.md` and a comprehensive guide in `docs/MASTER_KEY_WORST_CASE.md` detailing recovery procedures for a permanently lost Master Key.
