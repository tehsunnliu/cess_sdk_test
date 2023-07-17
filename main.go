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
const Path = "./TEST_FILES"

// const FileName = "154MB.mp4"
const FileName = "8MB.jpg"

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

	keyringPair, err := signature.KeyringPairFromSecret(MY_MNEMONIC, 0)
	if err != nil {
		panic(err)
	}

	if !utils.CheckBucketName(BucketName) {
		panic("invalid bucket name")
	}

	// Create Bucket
	bucketList, err := sdk.QueryAllBucketName(keyringPair.PublicKey)
	if err != nil {
		panic(err)
	}

	if !containsBucket(bucketList, BucketName) {
		fmt.Println("Creating bucket...")
		fmt.Println(sdk.CreateBucket(keyringPair.PublicKey, BucketName))
	}

	// Upload File
	fmt.Println("Uploading File...")
	start := time.Now()
	fileHash, err := sdk.StoreFile(Path+"/"+FileName, BucketName)
	fmt.Println("Error: ", err)
	fmt.Println(fileHash, ", Uploaded in: ", time.Since(start))

	// Store File hashes in a file for future reference.
	myfile, err := os.OpenFile(Workspace+"/filehashes.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer myfile.Close()

	loc, _ := time.LoadLocation("Asia/Kolkata")

	// Write the string to the file
	_, err = myfile.WriteString(fileHash + " " + time.Now().In(loc).String() + "\n")
	if err != nil {
		fmt.Println(err)
		return
	}

	start = time.Now()

	fmt.Println("Querrying Bucket Info...")
	for {
		bucketInfo, err := sdk.QueryBucketInfo(keyringPair.PublicKey, BucketName)
		if err != nil {
			panic(err)
		}

		if containsFilehash(bucketInfo.ObjectsList, fileHash) {
			fmt.Println("File Uploaded in: ", time.Since(start))

			// Download File
			fmt.Println("Downloading File...")
			start = time.Now()
			err := sdk.RetrieveFile(fileHash, Workspace+"/"+fileHash+FileName)
			if err != nil {
				panic(err)
			}
			fmt.Println("File Dwonloaded in: ", time.Since(start))

			break
		} else {
			storageOrder, err := sdk.QueryStorageOrder(fileHash)
			fmt.Println("File hash: ", fileHash)
			fmt.Println("Storage Order: ", storageOrder)
			if err != nil {
				fmt.Println("Failed to query StorageDeal ", err)
				break
			}
		}
		// Hash not found try again after 10 sec
		time.Sleep(30 * time.Second)
	}
}

func containsFilehash(hashes []pattern.FileHash, hash string) bool {
	var h pattern.FileHash
	for i := 0; i < len(h); i++ {
		h[i] = types.U8(hash[i])
	}

	for _, v := range hashes {
		if v == h {
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
