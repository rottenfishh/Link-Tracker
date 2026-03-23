# LinkTracker

**LinkTracker** – Telegram-бот, который отслеживает изменения на веб-страницах и оперативно информирует пользователя о них.

### Инструкция по запуску


Получить токен тг-бота по [ссылке](./https://help.botman.pro/article/23763) .


Получить токены github и stackOverFlow: [тык1](./https://github.com/settings/tokens) и [тык2](./https://stackapps.com/applications)

Создать .env файл в корне проекта и положить туда токены, согласно env.example.

Исполнить следующие команды:


```
go mod tidy

go build -o bot cmd/bot/main.go

./bot 

go build -o scrapper cmd/scrapper/main.go

./scrapper 
```
