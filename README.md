# Valorant Utility API

Backend REST API для утилиты Valorant — позволяет пользователям привязывать свои аккаунты Riot Games и просматривать персонализированные данные: ежедневный магазин, скины, ранг, историю матчей и баланс валюты.

**Живое демо:** [valorant.dmurygin.ru](https://valorant.dmurygin.ru) — фронтенд, работающий поверх этого API.

---

## Содержание

- [Стек](#стек)
- [Архитектура](#архитектура)
- [Функционал](#функционал)
- [API Reference](#api-reference)
- [Кэширование](#кэширование)
- [Запуск через Docker](#запуск-через-docker)
- [Локальная разработка](#локальная-разработка)
- [Переменные окружения](#переменные-окружения)

---

## Стек

| Слой | Технология |
|---|---|
| Язык | Go 1.26 |
| Web-фреймворк | Gin |
| База данных | PostgreSQL 17 |
| Кэш | Redis 7 |
| Аутентификация | JWT (HTTP-only cookies) + bcrypt |
| Документация | Swagger / OpenAPI |
| Логирование | Zap (uber-go) |
| Конфигурация | Viper + YAML + ENV |

---

## Архитектура

```
cmd/api/              — точка входа, настройка сервера
internal/
  api/v1/             — HTTP-хэндлеры и роутинг
    users/            — регистрация, авторизация, список аккаунтов
    riot/             — привязка Riot аккаунта (OAuth + прямой логин)
    valorant/         — магазин, кошелёк, ранг, матчи, скины
  domain/             — бизнес-логика и интерфейсы репозиториев
    user/
    valorant/
    match/
    asset/
  service/
    assets/           — сервис синхронизации ассетов (иконки, имена)
  storage/
    postgres/         — репозитории PostgreSQL
    redis/            — репозитории Redis (кэш)
  riot/               — клиенты Riot API
    auth/             — авторизация на серверах Riot
    store/            — витрина магазина
    mmr/              — ранг и RR
    match/            — история матчей
    wallet/           — баланс валюты
    assets/           — данные об агентах, скинах, картах
    loadout/          — инвентарь скинов игрока
    entitlements/     — инвентарь (entitlement token)
    nameservice/      — имена игроков
    content/          — контент (титулы)
  http/
    response/         — единая обёртка HTTP-ответов
  pkg/
    hash/             — bcrypt-хэширование паролей
    jwt/              — генерация и валидация JWT
  config/             — конфигурация (YAML + ENV)
  deps/               — dependency injection
  middleware/         — JWT-аутентификация
  logger/             — обёртка над Zap
migrations/           — SQL-миграции (golang-migrate)
```

Слои взаимодействуют строго сверху вниз: `handler → service → repository`. Riot API-клиенты вызываются только из хэндлеров — они не попадают в доменный слой.

---

## Функционал

### Пользователи
- Регистрация и вход по логину/паролю
- JWT access-токен (15 мин) + refresh-токен (7 дней) в HTTP-only cookies
- Ротация refresh-токена при обновлении
- Список привязанных Riot-аккаунтов с именем, тегом, рангом и RR

### Привязка Riot-аккаунта
Поддерживаются три способа:

1. **Прямой логин** (`POST /v1/riot/login`) — логин и пароль Riot. Автоматически обрабатывает MFA и капчу через дополнительные шаги.
2. **OAuth через браузер** (`GET /v1/riot/auth/url`) — генерирует ссылку, пользователь авторизуется в браузере и вставляет итоговый URL.
3. **Legacy callback** (`POST /v1/riot/callback`) — токены передаются напрямую клиентом.

После привязки сессия Riot (access token + entitlement token) хранится в Redis на 20 дней.

### Игровые данные
- **Ежедневный магазин** — 4 скина дня + аксессуары + бандлы, с именами и иконками
- **Кошелёк** — Valorant Points (VP), Radianite Points, Kingdom Credits
- **Ранг (MMR)** — текущий ранг (tier 0–27), RR, изменение RR за последний матч
- **История матчей** — последние 20 конкурентных матчей с фрагами, ассистами, смертями, агентами
- **Скины аккаунта** — полный инвентарь скинов с иконками

---

## API Reference

Интерактивная документация: `/swagger/index.html`

### Аутентификация

| Метод | Путь | Описание |
|---|---|---|
| `POST` | `/v1/users/register` | Регистрация |
| `POST` | `/v1/users/login` | Вход |
| `POST` | `/v1/users/logout` | Выход |
| `POST` | `/v1/users/refresh` | Обновление access-токена |
| `GET` | `/v1/users/me` | Данные текущего пользователя |
| `GET` | `/v1/users/accounts` | Список привязанных Riot-аккаунтов |

### Привязка Riot-аккаунта (требует JWT)

| Метод | Путь | Описание |
|---|---|---|
| `POST` | `/v1/riot/login` | Прямой логин по логину/паролю |
| `POST` | `/v1/riot/login/mfa` | Подтверждение 2FA кода |
| `POST` | `/v1/riot/login/captcha` | Отправка решения капчи |
| `GET` | `/v1/riot/auth/url` | Получить OAuth URL для браузера |
| `POST` | `/v1/riot/auth/submit-url` | Завершить OAuth через вставку URL |
| `POST` | `/v1/riot/callback` | Legacy callback с токенами |

### Игровые данные (требует JWT)

| Метод | Путь | Описание |
|---|---|---|
| `GET` | `/v1/valorant/store/:puuid` | Ежедневный магазин |
| `GET` | `/v1/valorant/wallet/:puuid` | Баланс валюты |
| `GET` | `/v1/valorant/mmr/:puuid` | Ранг и RR |
| `GET` | `/v1/valorant/matches/:puuid` | История матчей |
| `GET` | `/v1/valorant/account/:puuid` | Скины аккаунта |

Параметр `?force=true` у `/store`, `/matches` и `/account` инвалидирует кэш и принудительно подтягивает свежие данные с Riot API.

### Примеры запросов

```bash
# Регистрация
curl -c cookies.txt -X POST https://api.valorant.dmurygin.ru/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{"login": "myuser", "password": "secret"}'

# Прямой логин Riot-аккаунта
curl -b cookies.txt -X POST https://api.valorant.dmurygin.ru/v1/riot/login \
  -H "Content-Type: application/json" \
  -d '{"username": "RiotName", "password": "RiotPass"}'

# Получить ежедневный магазин
curl -b cookies.txt https://api.valorant.dmurygin.ru/v1/valorant/store/<puuid>

# Принудительно обновить историю матчей
curl -b cookies.txt "https://api.valorant.dmurygin.ru/v1/valorant/matches/<puuid>?force=true"
```

### Ответы

Все ответы имеют единую обёртку:

```json
{
  "success": true,
  "error": null,
  "...поля ответа..."
}
```

При ошибке:
```json
{
  "success": false,
  "error": {
    "message": "Login to your Riot account again"
  }
}
```

---

## Кэширование

| Данные | Хранилище | TTL | Ключ |
|---|---|---|---|
| Riot-сессия (токены) | Redis | 20 дней | `session:{puuid}` |
| Riot-cookies | Redis | 20 дней | `cookies:{puuid}` |
| Ежедневный магазин | Redis | До следующей ротации | `storefront:{puuid}` |
| Скины аккаунта | Redis | Бессрочно | `account:{puuid}` |
| Ранг / RR | Redis | 1 час | `account_rank:{puuid}` |
| Имена игроков | Redis | 7 дней | `player_name:{puuid}` |
| История матчей (кэш) | Redis | 15 мин | `matches:{puuid}` |
| История матчей (постоянно) | PostgreSQL | Постоянно | таблица `matches` |
| Данные ассетов (иконки, имена) | PostgreSQL | Постоянно | таблица `assets` |
| Refresh-токены | Redis | 7 дней | `refresh:{uuid}` |

---

## Запуск через Docker

### Требования

- Docker 24+
- Docker Compose v2

### Быстрый старт

```bash
# 1. Клонировать репозиторий
git clone <repo-url>
cd ValorantAPI

# 2. Создать .env из примера
cp .env.example .env
# Отредактировать .env — задать пароли и JWT_SECRET

# 3. Собрать и запустить
make docker-up

# Приложение доступно на http://localhost:8080
# Swagger: http://localhost:8080/swagger/index.html
# Health: http://localhost:8080/health
```

Миграции применяются автоматически при старте через сервис `migrate`.

### Команды

```bash
make docker-build   # Пересобрать образ
make docker-up      # Запустить все сервисы
make docker-down    # Остановить все сервисы
make docker-logs    # Логи приложения в реальном времени
```

### Состав docker-compose

| Сервис | Образ | Порт |
|---|---|---|
| `app` | Собирается из Dockerfile | `8080` |
| `db` | `postgres:17-alpine` | внутренний |
| `redis` | `redis:7-alpine` | внутренний |
| `migrate` | `migrate/migrate:v4.18.2` | — |

---

## Локальная разработка

### Требования

- Go 1.26+
- PostgreSQL 14+
- Redis 7+
- [golang-migrate](https://github.com/golang-migrate/migrate) (для миграций)
- [swag](https://github.com/swaggo/swag) (для генерации Swagger)

### Запуск

```bash
# Применить миграции
export DATABASE_URL="postgres://user:pass@localhost:5432/valorant_util?sslmode=disable"
make migration-up

# Запустить приложение
go run ./cmd/api
```

### Создать миграцию

```bash
make migration-create name=add_something
```

### Обновить Swagger-документацию

```bash
make swagger
```

---

## Переменные окружения

| Переменная | Описание | Обязательная |
|---|---|---|
| `POSTGRES_USER` | Пользователь PostgreSQL | Да |
| `POSTGRES_PASSWORD` | Пароль PostgreSQL | Да |
| `POSTGRES_DB` | Имя базы данных | Да |
| `POSTGRES_HOST` | Хост PostgreSQL | Нет (по умолчанию `localhost`) |
| `POSTGRES_PORT` | Порт PostgreSQL | Нет (по умолчанию `5432`) |
| `REDIS_HOST` | Хост Redis | Нет (по умолчанию `localhost`) |
| `REDIS_PORT` | Порт Redis | Нет (по умолчанию `6379`) |
| `REDIS_PASSWORD` | Пароль Redis | Нет |
| `JWT_SECRET` | Секрет для подписи JWT | Да |
| `LOG_LEVEL` | Уровень логов (`debug`, `info`, `warn`) | Нет (по умолчанию `info`) |
| `RIOT_ASSETS_API_BASE_URL` | Base URL для valorant-api.com | Нет (по умолчанию `https://valorant-api.com`) |

Все переменные можно задать через `.env` файл в корне проекта или через `internal/config/config.yaml`.
