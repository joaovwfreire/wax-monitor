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

type AtomicAssetsActionData struct {
	From     eos.AccountName `json:"from"`
	To       eos.AccountName `json:"to"`
	AssetIds []uint64        `json:"asset_ids"`
	Memo     string          `json:"memo"`
}

type BreederStakeActionData struct {
	PoolId   uint64          `json:"pool_id"`
	Username eos.AccountName `json:"username"`
	AssetIds []uint64        `json:"asset_ids"`
}

const (
	contractName         = "breederstake"
	atomicAssetsContract = "atomicassets"
	QueryInterval        = 5 * time.Second
	maxRetries           = 25
	retryDelay           = 2 * time.Second
	initialTransaction   = 0
	poolId               = 4455366986717
)

func handleStakeCreate(db *sql.DB, transactionId eos.Checksum256, poolId uint64, assetIds []uint64, user eos.AccountName, timestamp string) {
	// Handle the stake creation logic, including adding asset IDs to a table
	// Update the processed flag to false initially and then set to true after processing
	// SQL logic should be implemented here
}

func handleStakeRemove(db *sql.DB, transactionId eos.Checksum256, poolId uint64, assetIds []uint64, user eos.AccountName) {
	transactionIdString := transactionId.String()

	alreadyProcessed, err := common.CheckTransactionProcessed(db, transactionIdString)
	if err != nil {
		fmt.Println("Error: ", err)
		time.Sleep(retryDelay)
		handleStakeRemove(db, transactionId, poolId, assetIds, user)
		return
	}

	if !alreadyProcessed {
		for retryCount := 0; retryCount < maxRetries; retryCount++ {
			stakeRemovalTx, err := common.RemoveStakeFromChain(poolId, assetIds, user)
			if err == nil {
				for {
					_, err = common.ProcessTransactionId(db, transactionIdString, stakeRemovalTx, poolId, assetIds)
					if err != nil {
						fmt.Println("Error: ", err)
						time.Sleep(retryDelay)
						// Continue to retry processing the transaction ID
						continue
					}
					break // Break inner loop if transaction ID processing is successful
				}
				break // Break outer loop if stake removal is successful
			}
			fmt.Println("Error: ", err)
			time.Sleep(retryDelay)
		}
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
			contractName := action.Trace.Action.Account
			var actionData interface{}
			if contractName == "atomicassets" {
				actionData = &AtomicAssetsActionData{}
			} else if contractName == "breederstake" {
				actionData = &BreederStakeActionData{}
			}

			bytesData, _ := json.Marshal(action.Trace.Action.Data)
			json.Unmarshal(bytesData, actionData)

			// Handle actions accordingly
			switch action.Trace.Action.Name {
			case "stakecreate":
				if contractName == "breederstake" {
					data := actionData.(*BreederStakeActionData)
					handleStakeCreate(db, action.Trace.TransactionID, data.PoolId, data.AssetIds, data.Username, action.BlockTime.String())
				}
			case "logtransfer":
				if contractName == "atomicassets" {
					data := actionData.(*AtomicAssetsActionData)
					fmt.Println("Atomic Assets Action Data: ", data)
					if data.From == "atomicmarket" {
						// Do nothing if 'from' is 'atomicmarket'
						fmt.Println("From field is 'atomicmarket', no action taken.")
					} else {
						fmt.Println("Handling logtransfer.")
						handleStakeRemove(db, action.Trace.TransactionID, poolId, data.AssetIds, data.From) // Example action
					}
				}
			default:
				fmt.Println("%s contract action acknowledged", action.Trace.Action.Name)
			}

			lastProcessedAction = startingPoint + int64(i)
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

	startingPoint := int64(initialTransaction)

	pollTransactions(db, startingPoint)
}
