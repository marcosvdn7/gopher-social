package env

import (
	"log"
	"os"
	"strconv"
)

func GetString(key, fallback string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		log.Printf("log key[%s] not found, using fallback [%s]\n", key, fallback)
		return fallback
	}

	return v
}

func GetInt(key string, fallback int) int {
	v := GetString(key, strconv.FormatInt(int64(fallback), 36))

	vAsInt, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}

	return vAsInt
}

func GetBool(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok {
		log.Printf("log key[%s] not found, using fallback [%t]\n", key, fallback)
		return fallback
	}

	boolValue, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}

	return boolValue
}
