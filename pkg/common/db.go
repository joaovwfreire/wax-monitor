package common

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// Check if the transaction has already been processed
func CheckTransactionProcessed(db *sql.DB, transactionId string) (bool, error) {
	query := "SELECT processed FROM stakesRemoved WHERE transaction_id = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		fmt.Println("Error: ", err)
		return false, err
	}
	defer stmt.Close()

	var processed bool
	err = stmt.QueryRow(transactionId).Scan(&processed)
	if err != nil {
		if err == sql.ErrNoRows {
			// Transaction ID not found, so insert a new row with processed = false
			insertQuery := "INSERT INTO stakesRemoved (transaction_id, processed) VALUES (?, 0)"
			_, insertErr := db.Exec(insertQuery, transactionId)
			if insertErr != nil {
				fmt.Println("Insert Error: ", insertErr)
				return false, insertErr
			}
			return false, nil
		}
		fmt.Println("Error: ", err)
		return false, err
	}

	return processed, nil
}

// Remove the stake from the database
func ProcessTransactionId(db *sql.DB, transactionId, stakeRemovalTx string, poolId uint64, assetIds []uint64) (bool, error) {
	query := "UPDATE stakesRemoved SET processed = 1, stake_removal_tx = ? WHERE transaction_id = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		fmt.Println("Error: ", err)
		return false, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(stakeRemovalTx, transactionId)
	if err != nil {
		fmt.Println("Error: ", err)
		return false, err
	}

	return true, nil

}
