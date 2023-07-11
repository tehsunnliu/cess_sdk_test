package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	cess "github.com/CESSProject/cess-go-sdk"
	"github.com/CESSProject/cess-go-sdk/config"
	"github.com/CESSProject/cess-go-sdk/core/pattern"
	"github.com/CESSProject/cess-go-sdk/core/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// Substrate well-known mnemonic:
//
//	https://github.com/substrate-developer-hub/substrate-developer-hub.github.io/issues/613
var MY_MNEMONIC = "<ENTER_YOUR_MNEMONIC_HERE>"

var RPC_ADDRS = []string{
	"wss://testnet-rpc0.cess.cloud/ws/",
	"wss://testnet-rpc1.cess.cloud/ws/",
}

const BucketName = "bucket0"
const Path = "./TEST_FILES/"

// const FileName = "153_9mb.mp4"
const FileName = "8_4mb.jpg"

var Workspace = "./CESS_STORAGE"
var Port = 4001
var Bootstrap = []string{
	"_dnsaddr.boot-kldr-testnet.cess.cloud", // Testnet
	// "_dnsaddr.bootstrap-kldr.cess.cloud", // Devnet
}

func main() {
	if _, err := os.Stat(Workspace); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(Workspace, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}

	sdk, err := cess.New(
		context.Background(),
		config.CharacterName_Client,
		cess.ConnectRpcAddrs(RPC_ADDRS),
		cess.Mnemonic(MY_MNEMONIC),
		cess.TransactionTimeout(time.Second*10),
		cess.Workspace(Workspace),
		cess.P2pPort(Port),
		cess.Bootnodes(Bootstrap),
	)
	if err != nil {
		panic(err)
	}

	keyringPair, _ := signature.KeyringPairFromSecret(MY_MNEMONIC, 0)

	if !utils.CheckBucketName(BucketName) {
		panic("invalid bucket name")
	}

	// Create Bucket
	bucketList, _ := sdk.QueryAllBucketName(keyringPair.PublicKey)
	if !containsBucket(bucketList, BucketName) {
		fmt.Println("Creating bucket...")
		fmt.Println(sdk.CreateBucket(keyringPair.PublicKey, BucketName))
	}

	// Upload File
	fmt.Println("Uploading File...")
	start := time.Now()
	fileHash, _ := sdk.StoreFile(keyringPair.PublicKey, Path+FileName, BucketName)
	fmt.Println(fileHash, ", Uploaded in: ", time.Since(start))

	// Download File
	fmt.Println("Querying File...")
	start = time.Now()
	fmt.Println(sdk.RetrieveFile(fileHash, Workspace+"/downloaded"+FileName), ", Request Completed in: ", time.Since(start))

	storageOrder, _ := sdk.QueryStorageOrder(fileHash)
	fmt.Println("Storage Order: ", storageOrder)

	start = time.Now()
	for {
		fmt.Println("Querrying Bucket Info...")
		bucketInfo, _ := sdk.QueryBucketInfo(keyringPair.PublicKey, BucketName)

		if containsFilehash(bucketInfo.ObjectsList, fileHash) {
			fmt.Println("Filehash found in: ", time.Since(start))
			break
		}
		fmt.Println("Filehash not found!")
		time.Sleep(10 * time.Second)
	}
}

func containsFilehash(s []pattern.FileHash, str string) bool {
	var hash pattern.FileHash
	for i := 0; i < len(hash); i++ {
		hash[i] = types.U8(str[i])
	}

	for _, v := range s {
		if v == hash {
			return true
		}
	}

	return false
}

func containsBucket(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
