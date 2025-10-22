-- Create table for storing addresses with last used timestamps (PostgreSQL)
CREATE TABLE addresses (
    address VARCHAR(255) PRIMARY KEY,
    lastusedNyks TIMESTAMP,
    lastusedSats TIMESTAMP,
);

Create Table AddressMappings (
    twilightAddress VARCHAR(255) PRIMARY KEY,
    ethAddress VARCHAR(255) UNIQUE NOT NULL,
    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);