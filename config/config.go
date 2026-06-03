package config

import (
	"bufio"
	"log"
	"os"
	"strings"
)

func Load() error {
	file, err := os.Open(".env")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			log.Println("Skipping invalid line:", line)
			continue
		}

		if err := os.Setenv(parts[0], parts[1]); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
