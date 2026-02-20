## Board AI 
Телеграм-бот на Go, который моделирует заседание совета директоров для анализа бизнес-идей.  
Использует мультиагентную схему (Стратег-Оптимист, Финансист, Аудитор-Скептик, Аналитик Рынка, Модератор), локальный Ollama и PostgreSQL.

### Архитектура

- **Язык и рантайм**: Go
- **БД**: PostgreSQL (через Docker Compose)
- **LLM**: Ollama (совместимый с OpenAI `/v1/chat/completions`)
- **Интерфейс**: Telegram Bot API

Основная структура каталогов:

```plaintext
board-ai-bot/
├── cmd/bot/main.go                 // Инициализация зависимостей и запуск
├── internal/
│   ├── config/config.go            // Загрузка .env через os.LookupEnv
│   ├── models/analysis.go          // Структура Analysis
│   ├── repository/
│   │   ├── postgres.go             // Пул соединений sql.DB
│   │   └── analyses.go             // CRUD для сущности Analysis
│   ├── llm/
│   │   ├── client.go               // HTTP-клиент для Ollama
│   │   └── prompts.go              // Системные промпты агентов
│   ├── agents/
│   │   └── interface.go            // Определения ролей и агентов
│   ├── orchestrator/
│   │   └── orchestrator.go         // Параллельный запуск агентов
│   ├── bot/
│   │   ├── handlers.go             // Логика команд и state management
│   │   ├── keyboard.go             // Inline-кнопки
│   │   └── messages.go             // Рендер MarkdownV2
├── migrations/
│   └── 001_create_analyses.sql     // SQL-схема таблицы analyses
├── docker-compose.yml              // Сервис postgres:15-alpine
├── Makefile                        // Команды setup-models, docker-up, run и т.д.
└── go.mod
```

### Переменные окружения (`.env`)

Необходимо скопировать содержимое файла `.env.exmaple` и поменять значения под себя:
```bash
cp .env.example .env
```

```env
TELEGRAM_BOT_TOKEN= ___
DB_URL= ___
OLLAMA_BASE_URL= ___
MODEL_STRATEGIST=llama3:8b
MODEL_FINANCIER=gemma2:9b
MODEL_AUDITOR=mistral:7b
MODEL_ANALYST=qwen2.5:7b
MODEL_MODERATOR=llama3.1:8b
```

### Старт

```bash
make up
```
Что происходит:

- поднимается `postgres`;
- после того как Postgres становится `healthy`, контейнер `migrate` автоматически прогоняет миграции и завершается;
- поднимается `ollama` и слушает `http://localhost:11434`.

Остановка контейнеров:

### Стоп
```bash
make down
```

### Graceful Shutdown

Приложение перехватывает сигналы `SIGINT` и `SIGTERM`:

- плавно останавливает обработку обновлений Telegram;
- закрывает соединения с PostgreSQL;
- завершает выполнение без потери данных.
