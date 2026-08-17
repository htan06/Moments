package config

import "os"

type CacheNamespace string

const (
	UserRegisterPrefix     CacheNamespace = "user-register"
	UserChangeAvatarPrefix CacheNamespace = "user-change-avatar-session"
	UserActiveTokenPrefix  CacheNamespace = "user-active-token"

	UserUploadPostPrefix CacheNamespace = "user-upload-post-session"
)

type StorageNamespace string

var StorageAddress StorageNamespace

func GetStorageAddress() {
	StorageAddress = StorageNamespace(os.Getenv("STORAGE_ADDRESS"))
}

const (
	AvatarBucket StorageNamespace = "avatars"
	PostBucket   StorageNamespace = "posts"
	TempBucket   StorageNamespace = "tmp"
)
