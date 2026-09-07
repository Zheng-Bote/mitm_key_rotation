# MASTER_KEY Worst-Case Scenario: Permanent Key Loss

## Overview
The MitM system relies on **Envelope Encryption** (AES-256-GCM) to secure all sensitive payloads and system credentials. The `MASTER_KEY` (Key Encryption Key / KEK) is the root of this security architecture. It is used to encrypt and decrypt the individual Data Encryption Keys (DEKs) stored in the `storage_keys` table.

## The Worst-Case Scenario: Losing the MASTER_KEY
If the `MASTER_KEY` is permanently lost and you did not perform a key rotation (re-wrapping the DEKs) beforehand, you are facing a scenario known as **Crypto-Shredding**.

### Is Recovery Possible?
**No.** AES-256-GCM is mathematically secure. Without the original `MASTER_KEY`, there is absolutely no technical method, backdoor, or trick to decrypt the existing DEKs. 
- All previously ingested raw data (`raw_ingestion`) is unreadable.
- All transformed payload fragments (`target_fragments`) are unreadable.
- All credentials for Quell/Source systems (`source_credentials`) and Ziel/Target systems (`delivery_targets`) are permanently locked.

### Pre-Flight Check: Can you find it?
Before proceeding with a destructive system reset, thoroughly check your environment:
1. **Terminal History:** If you recently exported the key, run `history | grep MASTER_KEY` or check your `~/.bash_history` file.
2. **Environment Backups:** Check old `.env` files, `config.json.enc` backups, or your password manager.

If you find the old key, you can simply use the `mitm_key_rotation` tool to securely migrate to a new key.

---

## System Reset Guide (When the Key is Truly Lost)

If the `MASTER_KEY` is irretrievably lost, the existing encrypted data is essentially corrupted garbage. You must perform a "hard reset" of the encryption layer.

Because `storage_keys` is protected by foreign key constraints from `source_credentials`, `delivery_targets`, and `raw_ingestion`, you cannot simply delete the old keys. You must clear the dependent data first.

Connect to your MitM PostgreSQL database and execute the following commands carefully:

```sql
-- 1. Wipe unreadable raw data and fragments
TRUNCATE TABLE raw_ingestion CASCADE;
TRUNCATE TABLE target_fragments CASCADE;
TRUNCATE TABLE delivery_packages CASCADE;

-- 2. Delete all target endpoints and source credentials
-- (This releases the foreign key locks on the DEKs)
DELETE FROM delivery_targets;
DELETE FROM source_credentials;

-- 3. Delete the permanently locked Data Encryption Keys
DELETE FROM storage_keys;
```

### Post-Reset Steps
After clearing the database:
1. Ensure your new `MASTER_KEY` is securely backed up and injected into the Scheduler's environment.
2. Use the MitM Admin API or GUI (`mitm_adm-data`) to recreate your **Source Credentials** and **Delivery Targets**. 
3. The system will automatically generate fresh DEKs for these new entities, wrapped securely with your new `MASTER_KEY`.
