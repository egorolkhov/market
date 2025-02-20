# Avito-2025

Cервис, который позволяет сотрудникам обмениваться монетками и приобретать на них мерч.

Каждый сотрудник должен иметь возможность видеть:
-Список купленных им мерчовых товаров
-Сгруппированную информацию о перемещении монеток в его кошельке, включая:
--Кто ему передавал монетки и в каком количестве
--Кому сотрудник передавал монетки и в каком количестве
-Количество монеток не может быть отрицательным, запрещено уходить в минус при операциях с монетками.
---

## ⚙️ **Запуск проекта**

### 1️⃣ **Склонируйте репозиторий**
Сначала клонируйте проект с GitHub:

```sh
git clone https://github.com/egorolkhov/market
cd market
```

### **Запуск с Docker**
Для развертывания сервиса используйте команду:

```sh
docker-compose up --build
```

Структура проекта
```
market/
│── cmd/                     # Основной entrypoint сервера
│   └── main.go
│── internal/                 # Основной код приложения
│   ├── app/                  # Логика приложения
│   ├── cache/                # Кеширование (LFU)
│   ├── config/               # Конфигурация
│   ├── logger/               # Логирование
│   ├── middleware/           # HTTP middleware
│   ├── models/               # Модели данных
│   ├── storage/              # Работа с БД
│   │   ├── transactionManager/  # Управление транзакциями
│   │   ├── db.go
│   │   ├── errors.go
│   │   ├── storage.go
│── migrations/               # SQL-миграции
│── tests/                    # Интеграционные тесты
│── docker-compose.yaml        # Контейнеры для БД
│── Dockerfile                 # Dockerfile для сервера
│── go.mod                     # Go-модули
```
