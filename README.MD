# Transaction manager
This project is an API for a merch shop. [Here](https://github.com/avito-tech/tech-internship/tree/main/Tech%20Internships/Backend/Backend-trainee-assignment-winter-2025) you can see a TA.
## Configuration
The settings are stored in `secrets.env`.
The structure of this file should be as follows:
```
ENV=dev # options: dev, prod

TOKEN_TTL=1h
SALT=EXAMPLE

POSTGRES_USER=user
POSTGRES_PASSWORD=pass
POSTGRES_DB=db
POSTGRES_DSN=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@db:5432/${POSTGRES_DB}

REDIS_USER=default
REDIS_PASSWORD=eHAiOjE3Mzk3M
REDIS_DSN=redis://${REDIS_USER}:${REDIS_PASSWORD}@redis_container:6379

GOOSE_DRIVER=postgres
GOOSE_DBSTRING=${POSTGRES_DSN}?sslmode=disable
```
## Installation
1. Clone the repository:
```
git clone https://github.com/Lesion45/avito-internship-2025.git
cd avito-internship-2025
```
2. Run makefile:
```
make start
```
## Decisions
* Добавил в проект Redis, чтобы кешировать информацию о пользователях.
* Использовал утилиту Goose для выполнения миграций.
## Possible improvements 
* Не успел написать e2e-тесты, поэтому постарался максимально покрыть весь важный код unit-тестами.
* Вынести ошибки в controller в отдельное место.