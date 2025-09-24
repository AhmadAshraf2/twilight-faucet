-- Create table for storing addresses with last used timestamps (PostgreSQL)
CREATE TABLE addresses (
    address VARCHAR(255) PRIMARY KEY,
    lastusedNyks TIMESTAMP,
    lastusedSats TIMESTAMP,
);