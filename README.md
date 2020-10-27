Тестовое задание HTTP-мультиплексор:

- [x] приложение представляет собой http-сервер с одним хендлером
- [x] хендлер на вход получает POST-запрос со списком url в json-формате
- [x] сервер запрашивает данные по всем этим url
- [x] возвращает результат клиенту в json-формате
- [x] если в процессе обработки хотя бы одного из url получена ошибка, обработка всего списка прекращается и клиенту возвращается текстовая ошибка

Ограничения:

- [x] для реализации задачи следует использовать Go 1.13 или выше
- [X] использовать можно только компоненты стандартной библиотеки Go
- [x] сервер не принимает запрос если количество url  в нем больше 20
- [x] сервер не обслуживает больше чем 100 одновременных входящих http-запросов
- [x] для каждого входящего запроса должно быть не больше 4 одновременных исходящих
- [x] таймаут на запрос одного url - секунда
- [x] обработка запроса может быть отменена клиентом в любой момент, это должно повлечь за собой остановку всех операций связанных с этим запросом
- [x] сервис должен поддерживать 'graceful shutdown'
- [x] результат должен быть выложен на github