package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// Alternative: Insert with explicit lastused value (if needed)
func insertAddressWithTime(address string) error {
	db, error := connectToDatabase()
	if error != nil {
		return fmt.Errorf("failed to connect to database: %w", error)
	}
	defer db.Close()
	query := `INSERT INTO addresses (address, lastusedNyks, lastusedSats) VALUES ($1, NOW() AT TIME ZONE 'UTC', NOW() AT TIME ZONE 'UTC')`

	_, err := db.Exec(query, address)
	if err != nil {
		return fmt.Errorf("failed to insert address: %w", err)
	}

	return nil
}

func insertAddressWithNyksTime(address string) error {
	db, error := connectToDatabase()
	if error != nil {
		return fmt.Errorf("failed to connect to database: %w", error)
	}
	defer db.Close()
	query := `INSERT INTO addresses (address, lastusedNyks) VALUES ($1, NOW() AT TIME ZONE 'UTC')`

	_, err := db.Exec(query, address)
	if err != nil {
		return fmt.Errorf("failed to insert address: %w", err)
	}

	return nil
}

func insertAddressWithBtcTime(address string) error {
	db, error := connectToDatabase()
	if error != nil {
		return fmt.Errorf("failed to connect to database: %w", error)
	}
	defer db.Close()
	query := `INSERT INTO addresses (address, lastusedSats) VALUES ($1, NOW() AT TIME ZONE 'UTC')`

	_, err := db.Exec(query, address)
	if err != nil {
		return fmt.Errorf("failed to insert address: %w", err)
	}

	return nil
}

func updateNyksTime(address string) error {
	db, error := connectToDatabase()
	if error != nil {
		return fmt.Errorf("failed to connect to database: %w", error)
	}
	defer db.Close()
	query := `UPDATE addresses SET lastusedNyks = NOW() AT TIME ZONE 'UTC' WHERE address = $1`

	_, err := db.Exec(query, address)
	if err != nil {
		return fmt.Errorf("failed to update address: %w", err)
	}

	return nil
}

func updateBtcTime(address string) error {
	db, error := connectToDatabase()
	if error != nil {
		return fmt.Errorf("failed to connect to database: %w", error)
	}
	defer db.Close()
	query := `UPDATE addresses SET lastusedSats = NOW() AT TIME ZONE 'UTC' WHERE address = $1`

	_, err := db.Exec(query, address)
	if err != nil {
		return fmt.Errorf("failed to update address: %w", err)
	}

	return nil
}

func addressExists(address string) (bool, error) {
	db, error := connectToDatabase()
	if error != nil {
		return false, fmt.Errorf("failed to connect to database: %w", error)
	}
	defer db.Close()

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM addresses WHERE address=$1)`

	err := db.QueryRow(query, address).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check address existence: %w", err)
	}

	return exists, nil
}

// Check if an address was used less than a day ago
func nyksRecentlyUsed(address string) (bool, error) {
	db, error := connectToDatabase()
	if error != nil {
		return false, fmt.Errorf("failed to connect to database: %w", error)
	}
	defer db.Close()

	query := `SELECT lastusedNyks FROM addresses WHERE address=$1`
	var lastused time.Time

	err := db.QueryRow(query, address).Scan(&lastused)
	if err != nil {
		if err == sql.ErrNoRows {
			// Address doesn't exist, so it wasn't used recently
			return false, fmt.Errorf("address not found")
		}
		return false, fmt.Errorf("failed to check address usage: %w", err)
	}

	// Check if the timestamp is less than 24 hours old
	oneDayAgo := time.Now().UTC().Add(-24 * time.Hour)
	return lastused.After(oneDayAgo), nil
}

func btcRecentlyUsed(address string) (bool, error) {
	db, error := connectToDatabase()
	if error != nil {
		return false, fmt.Errorf("failed to connect to database: %w", error)
	}
	defer db.Close()

	query := `SELECT lastusedSats FROM addresses WHERE address=$1`
	var lastused time.Time

	err := db.QueryRow(query, address).Scan(&lastused)
	if err != nil {
		if err == sql.ErrNoRows {
			// Address doesn't exist, so it wasn't used recently
			return false, fmt.Errorf("address not found")
		}
		return false, fmt.Errorf("failed to check address usage: %w", err)
	}

	// Check if the timestamp is less than 24 hours old
	oneDayAgo := time.Now().UTC().Add(-24 * time.Hour)
	return lastused.After(oneDayAgo), nil
}

func connectToDatabase() (*sql.DB, error) {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
