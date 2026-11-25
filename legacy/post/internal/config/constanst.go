package config

import "os"

const (
	TmpStorage     string = "tmp"
	PostsStorage   string = "posts"
	AvatarsStorage string = "avatars"
)

const (
	UserRegisterPrefix     string = "user-register"
	UserChangeAvatarPrefix string = "user-change-avatar-session"
	UserActiveTokenPrefix  string = "user-active-token"
)

var StorageAddress string

func GetStorageAddress() {
	StorageAddress = os.Getenv("STORAGE_ADDRESS")
}
