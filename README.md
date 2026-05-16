# LinkTracker

**LinkTracker** – Telegram-бот, который отслеживает изменения на веб-страницах и оперативно информирует пользователя о них.

## Инструкция по запуску

### Получение токенов API
Для работы приложения, необходимо иметь доступ к API Telegram, Github и Stackoverflow.

Для этого нужно:

Получить токен тг-бота по [ссылке](https://help.botman.pro/article/23763) .

Получить токены github и stackOverFlow: [github](https://github.com/settings/tokens) и [stackoverflow](https://stackapps.com/applications)

Полученные токены следует скопировать в env файлы, согласно инструкции ниже
### Окружение и env файлы
Для запуска приложения необходимо три env файла сервисов в корне проекта: `.env.bot`, `.env.scrapper`, `.env.agent`

А также общий env файл **`.infra.env`** - который содержит настройки конфигурации кафки и ее топиков, и далее используется в [./scripts/create_topics.sh](./scripts/create_topics.sh)
для создания топиков путем инициализации в отдельном контейнере при старте сервисов. В этом env нет секретов, поэтому он уже присутствует в проекте

Их содержимое должно соответствовать структуре в [.env.bot.example](./.env.bot.example), [.env.scrapper.example](./.env.scrapper.example), [.env.agent.example](./.env.agent.example)
и [./infra.env](./.infra.env)

#### Переменные окружения 
##### env.bot
| Переменная        | Default  |
|-------------------|----------|
| TG_API_TOKEN      | token    |
| DATABASE_HOST     | bot-db   |
| DATABASE_PORT     | 5432     |
| POSTGRES_USER     | postgres |
| POSTGRES_PASSWORD | postgres |
| POSTGRES_DB       | bot-db   |
##### env.scrapper
| Переменная        | Default      |
|-------------------|--------------|
| GITHUB_API_TOKEN  | token        |
| STACK_OF_TOKEN    | token        |
| DATABASE_HOST     | scrapper-db  |
| DATABASE_PORT     | 5432         |
| POSTGRES_USER     | postgres     |
| POSTGRES_PASSWORD | postgres     |
| POSTGRES_DB       | link-tracker |
#### env.agent
| Переменная        | Default                                                       |
|-------------------|---------------------------------------------------------------|
|hf_token| hf_token -  токен API hugging face для работы AI суммаризации |

#### .infra.env
| Переменная        | Default                                                       |
|-------------------|---------------------------------------------------------------|
|KAFKA_BROKERS | "kafka1:9092,kafka2:9092,kafka3:9092" |
|KAFKA_TOPICS | "link.raw-updates link.processed-updates reports dead-letter-queue"|
|KAFKA_REPLICATION_FACTOR | 2 |
|KAFKA_PARTITIONS | 3|

### Файлы .conf
Это файлы конфигурации приложения, использующие библиотеку type-safe config.

Они содержат информацию о портах сервисов, конфигурации количества потоков-воркеров, о настройках запуска различных модулей, в том числе кафки, и т.д

Конфигурация скраппера: [scrapper.conf](./scrapper.conf)

Конфигурация бота: [bot.conf](./bot.conf)

Конфигурация ai агента: [agent.conf](./agent.conf)

В них уже описаны все нужные значения, но можно их изменить.

#### Тип работы с бд
**В частности, в файле scrapper.conf указан тип работы с базой данных, его можно менять между значениями *sql* и *query***

#### Транспорт
Выбор **транспорта** также задается в `scrapper.conf` и `bot.conf`:

`notification_type = "kafka"` - возможные значения: "kafka", "http",  "grpc"

Там же описана конфигурация kafka: 
``` 
brokers 
topics (названия) 
batch config
max_retries
replication_factor
required acks - "all" | "none"
```
#### Настройки кэширования
В `scrapper.conf` указаны настроки valkey cache. Есть возможность отключить client side кэширование : use_client_side_cache = bool. 
```
cache {
    addr = ["valkey-node-0:6379", "valkey-node-1:6379", "valkey-node-2:6379"]
    user = ""
    password = "bitnami"
    timeout=5s
    expire=10m
    use_client_side_cache = true
}
```

#### Конфигурация отказоустойчивости
И в `scrapper.conf`, и в `bot.conf` есть конфиг, отвечающий за настройку отказоустойчивого http-клиента и circuit breaker-а:
```
http_req_config {
    timeout = 120s
    delay = 5s
    max_delay = 30s
    delay_type = "constant"
    retries = 5
    retriable_codes = [408, 429, 500, 502, 503]
    circuit_breaker {
        min_required_calls = 10
        max_requests_half_open = 10
        closed_state_interval = 10s
        open_state_interval = 10s
        sliding_window_size = 15s
        failure_rate_threshold = 50
    }
} 
```
В `scrapper.conf` можно также настроить rate-limiting:
``` 
rate_limit {
    burst = 30
    limit = 10
    ip_expiration = 10m # параметр когда ip клиента устаревает и удаляется из сохраненного кэша
}
```
#### Настройки фильтрации сообщений в AI агенте

В конфигурации агента содержатся параметры **фильтрации** сообщений:

```
filter {
    stop-words = ["spam", "ads", "promo", "this is spam message"]
    excluded-authors = ["bot"]
    min-length = 20
    summarization {
        threshold =  50
    }
}
```

### Запуск приложения
Все приложение, включая Kafka-cluster и бд, поднимается в докере:

``docker compose up`` - запустит контейнер с базами данных на портах 5432 и 5433, **kafka-ui** на порту 8088.

scrapper будет доступен на порту 8082, bot - 8080

Внешний доступ к Kafka включён через порты 29091–29093 (dev-only)

#### Тесты
Для исполнения тестов:

```
docker compose up tests
```

Тесты также исполняются в докер контейнере.

Для корректной работы интеграционных тестов необходим доступ к docker-daemon, иначе testContainers не сможет работать

#### Локальный запуск
Для локального запуска сервиса можно исполнить следующие команды:

```
go mod tidy

go build -o bot cmd/bot/main.go

./bot 

go build -o scrapper cmd/scrapper/main.go

./scrapper 
```

Однако из-за библиотеки confluentic/confluent-kafka-go, нужно иметь Linux-окружение:(


#### Запуск тестов локально
```
go test -v ./...
```

Перед запуском тестов также стоит убедиться, что docker запущен, иначе testContainers не сможет работать 