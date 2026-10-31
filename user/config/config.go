package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

type Config struct {
	AppConfig         *ApplicationConfig
	AuthConfig        *AuthenticationConfig
	RustFSConfig      *aws.Config
	PostgresConfig    *pgxpool.Config
	RedisConfig       *RedisConfig
	KafkaWriterConfig *kafka.WriterConfig
	// KafkaReaderConfig *kafka.ReaderConfig
	StorageAddress    string
}

type AuthenticationConfig struct {
	JWTPublicKey *rsa.PublicKey
}

type ApplicationConfig struct {
	Address string
}

type RustFSConfig struct {
	Address         string
	AccessKeyID     string
	SecretAccessKey string
}

type PostgresConfig struct {
	URI      string
	MinConns int
	MaxConns int
}

type RedisConfig struct {
	Address  string
	Password string
}

func LoadConfig() (*Config, error) {
	appCfg, err := loadApplicationConfig()
	if err != nil {
		return nil, err
	}

	authCfg, err := loadAuthenticationConfig()
	if err != nil {
		return nil, err
	}

	rustFSCfg, err := loadRustFSConfig()
	if err != nil {
		return nil, err
	}

	postgresCfg, err := loadPostgresConfig()
	if err != nil {
		return nil, err
	}

	redisCfg, err := loadRedisConfig()
	if err != nil {
		return nil, err
	}

	storageAdress, err := loadStorageAddress()
	if err != nil {
		return nil, err
	}

	kafkaWriterCfg, err := loadKafkaWriterConfig()
	if err != nil {
		return nil, err
	}

	// kafkaReaderCfg, err := loadKafkaReaderConfig()
	// if err != nil {
	// 	return nil, err
	// }

	return &Config{
		AppConfig:         appCfg,
		AuthConfig:        authCfg,
		RustFSConfig:      rustFSCfg,
		PostgresConfig:    postgresCfg,
		RedisConfig:       redisCfg,
		StorageAddress:    storageAdress,
		KafkaWriterConfig: kafkaWriterCfg,
		// KafkaReaderConfig: kafkaReaderCfg,
	}, nil
}

func loadApplicationConfig() (*ApplicationConfig, error) {
	address := os.Getenv("APPLICATION_ADDRESS")
	if !isHostPort(address) {
		return nil, fmt.Errorf("Application address invalid: %s", address)
	}

	return &ApplicationConfig{
		Address: address,
	}, nil
}

func isHostPort(address string) bool {
	host, port, err := net.SplitHostPort(address)
	if err != nil || host == "" || port == "" {
		return false
	}

	p, err := strconv.Atoi(port)
	if err != nil || p < 1 || p > 65535 {
		return false
	}

	return true
}

func loadAuthenticationConfig() (*AuthenticationConfig, error) {
	publicKeyPath := os.Getenv("JWT_PUBLIC_KEY_PATH")
	data, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("Error load jwt public key %v", err.Error())
	}

	block, _ := pem.Decode(data)

	if block.Type != "ACCESS PUBLIC KEY" {
		return nil, fmt.Errorf("Error load jwt public key %v", err.Error())
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("Error load jwt public key %v", err.Error())
	}
	return &AuthenticationConfig{JWTPublicKey: key.(*rsa.PublicKey)}, nil
}

func loadStorageAddress() (string, error) {
	adr := os.Getenv("STORAGE_ADDRESS")
	if adr == "" {
		return "", fmt.Errorf("MISSING VALUE STORAGE_ADDRESS")
	}
	return adr, nil
}

func loadRustFSConfig() (*aws.Config, error) {
	accessKeyID := os.Getenv("RUSTFS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("RUSTFS_SECRET_ACCESS_KEY")
	enpointURL := os.Getenv("RUSTFS_ENDPOINT_URL")

	if accessKeyID == "" || secretAccessKey == "" || enpointURL == "" {
		return nil, fmt.Errorf("missing the env: RUSTFS_ACCESS_KEY_ID / RUSTFS_SECRET_ACCESS_KEY / RUSTFS_ADDRESS")
	}

	return &aws.Config{
		Region:       "us-east-1",
		BaseEndpoint: &enpointURL,
		Credentials:  aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
	}, nil
}

func loadPostgresConfig() (*pgxpool.Config, error) {
	uri := os.Getenv("DATABASE_POSTGRES_URI")

	maxConns, err := strconv.ParseInt(os.Getenv("DATABASE_POSTGRES_MAX_CONNS"), 10, 32)
	if err != nil || maxConns == 0 {
		return nil, fmt.Errorf("GetPostgres config error: %v", err.Error())
	}

	minConns, err := strconv.ParseInt(os.Getenv("DATABASE_POSTGRES_MIN_CONNS"), 10, 32)
	if err != nil || minConns == 0 {
		return nil, fmt.Errorf("GetPostgres config error: %v", err.Error())
	}

	config, err := pgxpool.ParseConfig(uri)
	if err != nil {
		log.Fatal("Postgres config error: " + err.Error())
	}

	config.MaxConns = int32(maxConns)
	config.MinConns = int32(minConns)

	return config, nil
}

func loadRedisConfig() (*RedisConfig, error) {
	address := os.Getenv("REDIS_ADDRESS")
	password := os.Getenv("REDIS_PASSWORD")

	return &RedisConfig{
		Address:  address,
		Password: password,
	}, nil
}

// func loadKafkaReaderConfig() (*kafka.ReaderConfig, error) {
// 	brokers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")

// 	return &kafka.ReaderConfig{
// 		Brokers: []string{brokers},
// 	}, nil
// }

func loadKafkaWriterConfig() (*kafka.WriterConfig, error) {
	brokers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")

	return &kafka.WriterConfig{
		Brokers: []string{brokers},
	}, nil
}
