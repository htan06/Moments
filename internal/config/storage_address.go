package config

import "os"

var	StorageAddress string

func GetStorageAddress() {
	StorageAddress  = os.Getenv("STORAGE_ADDRESS")
}