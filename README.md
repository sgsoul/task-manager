# task-manager

## Телеграм бот

Лабораторная работа с курсов на python, дополнительно переписанная под golang. Бот принимает и выполняет задачи (отмечает выполненными), выводит задачи отсортированными по статусу. Работает на Clickhouse. 

Токен бота необходимо положить в переменную среды  ```$APP_TOKEN```.

Для запуска: 
```
docker pull yandex/clickhouse-server
docker network create bot-net
docker build -t tg-bot-go .

docker run --rm -d -p 8123:8123 -p 9000:9000 --name clickhouse --net=bot-net yandex/clickhouse-server
docker run --rm --name=tg-bot-go --net=bot-net tg-bot-go
```

## Веб сервис 

Лабораторная работа с курсов на python. Сервис также принимает и выполняет задачи (отмечает выполненными). Можно вывести задачи по статусу. Бекэнд написан на Flask. Работает на PostgreSQL.  

Для запуска: 
```
docker build -t back ./backend
docker build -t database ./database
docker build -t nginx ./frontend

docker volume create todo-db
docker network create todo-net

docker run --rm -d --name database --net=todo-net -v todo-db:/var/lib/postgresql/data -e POSTGRES_DB=docker_app_db -e POSTGRES_USER=docker_app -e POSTGRES_PASSWORD=docker_app database
docker run --rm -d --name backend --net=todo-net -e HOST=database -e PORT=5432 -e DB=docker_app_db -e DB_USERNAME=docker_app -e DB_PASSWORD=docker_app back
docker run --rm -d --name frontend --net=todo-net -p 80:80 -v $(pwd)/nginx/nginx.conf:/etc/nginx/nginx.conf:ro 7_nginx
```

Сервис будет доступен по адресу [localhost](http://localhost/)
