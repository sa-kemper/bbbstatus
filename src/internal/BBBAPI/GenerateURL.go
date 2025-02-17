package BBBAPI

import (
	"crypto/sha1"
	"fmt"
)

func GenerateURL(config URLConfig) string {
	result := fmt.Sprintf("https://%s/%s?", config.Hostname, config.Methode)
	for key, value := range config.Parameters {
		result += fmt.Sprintf("%s=%s&", key, value)
	}
	checksum := sha1.New()
	checksum.Write([]byte(result + config.SharedSecret))
	return fmt.Sprintf("%s&checksum=%s", result, checksum.Sum(nil))
}
