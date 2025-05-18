package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/icrowley/fake"
)

const (
	count          = 100000
	targetsDir     = "docs/perf_test"
	createUserFile = "create-user-targets.txt"
	getProfileFile = "get-profile-targets.txt"
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
	file, err := os.Create(fmt.Sprintf("%s/%s", targetsDir, createUserFile))
	if err != nil {
		return fmt.Errorf("не удалось создать файл: %w", err)
	}
	defer file.Close()

	for i := 1; i <= count; i++ {
		sessionKey, userID := fmt.Sprintf("session_%d", i), fmt.Sprintf("%d", i%count)
		if err := ensureSession(ctx, client, sessionKey, userID); err != nil {
			fmt.Printf("Ошибка при работе с Redis: %v\n", err)
			continue
		}

		body := generateUserProfile(i)
		jsonBody, err := json.Marshal(body)
		if err != nil {
			fmt.Printf("Ошибка сериализации JSON: %v\n", err)
			continue
		}

		jsonStr := strings.TrimSpace(string(jsonBody))
		adress := "localhost"
		createTarget := fmt.Sprintf(
			"POST http://%s:8080/users\r\nContent-Type: application/json\r\nCookie: session_id=%s; csrf_token=dummy_csrf\r\nX-CSRF-Token: dummy_csrf\r\n%s\r\n",
			adress, sessionKey, jsonStr)

		if _, err := file.WriteString(createTarget + "\r\n"); err != nil {
			fmt.Printf("POST error: %v\n", err)
		}
	}

	fmt.Printf("Создан файл %s с %d запросами\n", createUserFile, count)
	return nil
}

func generateGetProfiles(ctx context.Context, client *redis.Client) error {
	file, err := os.Create(fmt.Sprintf("%s/%s", targetsDir, getProfileFile))
	if err != nil {
		return fmt.Errorf("не удалось создать файл: %w", err)
	}
	defer file.Close()

	for i := 1; i <= count; i++ {
		sessionKey, userID := fmt.Sprintf("session_%d", i), fmt.Sprintf("%d", i%count)
		if err := ensureSession(ctx, client, sessionKey, userID); err != nil {
			fmt.Printf("Ошибка при работе с Redis: %v\n", err)
			continue
		}

		adress := "213.219.214.83"
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

func generateUserProfile(i int) map[string]interface{} {
	login := fmt.Sprintf("user%d_%s", i, fake.FirstName())
	email := fmt.Sprintf("%d_%s", i, fake.EmailAddress())
	phone := fmt.Sprintf("+1774707777%d", i)

	return map[string]interface{}{
		"user": map[string]interface{}{
			"login":    login,
			"password": "password123",
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
