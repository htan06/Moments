package config

import "os"

type CacheNamespace string

const (
	UserRegisterPrefix     CacheNamespace = "user-register"
	UserChangeAvatarPrefix CacheNamespace = "user-change-avatar-session"
	UserActiveTokenPrefix  CacheNamespace = "user-active-token"
)

type StorageNamespace string

var StorageAddress string

func GetStorageAddress() {
	StorageAddress = os.Getenv("STORAGE_ADDRESS")
}

const (
	AvatarBucket StorageNamespace = "avatars"
	PostBucket   StorageNamespace = "posts"
	TempBucket   StorageNamespace = "tmp"
)
