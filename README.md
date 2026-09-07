# mitm_key_rotation

`mitm_key_rotation` is a critical maintenance utility for the MitM Data Aggregator. It performs secure Key Rotation for the Envelope Encryption system by re-wrapping all existing Data Encryption Keys (DEKs) in the database from an old `MASTER_KEY` to a new one.

## Features

- **Secure Re-Wrapping**: Safely decrypts DEKs using the old AES-256-GCM Master Key and immediately re-encrypts them with the new Master Key (generating new nonces).
- **Idempotency**: Skips DEKs that are already decryptable by the *new* Master Key, preventing double-wrapping if the script is interrupted or run twice.
- **Detailed Logging**: Reports success, skip, and failure counts directly to the console for auditing.

## Prerequisites

- Go 1.25+
- Access to the MitM PostgreSQL database.
- Both the **OLD** and **NEW** `MASTER_KEY`s (base64 encoded).

## Building

```bash
cd maintenance-layer/mitm_key_rotation
go build -o bin/mitm_key_rotation main.go
```

## Usage

```bash
./bin/mitm_key_rotation -old "<OLD_BASE64_MASTER_KEY>" -new "<NEW_BASE64_MASTER_KEY>" -db "postgres://user:password@localhost:5432/mitm_db"
```

### Command-Line Arguments

| Flag | Description | Required |
|------|-------------|----------|
| `-old` | The current (old) base64-encoded `MASTER_KEY` | Yes |
| `-new` | The new base64-encoded `MASTER_KEY` to migrate to | Yes |
| `-db` | PostgreSQL connection string URL | Yes |

## Documentation

For instructions on what to do if the old `MASTER_KEY` is completely lost, please refer to the [MASTER_KEY Worst-Case Guide](docs/MASTER_KEY_WORST_CASE.md).

## License

This project is licensed under the Apache 2.0 License.
