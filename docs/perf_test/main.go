package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/icrowley/fake"
)

const (
	count          = 1000
	cookie_cnt     = 1000
	targetsDir     = "docs/perf_test"
	createUserFile = "create-user-targets.txt"
	getProfileFile = "get-profile-targets.txt"
	adress         = "213.219.214.83"
)

func main() {
	ctx := context.Background()

	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: "",
		DB:       0,
	})

	if err := os.MkdirAll(targetsDir, os.ModePerm); err != nil {
		panic(fmt.Errorf("не удалось создать директорию %s: %w", targetsDir, err))
	}

	if !isFileEmpty(targetsDir + "/" + createUserFile) {
		fmt.Println("Файл создания пользователей не пустой — пропускаем генерацию.")
	} else {
		if err := generateCreateUsers(ctx, client); err != nil {
			fmt.Printf("Ошибка при генерации create-user: %v\n", err)
		}
	}

	if !isFileEmpty(targetsDir + "/" + getProfileFile) {
		fmt.Println("Файл получения профилей не пустой — пропускаем генерацию.")
	} else {
		if err := generateGetProfiles(ctx, client); err != nil {
			fmt.Printf("Ошибка при генерации get-profile: %v\n", err)
		}
	}
}

func isFileEmpty(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return true
	}
	return info.Size() == 0
}

func generateCreateUsers(ctx context.Context, client *redis.Client) error {
	bodiesDir := fmt.Sprintf("%s/bodies", targetsDir)
	if err := os.MkdirAll(bodiesDir, os.ModePerm); err != nil {
		return fmt.Errorf("не удалось создать директорию с JSON: %w", err)
	}

	targetsPath := fmt.Sprintf("%s/%s", targetsDir, createUserFile)
	targetsFile, err := os.Create(targetsPath)
	if err != nil {
		return fmt.Errorf("не удалось создать файл целей: %w", err)
	}
	defer targetsFile.Close()

	for i := 1; i <= count; i++ {
		body := generateUserProfile(i)
		jsonBody, err := json.MarshalIndent(body, "", "  ")
		if err != nil {
			fmt.Printf("Ошибка сериализации JSON: %v\n", err)
			continue
		}

		jsonFilename := fmt.Sprintf("user_%05d.json", i)
		jsonPath := fmt.Sprintf("%s/%s", bodiesDir, jsonFilename)
		if err := os.WriteFile(jsonPath, jsonBody, 0644); err != nil {
			fmt.Printf("Ошибка записи JSON: %v\n", err)
			continue
		}

		target := fmt.Sprintf("POST http://%s:8080/users\n@%s\n", adress, jsonPath)
		if _, err := targetsFile.WriteString(target); err != nil {
			fmt.Printf("Ошибка записи в targets файл: %v\n", err)
		}
	}

	fmt.Printf("Создано %d JSON-профилей в %s и файл целей %s\n", count, bodiesDir, createUserFile)
	return nil
}

func generateGetProfiles(ctx context.Context, client *redis.Client) error {
	file, err := os.Create(fmt.Sprintf("%s/%s", targetsDir, getProfileFile))
	if err != nil {
		return fmt.Errorf("не удалось создать файл: %w", err)
	}
	defer file.Close()

	for i := 1; i <= count; i++ {
		sessionKey, userID := fmt.Sprintf("session_%d", i%cookie_cnt), fmt.Sprintf("%d", i%cookie_cnt)
		if err := ensureSession(ctx, client, sessionKey, userID); err != nil {
			fmt.Printf("Ошибка при работе с Redis: %v\n", err)
			continue
		}

		getTarget := fmt.Sprintf(
			"GET http://%s:8080/profiles\r\nCookie: session_id=%s; csrf_token=dummy_csrf\r\nX-CSRF-Token: dummy_csrf\r\n",
			adress, sessionKey)
		if _, err := file.WriteString(getTarget + "\r\n"); err != nil {
			fmt.Printf("Ошибка записи GET: %v\n", err)
		}
	}

	fmt.Printf("Создан файл %s с %d запросами\n", getProfileFile, count)
	return nil
}

func ensureSession(ctx context.Context, client *redis.Client, sessionKey, userID string) error {
	exists, err := client.Exists(ctx, sessionKey).Result()
	if err != nil {
		return err
	}

	if exists == 0 {
		if err := client.Set(ctx, sessionKey, userID, 24*time.Hour).Err(); err != nil {
			return err
		}
		fmt.Printf("Добавлена новая сессия: %s = %s\n", sessionKey, userID)
	}
	return nil
}
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func generateUserProfile(i int) map[string]interface{} {
	login := fmt.Sprintf("user%d_%s", i, fake.FirstName())
	login = truncate(login, 255)

	emailPrefix := fmt.Sprintf("%d", i)
	emailSuffix := fake.EmailAddress()
	if len(emailPrefix)+len(emailSuffix) > 254 {
		emailSuffix = truncate(emailSuffix, 254-len(emailPrefix))
	}
	email := emailPrefix + "_" + emailSuffix
	email = truncate(email, 255)

	phone := fmt.Sprintf("+1%09d", i)

	password := "password123"

	return map[string]interface{}{
		"user": map[string]interface{}{
			"login":    login,
			"password": password,
			"email":    email,
			"phone":    phone,
			"status":   1,
		},
		"profile": map[string]interface{}{
			"firstName":   fake.FirstName(),
			"lastName":    fake.LastName(),
			"isMale":      true,
			"height":      160 + i%30,
			"birthday":    "1990-01-01T00:00:00Z",
			"description": fake.Sentence(),
			"location":    fmt.Sprintf("%s@%s@%s", fake.Country(), fake.City(), fake.City()),
			"interests":   []string{fake.Word(), fake.Word()},
			"likedBy":     []string{},
			"preferences": []map[string]string{
				{
					"preference_description": "smoking",
					"preference_value":       "no",
				},
			},
			"photos": []string{},
		},
	}
}
