package monitor_service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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

type LastProcessedBlock struct {
	BlockNum uint32 `json:"block_num"`
}

const (
	breedersContract     = "breederstake"
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
		fmt.Println("🚨 Error: ", err)
		time.Sleep(retryDelay)
		handleStakeRemove(db, transactionId, poolId, assetIds, user)
		return
	}

	if !alreadyProcessed {
		fmt.Println("🔍 Found a non-processed transaction with ID: ", transactionIdString)
		for retryCount := 0; retryCount < maxRetries; retryCount++ {
			stakeRemovalTx, err := common.RemoveStakeFromChain(poolId, assetIds, user)
			if err == nil {
				for {
					_, err = common.ProcessTransactionId(db, transactionIdString, stakeRemovalTx, poolId, assetIds)
					if err != nil {
						fmt.Println("🚨 Error: ", err)
						time.Sleep(retryDelay)
						// Continue to retry processing the transaction ID
						continue
					}
					break // Break inner loop if transaction ID processing is successful
				}
				break // Break outer loop if stake removal is successful
			}
			fmt.Println("🚨 Error: ", err)
			time.Sleep(retryDelay)
		}
	}
}

func readLastProcessedBlockNum() uint32 {
	file, err := os.Open("last_processed_block.json")
	if err != nil {
		fmt.Println("🚨 Error opening last_processed_block.json: ", err)
		return 0
	}
	defer file.Close()

	data := LastProcessedBlock{}
	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		fmt.Println("🚨 Error decoding JSON from last_processed_block.json: ", err)
		return 0
	}

	return data.BlockNum
}

func writeLastProcessedBlockNum(blockNum uint32) {
	data := LastProcessedBlock{BlockNum: blockNum}

	file, err := os.Create("last_processed_block.json")
	if err != nil {
		fmt.Println("🚨 Error: ", err)
		return
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(data)
	if err != nil {
		fmt.Println("🚨 Error: ", err)
	}
}

func defineDataType(contractName eos.AccountName) interface{} {
	if contractName == "atomicassets" {
		return &AtomicAssetsActionData{}
	} else if contractName == "breederstake" {
		return &BreederStakeActionData{}
	} else {
		fmt.Println("🚨 Non-monitored action from contract 📜", contractName)
		return nil
	}
}

func unmarshalActionDataWithAssetIds(action eos.ActionResp, actionData interface{}, contractName eos.AccountName) interface{} {
	dataMap, ok := action.Trace.Action.Data.(map[string]interface{})
	if ok {
		assetIdsInterface, ok := dataMap["asset_ids"]
		if ok {
			assetIdsRaw, ok := assetIdsInterface.([]interface{})
			if ok {
				assetIds := make([]uint64, len(assetIdsRaw))
				for i, v := range assetIdsRaw {
					// Convert each item in the slice to uint64
					switch id := v.(type) {
					case float64:
						assetIds[i] = uint64(id)
					case string:
						parsedId, err := strconv.ParseUint(id, 10, 64)
						if err != nil {
							fmt.Printf("🚨 Error parsing string asset ID to uint64: %v\n", err)
						} else {
							assetIds[i] = parsedId
						}
					default:
						fmt.Printf("🚨 Unexpected type for asset ID: %v, type: %T\n", v, v)
					}
				}

				// Set the AssetIds field in the actionData object
				if contractName == "atomicassets" {
					data := actionData.(*AtomicAssetsActionData)
					data.AssetIds = assetIds
				} else if contractName == "breederstake" {
					data := actionData.(*BreederStakeActionData)
					data.AssetIds = assetIds
				}
			}
		}

	}
	bytesData, _ := json.Marshal(action.Trace.Action.Data)
	json.Unmarshal(bytesData, actionData)

	return actionData
}

func handleAction(action eos.ActionResp, actionData interface{}, db *sql.DB, contractName eos.AccountName) {
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
			if data.From != "atomicmarket" {
				handleStakeRemove(db, action.Trace.TransactionID, poolId, data.AssetIds, data.From) // Example action
			}
		}
	default:
		// Do nothing.. Can add more cases if needed, but this should be enough for now
	}
}

func queryLoop(db *sql.DB, failureCount int) {
	fmt.Println("🔄 Starting a new query iteration\n...")
	apiURL := common.GetAPIEndpoint(failureCount)
	api := eos.New(apiURL)

	name := eos.AccountName("breederstake")

	request := eos.GetActionsRequest{AccountName: name, Offset: eos.Int64(-100), Pos: eos.Int64(-1)}

	info, err := api.GetActions(context.Background(), request)

	if err != nil {
		fmt.Println("🚨 Error: ", err)
		failureCount++
		time.Sleep(400 * time.Millisecond)
		return
	}

	lastProcessedBlockNum := readLastProcessedBlockNum()

	for i := len(info.Actions) - 1; i >= 0; i-- {

		action := info.Actions[i]
		if i == len(info.Actions)-1 {
			fmt.Println("\n🔍 Last action at block: %d 🔍", action.BlockNum)
			fmt.Println("🔍 Last irreversible block: %d 🔍", info.LastIrreversibleBlock)
			fmt.Println("🔍 Last processed block: %d 🔍", lastProcessedBlockNum)
			fmt.Println(time.Now(), "\n\n")

		}

		if action.BlockNum < lastProcessedBlockNum || action.BlockNum >= info.LastIrreversibleBlock-100 {
			continue
		}

		contractName := action.Trace.Action.Account
		actionData := defineDataType(contractName)
		unmarshalActionDataWithAssetIds(action, actionData, contractName)
		handleAction(action, actionData, db, contractName)

		if action.BlockNum > lastProcessedBlockNum {
			writeLastProcessedBlockNum(action.BlockNum)
		}
	}

	time.Sleep(400 * time.Millisecond)
}

func pollTransactions(db *sql.DB) {
	fmt.Println("🚀 Starting transaction polling\n...")
	for {
		queryLoop(db, 0)
		time.Sleep(QueryInterval)
	}
}

func Start() {
	fmt.Println("🚀 Starting the monitor service\n...")
	err := godotenv.Load()
	if err != nil {
		fmt.Println("🚨 Error loading .env file")
		return
	}

	fmt.Println("🌍 Environment variables loaded successfully.")

	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		fmt.Println("🚨 Error: ", err)
		return
	}
	defer db.Close()

	fmt.Println("📚 Database connection established successfully.")

	pollTransactions(db)
}
