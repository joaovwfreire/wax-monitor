package rewards_service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	eos "github.com/eoscanada/eos-go"
	_ "github.com/go-sql-driver/mysql"

	common "github.com/joaovwfreire/wax-monitor/pkg/common"
)

const (
	MakeRewardActionName = "makereward"
	poolId               = 1
	contractName         = "breederstake"
	tableName            = "pools"
	QueryInterval        = 5 * time.Second
	maxRetries           = 25
	retryDelay           = 2 * time.Second
	longDelay            = 25 * time.Second
)

func makeRewards(poolId int64, api *eos.API) (eos.Checksum256, error) {
	// Do an on-chain makereward call

	// Return the transaction ID and/or error
	return eos.Checksum256{}, nil
}

type Pool struct {
	PoolId         int64  `json:"pool_id"`
	CollectionName string `json:"collection_name"`
	RewardSymbol   string `json:"reward_symbol"`
	Reward         int64  `json:"reward"`
	TotalNfts      int64  `json:"total_nfts"`
	TotalWeight    int64  `json:"total_weight"`
	// interface{} those fields as we don't really care about the data
	Rarities      interface{} `json:"rarities"`
	RarityTotals  interface{} `json:"rarities_total"`
	Created       int64       `json:"created"`
	Updated       int64       `json:"updated"`
	LastReward    int64       `json:"last_reward"`
	LastScheduler int64       `json:"last_scheduler"`
	NextReward    int64       `json:"next_reward"`
}

func queryLoop(failureCount int) {
	for failureCount < maxRetries {

		apiURL := common.GetAPIEndpoint(failureCount)
		api := eos.New(apiURL)

		request := eos.GetTableRowsRequest{Code: contractName, Scope: contractName, Table: tableName, Limit: 100, JSON: true}

		info, err := api.GetTableRows(context.Background(), request)
		if err != nil {
			fmt.Println("Error: ", err)
			failureCount++
			time.Sleep(retryDelay)
			continue
		}

		var pools []Pool
		err = json.Unmarshal(info.Rows, &pools)
		if err != nil {
			fmt.Println("Error: ", err)
			queryLoop(failureCount)
			return
		}

		for _, pool := range pools {

			if pool.NextReward < time.Now().Unix() {
				_, err := makeRewards(pool.PoolId, api)
				if err != nil {
					fmt.Println("Error making rewards:", err)
					// not to worry, we will try again in the next iteration
				}
				fmt.Println("🚢 Made rewards for pool id: ", pool.PoolId)
			}
		}
		return
	}
	fmt.Println("Max retries reached, awaiting restart.")
	time.Sleep(longDelay)
}

func pollTransactions() {
	for {
		queryLoop(0)
		time.Sleep(QueryInterval)
	}
}

func Start() {

	pollTransactions()
}
