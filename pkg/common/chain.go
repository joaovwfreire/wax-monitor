package common

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/eoscanada/eos-go"
	"github.com/joho/godotenv"
)

var endpoints = []string{
	"https://wax.greymass.com", // Priority endpoint
	"https://wax.eosdac.io",
	"https://wax.pink.gg",
	"https://wax.api.eosnation.io",
	"https://wax.eosrio.io",
	"https://wax.eosusa.io",
	"https://wax.eu.eosamsterdam.net",
	"https://api.wax.bountyblok.io",
	"https://api.waxsweden.org",
	"https://api.wax.alohaeos.com",
	"https://waxapi.ledgerwise.io",
	"https://api.wax.detroitledger.tech",
	"https://wax.eosphere.io",
	"https://api-wax.eosauthority.com",
	"https://wax-public1.neftyblocks.com",
	"https://wax-public2.neftyblocks.com",
	"https://apiwax.3dkrender.com",
	"https://aa.wax.blacklusion.io",
	"https://hyperion.wax.blacklusion.io",
	"https://wax.blacklusion.io",
	"https://wax.hivebp.io",
}

type StakeRemove struct {
	PoolId   uint64          `json:"pool_id"`
	AssetIds []uint64        `json:"asset_ids"`
	Owner    eos.AccountName `json:"owner"`
}

type MakeReward struct {
	PoolId uint64 `json:"pool_id"`
}

const adminAccount = "breedersmint"

func GetAPIEndpoint(failureCount int) string {
	if failureCount >= len(endpoints) {
		fmt.Println("🔁 Reached max endpoint switches. 🔁 \n Waiting for 30 seconds...")
		time.Sleep(30 * time.Second)
		return endpoints[0] // Reset to the first endpoint
	} else if failureCount > 0 {
		fmt.Println("🔀 Non-preferential API request detected. 🔀 \n Waiting for 2 seconds...")
		time.Sleep(2 * time.Second)
	}
	return endpoints[failureCount]
}

func RemoveStakeFromChain(poolId uint64, assetIds []uint64, userName eos.AccountName) (string, error) {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return "", err
	}

	privateKey := os.Getenv("PRIVATE_KEY")

	apiURL := GetAPIEndpoint(0)
	api := eos.New(apiURL)

	keyBag := &eos.KeyBag{}
	err = keyBag.ImportPrivateKey(context.Background(), privateKey)
	if err != nil {
		fmt.Println("Error importing private key: ", err)
		return "", err
	}
	api.SetSigner(keyBag)
	txOpts := &eos.TxOptions{}
	if err := txOpts.FillFromChain(context.Background(), api); err != nil {
		fmt.Println("Error filling tx opts: ", err)
		return "", err
	}

	tx := eos.NewTransaction([]*eos.Action{stakeRemoveAction(poolId, assetIds, userName)}, txOpts)
	_, packedTx, err := api.SignTransaction(context.Background(), tx, txOpts.ChainID, eos.CompressionNone)
	if err != nil {
		fmt.Println("Error signing transaction: ", err)
		return "", err
	}

	response, err := api.PushTransaction(context.Background(), packedTx)
	if err != nil {
		fmt.Println("Error pushing transaction: ", err)
		return "", err
	}

	return response.TransactionID, nil

}

func MakeRewards(poolId uint64, api *eos.API) (string, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return "", err
	}

	privateKey := os.Getenv("PRIVATE_KEY")

	keyBag := &eos.KeyBag{}
	err = keyBag.ImportPrivateKey(context.Background(), privateKey)
	if err != nil {
		fmt.Println("Error importing private key: ", err)
		return "", err
	}
	api.SetSigner(keyBag)
	txOpts := &eos.TxOptions{}
	if err := txOpts.FillFromChain(context.Background(), api); err != nil {
		fmt.Println("Error filling tx opts: ", err)
		return "", err
	}

	tx := eos.NewTransaction([]*eos.Action{makeRewardAction(poolId)}, txOpts)
	_, packedTx, err := api.SignTransaction(context.Background(), tx, txOpts.ChainID, eos.CompressionNone)
	if err != nil {
		fmt.Println("Error signing transaction: ", err)
		return "", err
	}

	response, err := api.PushTransaction(context.Background(), packedTx)
	if err != nil {
		fmt.Println("Error pushing transaction: ", err)
		return "", err
	}

	return response.TransactionID, nil

}

func stakeRemoveAction(poolId uint64, assetIds []uint64, userName eos.AccountName) *eos.Action {

	action := &eos.Action{
		Account: eos.AN("breederstake"),
		Name:    eos.ActN("stakeremove"),
		Authorization: []eos.PermissionLevel{
			{Actor: eos.AN(adminAccount), Permission: eos.PN("active")},
		},
		ActionData: eos.NewActionData(StakeRemove{
			PoolId:   poolId,
			AssetIds: assetIds,
			Owner:    userName,
		}),
	}

	return action
}

func makeRewardAction(poolId uint64) *eos.Action {

	action := &eos.Action{
		Account: eos.AN("breederstake"),
		Name:    eos.ActN("makereward"),
		Authorization: []eos.PermissionLevel{
			{Actor: eos.AN(adminAccount), Permission: eos.PN("active")},
		},
		ActionData: eos.NewActionData(MakeReward{
			PoolId: poolId,
		}),
	}

	return action
}
