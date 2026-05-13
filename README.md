# SchavelBot

Telegram-бот для студии «Щавелевый суп». Бот показывает приветственное меню, помогает записаться на сеанс, отправляет прайс и показывает свободные эскизы альбомом фотографий.

## Возможности

- приветственное сообщение по команде `/start`;
- кнопка записи на сеанс с контактом мастера;
- кнопка расчета стоимости с прайсом;
- кнопка просмотра свободных эскизов;
- отправка эскизов одним Telegram-альбомом;
- кнопка-ссылка на Telegram-канал.

## Переменные окружения

Создай файл `.env` в корне проекта:

```env
BOT_TOKEN=your_bot_token
CONTACT=@your_contact
TELEGRAM_CHANNEL=https://t.me/your_channel
```

Описание:

- `BOT_TOKEN` - токен бота из `@BotFather`;
- `CONTACT` - контакт, куда пользователь пишет для записи;
- `TELEGRAM_CHANNEL` - ссылка на Telegram-канал.

Не коммить `.env` в репозиторий.

## Локальный запуск

Установи зависимости:

```powershell
go mod download
```

Запусти бота:

```powershell
go run .
```

После запуска открой бота в Telegram и отправь:

```text
/start
```

## Docker

Собрать образ:

```powershell
docker build -t schavelbot .
```

Запустить контейнер:

```powershell
docker run --env-file .env schavelbot
```

В Docker-образ копируется папка `assets`, поэтому фотографии эскизов будут доступны внутри контейнера.

## Эскизы

Фотографии свободных эскизов лежат в папке:

```text
assets/
```

Сейчас бот ожидает такие файлы:

```text
assets/sketch1.jpg
assets/sketch2.jpg
assets/sketch3.jpg
```

Если нужно добавить больше фото, обнови список `sketches` в функции `sendSketches`. Telegram media group поддерживает от 2 до 10 файлов за одно сообщение.

## Структура

```text
.
├── assets/
├── Dockerfile
├── go.mod
├── go.sum
├── main.go
└── README.md
```

## Примечания

- Для локального запуска бот читает `.env` через `godotenv`.
- В Docker лучше передавать переменные через `--env-file`.
- Если Docker пишет, что не может подключиться к `dockerDesktopLinuxEngine`, запусти Docker Desktop и дождись полной инициализации.
