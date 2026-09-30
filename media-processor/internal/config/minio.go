package config

import (
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func GetMinIOConn() *minio.Client {
	host := os.Getenv("MINIO_HOST")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	privateKey := os.Getenv("MINIO_SECRET_KEY")

	conn, err := minio.New(host, &minio.Options{
		Creds: credentials.NewStaticV4(accessKey, privateKey, ""),
	})

	if err != nil {
		log.Fatal("Err connect object storage: ", err.Error())
	}
	return conn
}
