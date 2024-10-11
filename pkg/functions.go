package pkg

import (
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
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
		CertificateExportPath:     os.Getenv(EnvServerCertificateExportPath),
	}, nil
}

func (p PriorityEnum) String() string {
	if p {
		return "PriorityHigh"
	}

	return "PriorityLow"
}

func InitializeRedisClient(address string, password string, db int) *redis.Client {
	if address == "" {
		return nil
	}

	return redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})
}

func getEnvWithDefault(envVar, defaultValue string) string {
	value := os.Getenv(envVar)
	if value == "" {
		return defaultValue
	}

	return value
}
