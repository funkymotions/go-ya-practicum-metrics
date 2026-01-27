# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

## Изыскания с pprof

Удалось сократить объем потребляемой памяти вдвое. В итерации 16 был добавлен аудит, через middleware аудита, который задваивал тело даже в выключенном состоянии.
В итоге реорганизовал аудит из цепочки вызовов, Notify() перенесён непосредственно в учаток логики, отвечающий за обновления метрик.
```
> go-ya-practicum-metrics > go tool pprof -top -diff_base=profiles/base.pb.gz profiles/result.pb.gz
```
*Output:*
```
File: main
Type: inuse_space
Time: Jan 14, 2026 at 6:35pm (MSK)
Duration: 20.01s, Total samples = 2049.37kB
Showing nodes accounting for -1535.37kB, 74.92% of 2049.37kB total
      flat  flat%   sum%        cum   cum%
     514kB 25.08% 25.08%      514kB 25.08%  bufio.NewReaderSize (inline)
    -513kB 25.03% 0.049%     -513kB 25.03%  io.ReadAll
 -512.22kB 24.99% 24.95%  -512.22kB 24.99%  runtime.malg
 -512.10kB 24.99% 49.93%  -512.10kB 24.99%  github.com/go-chi/chi.NewRouteContext
 -512.05kB 24.99% 74.92%  -512.05kB 24.99%  sync.runtime_SemacquireMutex
         0     0% 74.92%      514kB 25.08%  bufio.NewReader (inline)
         0     0% 74.92%  -512.05kB 24.99%  github.com/funkymotions/go-ya-practicum-metrics/internal/handler.(*metricHandler).SetMetricBulk
         0     0% 74.92% -1025.05kB 50.02%  github.com/funkymotions/go-ya-practicum-metrics/internal/middleware.(*AuditMiddleware).Audit.func1
         0     0% 74.92% -1025.05kB 50.02%  github.com/funkymotions/go-ya-practicum-metrics/internal/middleware.CompressHandler.func1
         0     0% 74.92%  -512.05kB 24.99%  github.com/funkymotions/go-ya-practicum-metrics/internal/repository.(*metricRepository).SetGauge
         0     0% 74.92%  -512.05kB 24.99%  github.com/funkymotions/go-ya-practicum-metrics/internal/repository.(*metricRepository).SetMetricBulk
         0     0% 74.92% -1025.05kB 50.02%  github.com/funkymotions/go-ya-practicum-metrics/internal/server.NewServer.HTTPLogMiddleware.func1.1
         0     0% 74.92%  -512.10kB 24.99%  github.com/funkymotions/go-ya-practicum-metrics/internal/server.NewServer.NewRouter.NewMux.func2
         0     0% 74.92%  -512.05kB 24.99%  github.com/funkymotions/go-ya-practicum-metrics/internal/service.(*metricService).SetMetricBulk
         0     0% 74.92% -1025.05kB 50.02%  github.com/go-chi/chi.(*ChainHandler).ServeHTTP
         0     0% 74.92% -1537.15kB 75.01%  github.com/go-chi/chi.(*Mux).ServeHTTP
         0     0% 74.92% -1025.05kB 50.02%  github.com/go-chi/chi.(*Mux).routeHTTP
         0     0% 74.92% -1023.15kB 49.92%  net/http.(*conn).serve
         0     0% 74.92% -1025.05kB 50.02%  net/http.HandlerFunc.ServeHTTP
         0     0% 74.92%      514kB 25.08%  net/http.newBufioReader
         0     0% 74.92% -1537.15kB 75.01%  net/http.serverHandler.ServeHTTP
         0     0% 74.92%  -512.22kB 24.99%  runtime.newproc.func1
         0     0% 74.92%  -512.22kB 24.99%  runtime.newproc1
         0     0% 74.92%  -512.22kB 24.99%  runtime.systemstack
         0     0% 74.92%  -512.05kB 24.99%  sync.(*Mutex).lockSlow
         0     0% 74.92%  -512.10kB 24.99%  sync.(*Pool).Get
         0     0% 74.92%  -512.05kB 24.99%  sync.(*RWMutex).Lock
```


## GRPC
```
protoc \
  -I . \
  --go_out=paths=source_relative:. \
  --go-grpc_out=paths=source_relative:. \
  internal/proto/metrics.proto
```
