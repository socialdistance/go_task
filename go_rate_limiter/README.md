Условия последней задачи
Дан код, делающий асинхронные вызовы функции: https://go.dev/play/p/DSku2XonSrT

- Нужно реализовать rate limiter, который ограничит максимально количество запросов: 10 в секунду
- Пусть лимит будет настраиваемым: 1, 10, 50 ... в секунду
- Выберите алгоритм для реализации. Например: Token Bucket
