TEST_DIR= ./tests
COVERAGE_TMP=coverage.out.tmp
COVERAGE_OUT=coverage.out

test:
	clear
	@echo "Запуск тестов..."
	go test -v -race -coverpkg=./... -coverprofile=$(COVERAGE_TMP) $(TEST_DIR)/...
	@echo "Обработка покрытия..."
	cat $(COVERAGE_TMP) | grep -v "mocks" > $(COVERAGE_OUT) && rm $(COVERAGE_TMP)
	go tool cover -func=$(COVERAGE_OUT)
	@echo "Тесты завершены"

clean:
	rm -f $(COVERAGE_OUT)

.PHONY: test clean
