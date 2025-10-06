## SkidIMG

**SkidIMG** это простое web-приложение для редактирования, хостинга и публиакации изображений, а так же группировки их в альбом.
 
Ключевая фича заключается в том, что сервис автоматически определяет `User-agent`. Eсли пользователь делится ссылкой на изображние, например в Telegeam или Discord, 
то будет отправленна **оптимизированная версия** для быстрой загрузки
Если открыть ту же ссылку в браузере, то отображается **оригинал изобращения**.

---

## Стек

- **Языки**: Go
- **Роутинг**: [chi](https://github.com/go-chi/chi)
- **Обработка изображений**: [bimg](https://github.com/h2non/bimg) (на осове libvips)
- **База данных**: PosgreSQL
- **CI/CD**: Github Actions + Docker Compose

---

## 🚀 Автосборка и deploy

При каждом новом пуше в `main` происходит автоматическая сборка и deploy приложения с помощью [Github Actions](.github/workflows/deploy.yml) + [Docker Compose](docker-compose.yml)

---

## 🔗 Демо-ссылка

62.182.192.227:8080/

http://img.downgrad.com/

---

## Требования для запуска

- Docker
- Docker Compose

---

## Сборка Проекта

В первую очередь нужно создать `.env` в корне проекта
```env
POSTGRES_PASSWORD=your_postgres_password
JWT_SECRET_KEY=your_jwt_secret
``` 
Запуск через Docker Compose 
```sh
docker-compose up --build
```
