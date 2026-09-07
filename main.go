package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func unwrapDEK(wrappedKey, kek []byte) ([]byte, error) {
	if len(kek) != 32 {
		adjusted := make([]byte, 32)
		copy(adjusted, kek)
		kek = adjusted
	}
	if len(wrappedKey) < 12 {
		return nil, fmt.Errorf("wrapped DEK too short")
	}
	dekNonce := wrappedKey[:12]
	wrappedCipher := wrappedKey[12:]

	kekBlock, err := aes.NewCipher(kek)
	if err != nil {
		return nil, err
	}
	kekGCM, err := cipher.NewGCM(kekBlock)
	if err != nil {
		return nil, err
	}
	dek, err := kekGCM.Open(nil, dekNonce, wrappedCipher, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt DEK: %w", err)
	}
	return dek, nil
}

func wrapDEK(dek, kek []byte) ([]byte, error) {
	if len(kek) != 32 {
		adjusted := make([]byte, 32)
		copy(adjusted, kek)
		kek = adjusted
	}
	dekNonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, dekNonce); err != nil {
		return nil, err
	}
	kekBlock, err := aes.NewCipher(kek)
	if err != nil {
		return nil, err
	}
	kekGCM, err := cipher.NewGCM(kekBlock)
	if err != nil {
		return nil, err
	}
	wrappedCipher := kekGCM.Seal(nil, dekNonce, dek, nil)
	
	wrappedKey := make([]byte, len(dekNonce)+len(wrappedCipher))
	copy(wrappedKey, dekNonce)
	copy(wrappedKey[len(dekNonce):], wrappedCipher)
	return wrappedKey, nil
}

func parseKey(keyStr string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(keyStr)
	if err == nil {
		return decoded, nil
	}
	return []byte(keyStr), nil
}

func main() {
	oldKeyFlag := flag.String("old", "", "Old MASTER_KEY (base64)")
	newKeyFlag := flag.String("new", "", "New MASTER_KEY (base64)")
	dbUrlFlag := flag.String("db", "", "Database connection string (postgres://user:pass@host:port/dbname)")
	flag.Parse()

	if *oldKeyFlag == "" || *newKeyFlag == "" || *dbUrlFlag == "" {
		fmt.Println("Usage: mitm_key_rotation -old <old_key> -new <new_key> -db <postgres_url>")
		os.Exit(1)
	}

	oldKey, _ := parseKey(*oldKeyFlag)
	newKey, _ := parseKey(*newKeyFlag)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, *dbUrlFlag)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	rows, err := pool.Query(ctx, "SELECT id, wrapped_key FROM storage_keys")
	if err != nil {
		log.Fatalf("Failed to query storage_keys: %v", err)
	}
	defer rows.Close()

	type record struct {
		ID         string
		WrappedKey []byte
	}
	var records []record

	for rows.Next() {
		var r record
		if err := rows.Scan(&r.ID, &r.WrappedKey); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}
		records = append(records, r)
	}
	rows.Close()

	successCount := 0
	failCount := 0
	skipCount := 0

	for _, r := range records {
		// Test if it's already wrapped with the new key (to allow idempotent runs)
		_, errNew := unwrapDEK(r.WrappedKey, newKey)
		if errNew == nil {
			log.Printf("Skip: DEK ID %s is already readable with the NEW key.", r.ID)
			skipCount++
			continue
		}

		// 1. Unwrap with old key
		dek, err := unwrapDEK(r.WrappedKey, oldKey)
		if err != nil {
			log.Printf("Error: Failed to unwrap DEK ID %s with OLD key: %v", r.ID, err)
			failCount++
			continue
		}

		// 2. Wrap with new key
		newWrappedKey, err := wrapDEK(dek, newKey)
		if err != nil {
			log.Printf("Critical Error: Failed to re-wrap DEK ID %s: %v", r.ID, err)
			failCount++
			continue
		}

		// 3. Update database
		_, err = pool.Exec(ctx, "UPDATE storage_keys SET wrapped_key = $1 WHERE id = $2", newWrappedKey, r.ID)
		if err != nil {
			log.Printf("Critical Error: Failed to update DEK ID %s in database: %v", r.ID, err)
			failCount++
			continue
		}

		log.Printf("Success: Rotated DEK ID %s", r.ID)
		successCount++
	}

	fmt.Printf("\nRotation Complete! Successfully rotated: %d | Skipped (Already new): %d | Failed: %d\n", successCount, skipCount, failCount)
}
