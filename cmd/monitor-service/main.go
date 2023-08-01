package monitor_service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	eos "github.com/eoscanada/eos-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"

	common "github.com/joaovwfreire/wax-monitor/pkg/common"
)

const (
	StakeCreateActionName = "stakecreate"
	StakeRemoveActionName = "stakeremove"
	QueryInterval         = 5 * time.Second
	maxRetries            = 25
	retryDelay            = 2 * time.Second
)

func handleStakeCreate(db *sql.DB, transactionId eos.Checksum256, poolId uint64, assetIds []uint64, user eos.AccountName, timestamp string) {
	// Handle the stake creation logic, including adding asset IDs to a table
	// Update the processed flag to false initially and then set to true after processing
	// SQL logic should be implemented here
}

func handleStakeRemove(db *sql.DB, transactionId eos.Checksum256, poolId uint64, assetIds []uint64, user eos.AccountName) {
	transactionIdString := transactionId.String()

	alreadyProcessed := common.CheckTransactionProcessed(transactionIdString)

	if !alreadyProcessed {
		var stakeRemovalTx eos.Checksum256
		for retryCount := 0; retryCount < maxRetries; retryCount++ {
			stakeRemovalTx, err := common.RemoveStakeFromChain(poolId, assetIds)
			if err == nil {
				common.ProcessTransactionId(transactionIdString, stakeRemovalTx.String(), poolId, assetIds)
				break
			}
			fmt.Println("Error: ", err)
			time.Sleep(retryDelay)
		}
		common.ProcessTransactionId(transactionIdString, stakeRemovalTx.String(), poolId, assetIds)
	}

}

func queryLoop(startingPoint int64, db *sql.DB, failureCount int) {
	for {
		apiURL := common.GetAPIEndpoint(failureCount)
		api := eos.New(apiURL)

		name := eos.AccountName("breederstake")

		request := eos.GetActionsRequest{AccountName: name, Offset: eos.Int64(100), Pos: eos.Int64(startingPoint)}

		info, err := api.GetActions(context.Background(), request)

		if err != nil {
			fmt.Println("Error: ", err)
			failureCount++
			time.Sleep(400 * time.Millisecond)
			continue
		}

		lastProcessedAction := startingPoint
		for i, action := range info.Actions {

			var actionData struct {
				PoolId   uint64          `json:"pool_id"`
				Username eos.AccountName `json:"username"`
				AssetIds []uint64        `json:"asset_ids"`
			}

			bytesData, _ := json.Marshal(action.Trace.Action.Data)
			json.Unmarshal(bytesData, &actionData)

			// yul function selector pattern =,)
			switch action.Trace.Action.Name {
			case StakeCreateActionName:
				handleStakeCreate(db, action.Trace.TransactionID, actionData.PoolId, actionData.AssetIds, actionData.Username, action.BlockTime.String())
			case StakeRemoveActionName:
				handleStakeRemove(db, action.Trace.TransactionID, actionData.PoolId, actionData.AssetIds, actionData.Username)
			default:
				fmt.Println("%s contract action acknowledged", action.Trace.Action.Name)
			}

			lastProcessedAction = startingPoint + int64(i)
			// You can add code here to store the transaction in a general manner, regardless of the action type

		}

		startingPoint = lastProcessedAction + 1
		time.Sleep(400 * time.Millisecond)
	}
}

func pollTransactions(db *sql.DB, startingPoint int64) {
	for {
		queryLoop(startingPoint, db, 0)
		time.Sleep(QueryInterval)
	}
}

func Start() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	defer db.Close()

	startingPoint := int64(0)

	pollTransactions(db, startingPoint)
}
