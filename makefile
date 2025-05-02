TEST_DIR= ./tests
COVERAGE_TMP=coverage.out.tmp
COVERAGE_OUT=coverage.out
FILES_TO_CLEAN=*.out *.out.tmp *DS_Store 
MOCKS="mocks"

test:
	clear
	@echo "Делаем моки..."
	mockgen -source=repository/userrepository.go -destination=mocks/mock_repository.go -package=mocks
	@echo "Запуск тестов..."
	go test -v -race -coverpkg=./... -coverprofile=$(COVERAGE_TMP) $(TEST_DIR)/...
	@echo "Обработка покрытия..."
	cat $(COVERAGE_TMP) | grep -v $(MOCKS) > $(COVERAGE_OUT) && rm $(COVERAGE_TMP)
	go tool cover -func=$(COVERAGE_OUT)
	@echo "Удаление временных файлов..."
	rm -f $(FILES_TO_CLEAN)
	rm -rf $(MOCKS)
	@echo "Тесты завершены"

clean:
	@echo "Удаление временных файлов..."
	rm -f $(FILES_TO_CLEAN)
	rm -d $(MOCKS)
	@echo "Очистка завершена."

mocks:
	@echo "Делаем моки..."
	mockgen -source=repository/userrepository.go -destination=mocks/mock_repository.go -package=mocks
	@echo "Лови моки"


.PHONY: test clean
