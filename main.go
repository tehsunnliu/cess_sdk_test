package main

import (
	"context"
	"errors"
	"example/cess_go_sdk/logger"
	"math/rand"
	"os"
	"time"

	cess "github.com/CESSProject/cess-go-sdk"
	"github.com/CESSProject/cess-go-sdk/config"
	"github.com/CESSProject/cess-go-sdk/core/pattern"
	"github.com/CESSProject/cess-go-sdk/core/sdk"
	"github.com/CESSProject/cess-go-sdk/core/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Substrate well-known mnemonic:
//
//	https://github.com/substrate-developer-hub/substrate-developer-hub.github.io/issues/613
var MY_MNEMONIC = "<ENTER_YOUR_SEED_HERE>"

var RPC_ADDRS = []string{
	"wss://testnet-rpc0.cess.cloud/ws/",
	"wss://testnet-rpc1.cess.cloud/ws/",
}

var GatewayURL = "http://deoss-pub-gateway.cess.cloud/" // Public Gateway
// var GatewayURL = "http://127.0.0.1:8080/" // Self hosted Gateway

var GatewayAccAddress = "cXhwBytXqrZLr1qM5NHJhCzEMckSTzNKw17ci2aHft6ETSQm9" // Public Gateway
// var GatewayAccAddress = "cXiHsknbhePZEwxM92dEFzBNG9q2MkRoASXj5NczdWUDcrEzx" // Self hosted Gateway

const Path = "./TEST_FILES"
const Workspace = "./CESS_STORAGE"
const FileName = "rand.txt"

const BucketName = "random"
const FileSize1MB = 1 * 1024 * 1024 // 1MB
const MinFileSize = 1
const MaxFileSize = 10

var Port = 4002

var Bootstrap = []string{
	"_dnsaddr.boot-kldr-testnet.cess.cloud", // Testnet
	// "_dnsaddr.bootstrap-kldr.cess.cloud", // Devnet
}

func main() {
	if _, err := os.Stat(Workspace); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(Workspace, os.ModePerm)
		if err != nil {
			logger.Log.Println(err)
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

	createBucket(sdk, keyringPair.PublicKey)

	_, err = sdk.AuthorizeSpace(GatewayAccAddress)
	if err != nil {
		panic(err)
	}

	logger.Log.Println("Gateway: ", GatewayURL)
	for {
		loc, _ := time.LoadLocation("Asia/Kolkata")
		logger.Log.Println("--------------Uploading File - " + time.Now().In(loc).String() + " --------------")

		fileUrl := generateFile()
		fileHash := uploadFile(sdk, fileUrl)
		saveFileHash(fileHash, FileName)
		verifyUploadAndDownloadFile(sdk, keyringPair.PublicKey, fileHash, FileName)

		loc, _ = time.LoadLocation("Asia/Kolkata")
		logger.Log.Println("--------------Completed - " + time.Now().In(loc).String() + " --------------")
	}

	// for _, fileName := range FileNames {
	// 	logger.Log.Println("--------------Uploading " + fileName + " File--------------")
	// 	fileHash := uploadFile(sdk, Path+"/"+fileName)
	// 	saveFileHash(fileHash, fileName)
	// 	verifyUploadAndDownloadFile(sdk, keyringPair.PublicKey, fileHash, fileName)
	// 	logger.Log.Println("--------------" + fileName + " Completed--------------")
	// }
}

func generateFile() string {
	fileSize := (rand.Intn(MaxFileSize-MinFileSize) + MinFileSize) * FileSize1MB

	logger.Log.Println("FileSize:", fileSize/(1024*1024), "MB")
	data := RandStringBytes(fileSize)
	fileUrl := Workspace + "/" + FileName
	// Store File hashes in a file for future reference.
	myfile, err := os.OpenFile(fileUrl, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer myfile.Close()

	// Write the string to the file
	_, err = myfile.WriteString(data)
	if err != nil {
		panic(err)
	}
	return fileUrl
}

func createBucket(sdk sdk.SDK, publicKey []byte) {
	if !utils.CheckBucketName(BucketName) {
		panic("invalid bucket name")
	}

	bucketList, err := sdk.QueryAllBucketName(publicKey)
	if err != nil {
		panic(err)
	}

	if !containsBucket(bucketList, BucketName) {
		logger.Log.Println("Creating bucket...")
		tx, err := sdk.CreateBucket(publicKey, BucketName)
		if err != nil {
			panic(err)
		}

		logger.Log.Println("Bucket ID: ", tx)
	}
}

func verifyUploadAndDownloadFile(sdk sdk.SDK, publicKey []byte, fileHash string, fileName string) {
	minerUploadTime := time.Now()

	for {
		bucketInfo, err := sdk.QueryBucketInfo(publicKey, BucketName)
		if err != nil {
			panic(err)
		}

		if containsFilehash(bucketInfo.ObjectsList, fileHash) {
			logger.Log.Println("File uploaded to Miners in: ", time.Since(minerUploadTime))

			downloadFile(sdk, fileHash, fileName)
			break
		} else {
			_, err := sdk.QueryStorageOrder(fileHash)
			if err != nil {
				start := time.Now()
				for {
					time.Sleep(1 * time.Second)
					_, err := sdk.QueryStorageOrder(fileHash)
					if err == nil {
						logger.Log.Println("Deal found in: ", time.Since(start))
						break
					}
				}
			}

		}
		// Hash not found try again after 10 sec
		time.Sleep(10 * time.Second)
	}
}

func uploadFile(sdk sdk.SDK, fileUrl string) string {

	start := time.Now()

	fileHash, err := sdk.UploadtoGateway(GatewayURL, fileUrl, BucketName)
	if err != nil {
		logger.Log.Println(err)
		panic(err)
	}
	logger.Log.Println("FID:", fileHash)
	logger.Log.Println("File uploaded to Gateway in: ", time.Since(start))
	return fileHash
}

func downloadFile(sdk sdk.SDK, fileHash string, fileName string) {
	logger.Log.Println("Downloading File...")
	start := time.Now()

	err := sdk.DownloadFromGateway(GatewayURL, fileHash, Workspace+"/"+fileHash+fileName)

	if err != nil {
		panic(err)
	}
	logger.Log.Println("File dwonloaded in: ", time.Since(start))
}

func saveFileHash(fileHash string, fileName string) {
	// Store File hashes in a file for future reference.
	myfile, err := os.OpenFile(Workspace+"/filehashes.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer myfile.Close()

	loc, _ := time.LoadLocation("Asia/Kolkata")

	// Write the string to the file
	_, err = myfile.WriteString(fileHash + " " + fileName + " " + time.Now().In(loc).String() + "\n")
	if err != nil {
		panic(err)
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

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
