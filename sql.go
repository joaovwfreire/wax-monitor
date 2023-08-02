package main

// commands to create sql tables
const createStakesRemoved = "CREATE TABLE stakesRemoved (id INT AUTO_INCREMENT PRIMARY KEY, transaction_id VARCHAR(64) NOT NULL UNIQUE, processed TINYINT(1) DEFAULT 0, stake_removal_tx VARCHAR(64), pool_id BIGINT UNSIGNED, asset_ids JSON, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP);"
