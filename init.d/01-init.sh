#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- Create new user
    CREATE USER $DB_APP_USER WITH PASSWORD '$DB_APP_PASSWORD';

    -- Create database
    CREATE DATABASE $DB_APP_USER;
EOSQL

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$DB_APP_USER" <<-EOSQL
    -- Create tables
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        email VARCHAR(255) UNIQUE NOT NULL,
        password VARCHAR(255) NOT NULL,
        is_email_verified BOOLEAN DEFAULT FALSE,      
        created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
    );

    CREATE TABLE IF NOT EXISTS urls (
        id SERIAL PRIMARY KEY,
        short_url VARCHAR(255) UNIQUE NOT NULL,
        original_url TEXT NOT NULL,
        user_id INT REFERENCES users(id) ON DELETE CASCADE,
        click_count INT DEFAULT 0,
        expiry_date TIMESTAMP WITH TIME ZONE,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
    );

    -- Grant privileges
    GRANT ALL PRIVILEGES ON DATABASE $DB_APP_USER TO $DB_APP_USER;
    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $DB_APP_USER;
    GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO $DB_APP_USER;
EOSQL
