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

Для создания профиля существует handler POST http://213.219.214.83:8080/users
На вход подается учетная запись user с конфиденциальными данными и profile, который требуется создать

Текст запроса:
```sql
INSERT INTO profiles (firstname, lastname, is_male, birthday, height, description, location_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING profile_id;
```

Также совершаются отдельные запросы в таблицу интересов и профилей по необходимости

```sql
INSERT INTO interests (description)
VALUES ($1)
RETURNING interest_id
```


```sql
INSERT INTO profile_interests (profile_id, interest_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING
```

Аналогично с preferences

Тестирование:

```bash
Генерация текстового отчёта...
Requests      [total, rate, throughput]         100000, 20.00, 19.45
Duration      [total, attack, wait]             1h23m0s, 1h23m0s, 78.186ms
Latencies     [min, mean, 50, 90, 95, 99, max]  1.898ms, 695.698ms, 59.143ms, 229.948ms, 1.122s, 24.497s, 31.054s
Bytes In      [total, mean]                     2640606, 26.41
Bytes Out     [total, mean]                     63736049, 637.36
Success       [ratio]                           97.27%
Status Codes  [code:count]                      0:2362  201:97273  500:365  
Error Set:
500 Internal Server Error
Post "http://localhost:8080/users": EOF
Post "http://localhost:8080/users": context deadline exceeded (Client.Timeout exceeded while awaiting headers)

Bucket           #      %       Histogram
[0s,     10ms]   1      0.00%   
[10ms,   20ms]   0      0.00%   
[20ms,   50ms]   29060  29.06%  #####################
[50ms,   100ms]  52009  52.01%  #######################################
[100ms,  200ms]  8024   8.02%   ######
[200ms,  500ms]  4366   4.37%   ###
[500ms,  1s]     1463   1.46%   #
[1s,     +Inf]   5077   5.08%   ###
```

Видно, что запросы идут достаточно быстро, так как запросы итак были изначально разделены, каждый из них по отдельности максимально отпимизирован


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
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:600  


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
Success       [ratio]                           100%
Status Codes  [code:count]                      200:6000  

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

Несмотря на увеличение интенсивности, значения все еще остались в приемлемом диапазоне. Проверим это на виртуальной машине
Для этого в файле инициализации make заменим localhost на адрес машины

```bash
Requests      [total, rate, throughput]         6000, 100.02, 42.81
Duration      [total, attack, wait]             1m28s, 59.99s, 28.161s
Latencies     [min, mean, 50, 90, 95, 99, max]  388.243ms, 11.637s, 6.12s, 30.001s, 30.001s, 30.005s, 30.023s
Bytes In      [total, mean]                     3176001, 529.33
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  


Генерация гистограммы латентности...
Bucket           #     %       Histogram
[0s,     10ms]   0     0.00%   
[10ms,   20ms]   0     0.00%   
[20ms,   50ms]   0     0.00%   
[50ms,   100ms]  0     0.00%   
[100ms,  200ms]  0     0.00%   
[200ms,  500ms]  175   2.92%   ##
[500ms,  1s]     2437  40.62%  ##############################
[1s,     +Inf]   3388  56.47%  ##########################################
```

При запуске на виртмашине обнаружилось, что запросы все еще остаются неэффективными
Самый длинный запрос на получение шел 30 секунд. Это слишком долго, поэтому необходимо провести дополнительную оптимизацию

- Многочисленные join и фильтрации
Сейчас JOIN'ы выполняются до фильтрации (WHERE), что увеличивает количество строк, проходящих через джойны
Оптимизация: сначала получить profile_id нужных профилей, затем сделать JOIN'ы по ним, чтобы JOIN'ы работали только по 30 записям (LIMIT $3), а не по всей таблице profiles:

```sql
WITH filtered_profiles AS (
    SELECT p.profile_id
    FROM profiles p
    LEFT JOIN likes liked 
        ON liked.liked_profile_id = p.profile_id AND liked.profile_id = $1
    WHERE p.profile_id != $1 
      AND liked.profile_id IS NULL 
      AND p.profile_id > $2
    ORDER BY p.profile_id
    LIMIT $3
)
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
FROM filtered_profiles fp
JOIN profiles p ON p.profile_id = fp.profile_id
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
    ON pp.preference_id = pr.preference_id;

```


- Добавить дополнительные индексы по join

Для устранения Seq Scan (полных проходов по таблице) PostgreSQL, были добавлены недостающие индексы, участвующие в JOIN'ах по profile_id:

```sql
CREATE INDEX IF NOT EXISTS idx_static_profile_id ON "static"(profile_id);
CREATE INDEX IF NOT EXISTS idx_profile_interests_profile_id ON profile_interests(profile_id);
CREATE INDEX IF NOT EXISTS idx_profile_preferences_profile_id ON profile_preferences(profile_id);
```
Эти индексы позволяют PostgreSQL использовать Index Scan, что значительно снижает время выполнения сложных объединений.


- Упорядочивание
Для более эффективной постраничной выборки (keyset pagination), используется сортировка по profile_id:

```sql
ORDER BY p.profile_id
```
В сочетании с фильтрацией p.profile_id > $2 это позволяет избежать пагинации через OFFSET, что критично при больших объемах данных.

- Оптимизация внутри go кода

Ранее: профили загружались из базы в неоптимальном порядке, хранились в map, затем сортировались вручную.

Теперь:

- Профили извлекаются уже отсортированными SQL-запросом.

- Используется keyset pagination.

- Обновление Redis-ключа вынесено из критического пути — выполняется асинхронно.


Тестирование кода после исправления

```sql
Генерация текстового отчёта...
Requests      [total, rate, throughput]         6000, 100.02, 98.81
Duration      [total, attack, wait]             1m1s, 59.99s, 730.255ms
Latencies     [min, mean, 50, 90, 95, 99, max]  493.767ms, 804.216ms, 766.496ms, 1.14s, 1.267s, 1.456s, 2s
Bytes In      [total, mean]                     97694103, 16282.35
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:6000  

Генерация гистограммы латентности...
Bucket           #     %       Histogram
[0s,     10ms]   0     0.00%   
[10ms,   20ms]   0     0.00%   
[20ms,   50ms]   0     0.00%   
[50ms,   100ms]  0     0.00%   
[100ms,  200ms]  0     0.00%   
[200ms,  500ms]  75    1.25%   
[500ms,  1s]     5171  86.18%  ################################################################
[1s,     +Inf]   754   12.57%  #########
```
Путем оптимизаций улалось снизить задержку и увеличить количество быстродействующих запросов. На всякий случай, убедимся и сделаем EXPLAM ANALYZE

```js
[
  {
    "QUERY PLAN": "Sort  (cost=334.17..334.61 rows=177 width=288) (actual time=2.774..2.877 rows=623 loops=1)"
  },
  {
    "QUERY PLAN": "  Sort Key: p.profile_id"
  },
  {
    "QUERY PLAN": "  Sort Method: quicksort  Memory: 182kB"
  },
  {
    "QUERY PLAN": "  ->  Hash Left Join  (cost=256.18..327.56 rows=177 width=288) (actual time=0.603..2.149 rows=623 loops=1)"
  },
  {
    "QUERY PLAN": "        Hash Cond: (pi.interest_id = i.interest_id)"
  },
  {
    "QUERY PLAN": "        ->  Nested Loop Left Join  (cost=252.70..323.60 rows=177 width=278) (actual time=0.508..1.670 rows=623 loops=1)"
  },
  {
    "QUERY PLAN": "              ->  Hash Left Join  (cost=252.28..290.55 rows=60 width=270) (actual time=0.484..0.855 rows=107 loops=1)"
  },
  {
    "QUERY PLAN": "                    Hash Cond: (pp.preference_id = pr.preference_id)"
  },
  {
    "QUERY PLAN": "                    ->  Nested Loop Left Join  (cost=247.71..285.82 rows=60 width=233) (actual time=0.352..0.641 rows=107 loops=1)"
  },
  {
    "QUERY PLAN": "                          ->  Nested Loop Left Join  (cost=247.29..270.64 rows=30 width=225) (actual time=0.332..0.437 rows=30 loops=1)"
  },
  {
    "QUERY PLAN": "                                ->  Hash Right Join  (cost=247.13..269.24 rows=30 width=137) (actual time=0.295..0.338 rows=30 loops=1)"
  },
  {
    "QUERY PLAN": "                                      Hash Cond: (s.profile_id = p.profile_id)"
  },
  {
    "QUERY PLAN": "                                      ->  Seq Scan on static s  (cost=0.00..18.80 rows=880 width=40) (actual time=0.023..0.027 rows=10 loops=1)"
  },
  {
    "QUERY PLAN": "                                      ->  Hash  (cost=246.76..246.76 rows=30 width=105) (actual time=0.249..0.252 rows=30 loops=1)"
  },
  {
    "QUERY PLAN": "                                            Buckets: 1024  Batches: 1  Memory Usage: 13kB"
  },
  {
    "QUERY PLAN": "                                            ->  Nested Loop  (cost=0.74..246.76 rows=30 width=105) (actual time=0.094..0.222 rows=30 loops=1)"
  },
  {
    "QUERY PLAN": "                                                  ->  Limit  (cost=0.44..1.46 rows=30 width=8) (actual time=0.069..0.091 rows=30 loops=1)"
  },
  {
    "QUERY PLAN": "                                                        ->  Merge Anti Join  (cost=0.44..3378.85 rows=100003 width=8) (actual time=0.068..0.084 rows=30 loops=1)"
  },
  {
    "QUERY PLAN": "                                                              Merge Cond: (p_1.profile_id = liked.liked_profile_id)"
  },
  {
    "QUERY PLAN": "                                                              ->  Index Only Scan using profiles_pkey on profiles p_1  (cost=0.29..3104.49 rows=100009 width=8) (actual time=0.056..0.065 rows=30 loops=1)"
  },
  {
    "QUERY PLAN": "                                                                    Index Cond: (profile_id > 0)"
  },
  {
    "QUERY PLAN": "                                                                    Filter: (profile_id <> 1)"
  },
  {
    "QUERY PLAN": "                                                                    Rows Removed by Filter: 1"
  },
  {
    "QUERY PLAN": "                                                                    Heap Fetches: 0"
  },
  {
    "QUERY PLAN": "                                                              ->  Index Only Scan using likes_profile_id_liked_profile_id_key on likes liked  (cost=0.15..24.26 rows=6 width=8) (actual time=0.007..0.007 rows=0 loops=1)"
  },
  {
    "QUERY PLAN": "                                                                    Index Cond: (profile_id = 1)"
  },
  {
    "QUERY PLAN": "                                                                    Heap Fetches: 0"
  },
  {
    "QUERY PLAN": "                                                  ->  Index Scan using profiles_pkey on profiles p  (cost=0.29..8.18 rows=1 width=105) (actual time=0.003..0.003 rows=1 loops=30)"
  },
  {
    "QUERY PLAN": "                                                        Index Cond: (profile_id = p_1.profile_id)"
  },
  {
    "QUERY PLAN": "                                ->  Memoize  (cost=0.16..0.18 rows=1 width=104) (actual time=0.002..0.002 rows=0 loops=30)"
  },
  {
    "QUERY PLAN": "                                      Cache Key: p.location_id"
  },
  {
    "QUERY PLAN": "                                      Cache Mode: logical"
  },
  {
    "QUERY PLAN": "                                      Hits: 24  Misses: 6  Evictions: 0  Overflows: 0  Memory Usage: 1kB"
  },
  {
    "QUERY PLAN": "                                      ->  Index Scan using locations_pkey on locations l  (cost=0.15..0.17 rows=1 width=104) (actual time=0.007..0.007 rows=1 loops=6)"
  },
  {
    "QUERY PLAN": "                                            Index Cond: (location_id = p.location_id)"
  },
  {
    "QUERY PLAN": "                          ->  Index Only Scan using profile_preferences_pkey on profile_preferences pp  (cost=0.42..0.49 rows=2 width=16) (actual time=0.004..0.005 rows=4 loops=30)"
  },
  {
    "QUERY PLAN": "                                Index Cond: (profile_id = p.profile_id)"
  },
  {
    "QUERY PLAN": "                                Heap Fetches: 0"
  },
  {
    "QUERY PLAN": "                    ->  Hash  (cost=3.14..3.14 rows=114 width=53) (actual time=0.107..0.108 rows=114 loops=1)"
  },
  {
    "QUERY PLAN": "                          Buckets: 1024  Batches: 1  Memory Usage: 19kB"
  },
  {
    "QUERY PLAN": "                          ->  Seq Scan on preferences pr  (cost=0.00..3.14 rows=114 width=53) (actual time=0.014..0.060 rows=114 loops=1)"
  },
  {
    "QUERY PLAN": "              ->  Index Only Scan using profile_interests_pkey on profile_interests pi  (cost=0.42..0.52 rows=3 width=16) (actual time=0.004..0.005 rows=6 loops=107)"
  },
  {
    "QUERY PLAN": "                    Index Cond: (profile_id = p.profile_id)"
  },
  {
    "QUERY PLAN": "                    Heap Fetches: 0"
  },
  {
    "QUERY PLAN": "        ->  Hash  (cost=2.10..2.10 rows=110 width=26) (actual time=0.077..0.078 rows=110 loops=1)"
  },
  {
    "QUERY PLAN": "              Buckets: 1024  Batches: 1  Memory Usage: 15kB"
  },
  {
    "QUERY PLAN": "              ->  Seq Scan on interests i  (cost=0.00..2.10 rows=110 width=26) (actual time=0.009..0.039 rows=110 loops=1)"
  },
  {
    "QUERY PLAN": "Planning Time: 4.785 ms"
  },
  {
    "QUERY PLAN": "Execution Time: 3.626 ms"
  }
]
```

Время выполнения: 3.626 ms

Ключевые моменты:

- Используются Index Only Scan для таблиц profiles, likes, profile_preferences, profile_interests.

- Активно задействован Memoize для кэширования location_id → locations, что ускоряет повторные запросы.

- Большинство соединений — Hash Left Join и Nested Loop Join, что допустимо при небольшом объеме данных (или хороших индексах).

- Последовательное сканирование (Seq Scan) осталось только на малых таблицах (static, preferences, interests), что не критично при их размере.

Вывод: план запроса теперь оптимален и использует индексы, результат достигается за миллисекунды.

