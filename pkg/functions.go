package pkg

import (
	"context"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func LoadConfig() (*Config, error) {
	redisDatabase, err := strconv.Atoi(os.Getenv(EnvRedisServerDatabase))
	if err != nil {
		redisDatabase = 0
	}

	return &Config{
		RedisServerAddress:        os.Getenv(EnvRedisServerAddress),
		RedisServerPassword:       os.Getenv(EnvRedisServerPassword),
		RedisServerDatabase:       redisDatabase,
		ServerAddress:             getEnvWithDefault(EnvServerAddress, ":8443"),
		CertificateOrganization:   getEnvWithDefault(EnvServerCertificateOrganization, "be-heroes"),
		CertificateCommonName:     getEnvWithDefault(EnvServerCertificateCommonName, "ultron-service.default.svc"),
		CertificateDnsNamesCSV:    getEnvWithDefault(EnvServerCertificateDnsNames, "ultron-service.default.svc,ultron-service,localhost"),
		CertificateIpAddressesCSV: getEnvWithDefault(EnvServerCertificateIpAddresses, "127.0.0.1"),
		CertificateExportPath:     getEnvWithDefault(EnvServerCertificateExportPath, "ultron_ca_cert.pem"),
	}, nil
}

func InitializeRedisClient(address string, password string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})
}

func InitializeRedisClientFromConfig(ctx context.Context, config *Config, sugar *zap.SugaredLogger) *redis.Client {
	redisClient := InitializeRedisClient(config.RedisServerAddress, config.RedisServerPassword, config.RedisServerDatabase)

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		sugar.Fatalf("Failed to connect to Redis server: %v", err)
	}

	return redisClient
}

func getEnvWithDefault(envVar, defaultValue string) string {
	value := os.Getenv(envVar)

	if value == "" {
		return defaultValue
	}

	return value
}

func ParseCsvIpAddressString(csv string) []net.IP {
	var certificateIpAddresses []net.IP

	for _, ipAddress := range strings.Split(csv, ",") {
		if ipAddress != "" {
			certificateIpAddresses = append(certificateIpAddresses, net.ParseIP(ipAddress))
		}
	}

	return certificateIpAddresses
}
