# ДЗ - 3 Оптимизация работы СУБД


## Подготовка

### Выбор основноый сущности тестирования

В качестве тестирования основной сущности для сайта знакомств была выбрана таблица Profiles и получение профилей

Содержание таблицы:


```sql
CREATE TABLE profiles (
    profile_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    firstname TEXT NOT NULL CHECK (LENGTH(firstname) <= 255),
    lastname TEXT NOT NULL CHECK (LENGTH(lastname) <= 255),
    is_male BOOLEAN NOT NULL,
    birthday DATE NOT NULL,
    height INT CHECK (height >= 50 AND height <= 280),
    description TEXT,
    location_id BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (location_id) REFERENCES locations(location_id) ON DELETE SET NULL ON UPDATE CASCADE
);
```

#### Причины:
- На поиске профилей и свайпах строится основная продуктовая логика приложения
- Для получения профилей, помимо самой таблицы, требуется подгрузить информацию из таблиц static, interests, preferences(также связующие таблицы profile_interests и profile_preferences), likes и matches
- Большую часть время пользователи будут проводить за "свайпами" поэтому каждому пользователю необходимо ускорить отдачу профилей 

#### Характер нагрузки

Тип нагрузки: Чтение и фильтрация большого объема данных

Характер обращений: Часто с параметрами (пол, возраст, рост, интересы и т.д.)

Цель оптимизации: Ускорение выдачи релевантных профилей пользователю с минимальной задержкой

### Выбор утилиты тестирования

Для тестирования была выбрана утилита Vegeta

#### Причины 

- Поддержка различных HTTP-методов, заголовков и тел.
- Управление:
  - Количеством соединений (`-connections`)
  - Таймаутами (`-timeout`)
  - Поддержкой HTTP/2 и Keep-Alive
  Информативные отчёты
- Форматы:
  - `text`
  - `json`
  - `html` (график)
- Метрики:
  - latency (min/avg/95%/max)
  - throughput
  - коды ответа

Команда для выполнения:

```bash
vegeta attack -targets=$(TARGETS_FILE_CREATE) -rate=$(RATE) -duration=$(DURATION) | tee /tmp/vegeta-test | vegeta report > $(REPORT_FILE)
```

- targets - задает цели атаки в данном случае запросы и тела запросов
- rate - количество запросов в секунду
- duration - продолжнительность атак

В отчетаз указывается информация в таком формате:
```bash
Requests      [total, rate, throughput]         600, 10.02, 10.01
Duration      [total, attack, wait]             59.912s, 59.9s, 12.264ms
Latencies     [min, mean, 50, 90, 95, 99, max]  4.267ms, 14.026ms, 11.514ms, 21.185ms, 27.709ms, 88.716ms, 191.838ms
Bytes In      [total, mean]                     502650, 837.75
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:600  
Error Set:

```

Так же можно указать гистограммное распределение для файла

```bash
Bucket           #    %       Histogram
[0s,     10ms]   260  43.33%  ################################
[10ms,   20ms]   276  46.00%  ##################################
[20ms,   50ms]   51   8.50%   ######
[50ms,   100ms]  8    1.33%   #
[100ms,  200ms]  5    0.83%   
[200ms,  500ms]  0    0.00%   
[500ms,  1s]     0    0.00%   
[1s,     +Inf]   0    0.00%   

```

Тестирование для каждого из файлов запускается через make файл

```makefile
perf_tests_get:
	clear
	@echo "Запуск установки сессий..."
	REDIS_ADDR=localhost:8010 go run docs/perf_test/main.go
	@echo "Запуск нагрузки..."
	$(MAKE) clean
	$(MAKE) make_perf_test_get
	$(MAKE) report
	$(MAKE) plot
	$(MAKE) histogram

make_perf_test_get:
	@echo "Запуск нагрузки на $(DURATION) с частотой $(RATE) запросов/сек..."
	vegeta attack -targets=$(TARGETS_FILE) -rate=$(RATE) -duration=$(DURATION) | tee /tmp/vegeta-test | vegeta report > $(REPORT_FILE)

report:
	@echo "Генерация текстового отчёта..."
	@cat $(REPORT_FILE)

plot:
	@echo "Генерация HTML-графика..."
	@cat /tmp/vegeta-test | vegeta plot > $(PLOT_FILE)
	@echo "Открой файл $(PLOT_FILE) в браузере."
	open $(PLOT_FILE)

histogram:
	@echo "Генерация гистограммы латентности..."
	@cat /tmp/vegeta-test | vegeta report -type=hist[0,10ms,20ms,50ms,100ms,200ms,500ms,1s] > $(HISTOGRAM_FILE)
	@cat $(HISTOGRAM_FILE)

```

- REDIS_ADDR=localhost:8010 go run docs/perf_test/main.go - Инициализация сессий, файлов запроса
- make_perf_test_get - запуск самого тестирования
- report - Генерация текстового отчёта
- plot - Генерация HTML-графика
- histogram - Генерация гистограммы латентности


## Процесс тестирования


### Cоздание профиля

### Получение профиля

В качестве получение профиля тестируется handler GET http://localhost:8080/profiles
На вход он получает ID текущего пользователя из cookie, на выход отдает 30 объектов типа профиль, которые
- не совпадают с текущим
- не имеют лайков с текущим

Текст первоначального запроса

```sql
SELECT 
    p.profile_id, 
    p.firstname, 
    p.lastname, 
    p.is_male,
    p.height,
    p.birthday, 
    p.description, 
    l.country, 
    l.city,
    l.district,
    s.path AS avatar,
    i.description AS interest,
    pr.preference_description,
    pr.preference_value 
FROM profiles p
LEFT JOIN locations l 
    ON p.location_id = l.location_id
LEFT JOIN "static" s 
    ON p.profile_id = s.profile_id
LEFT JOIN profile_interests pi 
    ON pi.profile_id = p.profile_id
LEFT JOIN interests i 
    ON pi.interest_id = i.interest_id
LEFT JOIN profile_preferences pp 
    ON pp.profile_id = p.profile_id
LEFT JOIN preferences pr 
    ON pp.preference_id = pr.preference_id
LEFT JOIN likes liked 
    ON liked.liked_profile_id = p.profile_id AND liked.profile_id = $1
WHERE p.profile_id != $1 AND liked.profile_id IS NULL AND p.profile_id > $2
LIMIT $3;
```

Проводим тестирование при 
RATE ?= 10
DURATION ?= 60s

```bash
Requests      [total, rate, throughput]         600, 10.02, 10.01
Duration      [total, attack, wait]             59.916s, 59.901s, 15.852ms
Latencies     [min, mean, 50, 90, 95, 99, max]  4.495ms, 19.302ms, 10.598ms, 23.309ms, 32.229ms, 228.396ms, 779.315ms
Bytes In      [total, mean]                     529759, 882.93
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:600  
Error Set:


Bucket           #    %       Histogram
[0s,     10ms]   275  45.83%  ##################################
[10ms,   20ms]   244  40.67%  ##############################
[20ms,   50ms]   62   10.33%  #######
[50ms,   100ms]  9    1.50%   #
[100ms,  200ms]  4    0.67%   
[200ms,  500ms]  3    0.50%   
[500ms,  1s]     3    0.50%   
[1s,     +Inf]   0    0.00%  
```

Значения являются приемлемыми:

- нет ни одной ошибки
- 95-й перцентиль: 32.2 мс 
- Средняя задержка (mean latency): 19.3 мс 
- Максимальная задержка: 779 мс — высокая, но редкая (менее 1%)

Однако это может создать проблемы при более высокой нагрузки приложения

Для ускорения работы приложения попробуем добавить индексы на запросы, включающие фильтрацию и джойны:

```sql
CREATE INDEX idx_profiles_profile_id ON profiles(profile_id);
CREATE INDEX idx_likes_by_liker ON likes(profile_id);
CREATE INDEX idx_likes_by_liked ON likes(liked_profile_id);

```

Тестирование с индексами на тех же данных:

```bash
Requests      [total, rate, throughput]         600, 10.02, 0.00
Duration      [total, attack, wait]             59.916s, 59.901s, 15.194ms
Latencies     [min, mean, 50, 90, 95, 99, max]  1.314ms, 16.522ms, 4.062ms, 9.053ms, 21.434ms, 449.168ms, 659.905ms
Bytes In      [total, mean]                     11400, 19.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           0.00%
Status Codes  [code:count]                      404:600  


Bucket           #    %       Histogram
[0s,     10ms]   544  90.67%  ####################################################################
[10ms,   20ms]   23   3.83%   ##
[20ms,   50ms]   11   1.83%   #
[50ms,   100ms]  4    0.67%   
[100ms,  200ms]  3    0.50%   
[200ms,  500ms]  11   1.83%   #
[500ms,  1s]     4    0.67%   
[1s,     +Inf]   0    0.00%   
```


Использование индексов заметно ускорило работу на получение основного профиля

Теперь попытаемся увеличить количество запросов в секунду с 10 до 100

```bash
Requests      [total, rate, throughput]         6000, 100.02, 0.00
Duration      [total, attack, wait]             59.992s, 59.989s, 2.567ms
Latencies     [min, mean, 50, 90, 95, 99, max]  710.5µs, 5.572ms, 2.293ms, 6.194ms, 13.23ms, 84.175ms, 312.17ms
Bytes In      [total, mean]                     114000, 19.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           0.00%
Status Codes  [code:count]                      404:6000  

Bucket           #     %       Histogram
[0s,     10ms]   5627  93.78%  ######################################################################
[10ms,   20ms]   154   2.57%   #
[20ms,   50ms]   110   1.83%   #
[50ms,   100ms]  59    0.98%   
[100ms,  200ms]  30    0.50%   
[200ms,  500ms]  20    0.33%   
[500ms,  1s]     0     0.00%   
[1s,     +Inf]   0     0.00%  
```

Несмотря на увеличение интенсивности, значения