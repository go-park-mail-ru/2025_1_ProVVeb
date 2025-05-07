TEST_DIR= ./tests
COVERAGE_TMP=coverage.out.tmp
COVERAGE_OUT=coverage.out
FILES_TO_CLEAN=*.out *.out.tmp *DS_Store 
MOCKS="mocks"

test:
	clear
	@echo "Делаем моки..."
	mockgen -source=auth_micro/server/service.go -destination=mocks/sessiomockgen -source=auth_micro/server/sessionrepository.go -destination=mocks/session_repo_mock.go -package=mocksn_repo_mock.go -package=mocks
	@echo "Запуск тестов..."
	go test -v -race -coverpkg=./... -coverprofile=$(COVERAGE_TMP) $(TEST_DIR)/...
	@echo "Обработка покрытия..."

	# Добавляем условие для исключения файлов
	cat $(COVERAGE_TMP) | grep -vE '/mocks/|.proto|.sql|go.mod|go.sum|Dockerfile|docker-compose.yml|github.com/go-park-mail-ru/2025_1_ProVVeb/repository/user_repository.go|github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/server/service.go|github.com/go-park-mail-ru/2025_1_ProVVeb/auth_micro/server/service.go|github.com/go-park-mail-ru/2025_1_ProVVeb/repository/static_repository.go|github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/server/postgres_con.go|github.com/go-park-mail-ru/2025_1_ProVVeb/repository/chat_repository.go' > $(COVERAGE_OUT) && rm $(COVERAGE_TMP)

	go tool cover -func=$(COVERAGE_OUT)

	@echo "Удаление временных файлов..."
	rm -f $(FILES_TO_CLEAN)
	rm -rf $(MOCKS)
	@echo "Тесты завершены"

clean:
	@echo "Удаление временных файлов..."
	rm -f $(FILES_TO_CLEAN)
	rm -rf $(MOCKS)
	@echo "Очистка завершена."

mocks:
	@echo "Делаем моки..."
	mockgen -source=auth_micro/server/service.go -destination=mocks/sessiomockgen -source=auth_micro/server/sessionrepository.go -destination=mocks/session_repo_mock.go -package=mocksn_repo_mock.go -package=mocks
	mockgen -source=query_micro/server/postgres_con.go.go -destination=mocks/querymockgen -source=auth_micro/server/sessionrepository.go -destination=mocks/query_repo.go -package=mocksn_repo_mock.go -package=mocks
	@echo "Лови моки"

.PHONY: test clean
