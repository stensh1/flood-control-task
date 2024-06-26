# Профильное задание Golang

# Выбор алгоритма
В поиске решения этого таска я пошел гуглить алгоритмы ограничения скорости обработки запросов.
Алгоритм Fixed Window не самый хороший и у него есть свои минусы, например: удвоение количества обработанных запросов на границе окна, но, как мне показалось, он быстрее и легче всего реализуемый, поэтому я выбрал его.
Хотя, алгоритм Sliding Window, наверное, работал бы для подобных задач лучше.

# Реализация
Для наглядности, я решил написать простой веб сервер, в котором, при переходе по определенному пути (/check) должен вызываться метод флуд контроля Check().

Здесь встал вопрос, как отличить запросы с одного устройства от других? - Либо куки, либо id пользователей
Прикручивать систему авторизации, хранения паролей и т.д. не хотелось, поэтому я решил сделать это на кукисах (вполне возможно, что метод может вызываться и пользователем, незалогиненым на сайте)
Куки для простоты генерируются рандомно с помощью rand.Int() из пакета math. Сделано это специально, чтобы без проблем отправить эти куки в качестве userID int64 в метод Check().

Для кэша запросов к серверу использовал Redis, как хороший переход от фронта к бэкенду, с высокой пропускной способностью.

# Принцип работы
При первом запросе к серверу (или при скинутых куках) сервер проверяет наличие кукисов в заголовках, если их нет, то высылает новые и регистрирует в Redis нового пользователя (при этом метод Check() не вызывается).
Далее при повторной попытке посетить /check с переданными кукис, сервер инициализирует первый запрос клиента, обновляет счетчик запросов и включает таймер обнуления счетчика, по данным из конфигурационного файла.
При последующих попытках клиента сделать запрос на сервер, сервер обновляет счетчик запросов и сравнивает их общее количество у клиента с установленным максимальным (также настроенным в конфигурационном файле).
Если запросов больше установленного - то выдает ошибку 429.

# Конфигурация
Большая часть данных записана в конфигурационном файле: cgf/config.yaml, как лимиты для сервера, так и настраиваемые хосты и порты для сервера и Redis, кроме пароля для бд, пароль задается в переменной окружения по пути: cfg/.env

# Послесловие
Закомитил все без гит игнора, в том чилсе и логи для наглядности. По хорошему, стоило покрыть тестами и запихнуть все в docker-compose, но, к сожалению, не успел.

# Обновление от 25.03.2024:
1. Сервер контейниризирован. Возможен запуск из docker compose.
2. Добавлен деструктор сервера, закрывает соединение с БД.
3. Удален исходный код приложения-клиента, предназначавшегося для тестирования сервера.
4. Изменен способ генерации кукис.

# Запуск
1. Загрузите файл docker-compose.yml
2. Из директории с файлом docker-compose вызовите терминал, выполните команду "docker compose up -d"
3. Из браузера перейдите по адресу: http://localhost:8080/check
4. Первое посещение адреса создаст куки для клиента браузера
5. Последующие посещения "http://localhost:8080/check" (обновления страницы) будут увеличивать счетчик запросов к серверу
6. Конфиг сервера настроен так, что хватает 10 быстрых нажатий обновления страницы, чтобы получить 429 ошибку - Too many requests
