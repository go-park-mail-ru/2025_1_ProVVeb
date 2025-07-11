TEST_DIR= .
COVERAGE_TMP=coverage.out.tmp
COVERAGE_OUT=coverage.out
FILES_TO_CLEAN=*.out *.out.tmp *DS_Store 
MOCKS="mocks"


RATE ?= 100
DURATION ?= 30s

TARGETS_FILE = docs/perf_test/get-profile-targets.txt
TARGETS_FILE_CREATE = docs/perf_test/create-user-targets.txt

REPORT_FILE = docs/perf_test/report.txt
PLOT_FILE = docs/perf_test/graph.html
HISTOGRAM_FILE = docs/perf_test/gist.txt

HISTOGRAM_FILE_HTML = docs/perf_test/plot.html

.PHONY: perf_tests make_perf_test test report plot histogram test clean

perf_tests_create:
	clear
	$(MAKE) init_perf_tests
	@echo "Запуск нагрузки..."
	$(MAKE) clean
	$(MAKE) make_perf_test_create
	$(MAKE) report
	$(MAKE) plot
	$(MAKE) histogram
	rm -rf docs/perf_test/bodies

make_perf_test_create:
	@echo "Запуск нагрузки на $(DURATION) с частотой $(RATE) запросов/сек..."
	vegeta attack -targets=$(TARGETS_FILE_CREATE) -rate=$(RATE) -duration=$(DURATION) | tee /tmp/vegeta-test | vegeta report > $(REPORT_FILE)


perf_tests_get:
	clear
	$(MAKE) init_perf_tests
	@echo "Запуск нагрузки..."
	$(MAKE) clean
	$(MAKE) make_perf_test_get
	$(MAKE) report
	$(MAKE) plot
	$(MAKE) histogram
	rm -rf docs/perf_test/bodies

make_perf_test_get:
	@echo "Запуск нагрузки на $(DURATION) с частотой $(RATE) запросов/сек..."
	vegeta attack -targets=$(TARGETS_FILE) -rate=$(RATE) -duration=$(DURATION) | tee /tmp/vegeta-test | vegeta report > $(REPORT_FILE)

report:
	@echo "Генерация текстового отчёта..."
	@cat $(REPORT_FILE)

init_perf_tests:
	@echo "Запуск установки сессий..."
	REDIS_ADDR=213.219.214.83:8010 go run docs/perf_test/main.go

plot:
	@echo "Генерация HTML-графика..."
	@cat /tmp/vegeta-test | vegeta plot > $(PLOT_FILE)
	@echo "Открой файл $(PLOT_FILE) в браузере."
	open $(PLOT_FILE)

histogram:
	@echo "Генерация гистограммы латентности..."
	@cat /tmp/vegeta-test | vegeta report -type=hist[0,10ms,20ms,50ms,100ms,200ms,500ms,1s] > $(HISTOGRAM_FILE)
	@cat $(HISTOGRAM_FILE)

clean:
	@echo "Удаление временных файлов..."
	rm -f $(FILES_TO_CLEAN)
	rm -rf $(MOCKS)
	@echo "Удаление временных и отчётных файлов..."
	@rm -f $(REPORT_FILE) $(PLOT_FILE) $(HISTOGRAM_FILE) /tmp/vegeta-test
	@echo "Очистка завершена."

launch:
	clear 
	@echo "Сейчас пойдет веселье"
	@echo "Копируем docker файл"
	cd ..
	cp backend/docker-compose.yml docker-compose.yml
	@echo "Запускаем docker файл"
	docker compose up --build 
	@echo "Все"

easyjson:
	go generate ./...

test_injection:
	clear 
	@echo "Делаем тесты на SQL инъекции"
	@echo "Улыбнитесь, вашу консоль снимают!"
	@echo "Запускаем тесты"
	./docs/administration/test.sh
	@echo "Закончили тесты"
	@echo "Лови запись тестов"

good:
	@echo "good"

test:
	# @echo "Делаем моки..."
	# mockgen -source=auth_micro/server/service.go -destination=mocks/sessiomockgen -source=auth_micro/server/sessionrepository.go -destination=mocks/session_repo_mock.go -package=mocksn_repo_mock.go -package=mocks
	@echo "Запуск тестов..."
	go test -v -race -coverpkg=./... -coverprofile=$(COVERAGE_TMP) $(TEST_DIR)/...
	@echo "Обработка покрытия..."

	# Добавляем условие для исключения файлов
	cat $(COVERAGE_TMP) | grep -vE 'github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/server/postgres_con.go|github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/tests/mock.go|github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/usecase/|github.com/go-park-mail-ru/2025_1_ProVVeb/usecase/|github.com/go-park-mail-ru/2025_1_ProVVeb/delivery/handlers.go|github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/tests/mock.go|github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery/profiles_grpc.pb.go|/mocks/|.proto|.sql|go.mod|go.sum|Dockerfile|docker-compose.yml|github.com/go-park-mail-ru/2025_1_ProVVeb/model/models_easyjson.go|github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/delivery/users_grpc.pb.go|github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/delivery/users.pb.go|github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery/profiles.pb.go' > $(COVERAGE_OUT) && rm $(COVERAGE_TMP)

	go tool cover -func=$(COVERAGE_OUT)

	@echo "Удаление временных файлов..."
	rm -f $(FILES_TO_CLEAN)
	# rm -rf $(MOCKS)
	@echo "Тесты завершены"

mocks:
	@echo "Делаем моки..."
	mockgen -source=auth_micro/server/service.go -destination=mocks/sessiomockgen -source=auth_micro/server/sessionrepository.go -destination=mocks/session_repo_mock.go -package=mocksn_repo_mock.go -package=mocks
	mockgen -source=query_micro/server/postgres_con.go.go -destination=mocks/querymockgen -source=auth_micro/server/sessionrepository.go -destination=mocks/query_repo.go -package=mocksn_repo_mock.go -package=mocks
	@echo "Лови моки"

