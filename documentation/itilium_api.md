# Itilium API (использование в проекте)

Документация описывает интеграцию с ITILIUM, реализованную в `src/api/itilium_api.py`, и реальные поля/форматы,
которые используются в коде.

## Базовый URL и доступ

- Базовый URL берется из `ITILIUM_URL` (см. `src/config/configuration.py` и `.env.example`).
- Аутентификация: Basic Auth (`ITILIUM_LOGIN` / `ITILIUM_PASSWORD`).
- HTTP клиент: общий httpx `AsyncClient` с `verify=False` (SSL проверка отключена).
- Запросы выполняются через `log_and_request` в `src/utils/http_client.py`.

## Общий формат запросов

- В большинстве методов используется `data=...` (form-encoded), а не `json=...`.
- Заголовки клиента по умолчанию: `Content-Type: application/json`, хотя фактически отправляется `data`.
- Ответы обрабатываются по HTTP статусам (успех: 200/201/204).
- В примерах ниже используется placeholder `{{ITILIUM_URL}}`.

## Эндпоинты ITILIUM

Ниже перечислены все endpoints, которые реально вызываются в проекте.

### 1) Найти сотрудника по Telegram

- **Method:** `POST`
- **Path:** `find_employee`
- **Request (data):**
  - `telegram`: `int` (по умолчанию)
  - (опционально) другой атрибут вместо `telegram`, если передан `attribute_code`
- **Response (JSON, используется в коде):**
  - `UUID`: идентификатор сотрудника
  - `servicecalls`: список номеров/ID заявок пользователя
  - `canCreateMarketingRequests`: `bool` (право на маркетинг)

Использование: `ItiliumBaseApi.get_employee_data_by_identifier`.

#### JSON схема (используемые поля)
```json
{
  "UUID": "string",
  "servicecalls": ["string"],
  "canCreateMarketingRequests": true
}
```

#### Пример запроса
```http
POST {{ITILIUM_URL}}/find_employee
Content-Type: application/x-www-form-urlencoded

telegram=123456789
```

#### Пример ответа
```json
{
  "UUID": "d2b4c3e0-1234-4f88-9c1f-58a1b1f9d2aa",
  "servicecalls": ["SC-000123", "SC-000124"],
  "canCreateMarketingRequests": true
}
```

---

### 2) Создать обычную заявку (Service Call)

- **Method:** `POST`
- **Path:** `create_sc`
- **Request (data):**
  - `client`: `UUID` пользователя
  - `shorDescription`: краткая тема (обратите внимание на написание в коде)
  - `Description`: полное описание
  - `files`: JSON-строка списка файлов (опционально)
    - пример элемента: `{ "path": "photos/file_19.jpg", "filename": "file_19.jpg" }`
- **Response:** важен только статус (`200/201/204` — успех)

Использование: `ItiliumBaseApi.create_new_sc`.

#### JSON схема (используемые поля)
```json
{
  "client": "string",
  "shorDescription": "string",
  "Description": "string",
  "files": "[{\"path\":\"string\",\"filename\":\"string\"}]"
}
```

#### Пример запроса
```http
POST {{ITILIUM_URL}}/create_sc
Content-Type: application/x-www-form-urlencoded

client=d2b4c3e0-1234-4f88-9c1f-58a1b1f9d2aa&
shorDescription=Проблема%20со%20входом&
Description=Не%20могу%20войти%20в%20систему&
files=[{"path":"photos/file_19.jpg","filename":"file_19.jpg"}]
```

#### Пример ответа
```json
{
  "status": "ok",
  "number": "SC-000125"
}
```

---

### 3) Согласовать/отклонить голосование по заявке

- **Method:** `POST`
- **Path:** `vote_change?telegram={telegram_user_id}&vote_number={vote_number}&state={state}`
- **Query params:**
  - `telegram`: Telegram ID
  - `vote_number`: номер голосования
  - `state`: `accept` или `reject`

Использование: `accept_callback_handler` / `reject_callback_handler`.

#### Пример запроса (accept)
```http
POST {{ITILIUM_URL}}/vote_change?telegram=123456789&vote_number=98765&state=accept
```

#### Пример ответа
```json
{
  "status": "ok"
}
```

---

### 4) Найти заявку по номеру

- **Method:** `POST`
- **Path:** `find_sc?telegram={telegram_user_id}&sc_number={sc_number}`
- **Response (JSON, используется в коде):**
  - `number`
  - `shortDescription`
  - `state`
  - `responsibleTeamTitle`
  - `deadlineDate`
  - `description`
  - `new_state`: список возможных статусов
  - `change_responsible`: `bool` (доступна ли смена ответственного)

Использование: `ItiliumBaseApi.find_sc_by_id`.

#### JSON схема (используемые поля)
```json
{
  "number": "string",
  "shortDescription": "string",
  "state": "string",
  "responsibleTeamTitle": "string",
  "deadlineDate": "string",
  "description": "string",
  "new_state": ["string"],
  "change_responsible": true
}
```

#### Пример запроса
```http
POST {{ITILIUM_URL}}/find_sc?telegram=123456789&sc_number=SC-000123
```

#### Пример ответа
```json
{
  "number": "SC-000123",
  "shortDescription": "Не работает доступ",
  "state": "registered",
  "responsibleTeamTitle": "Отдел ИТ",
  "deadlineDate": "2026-01-30",
  "description": "Описание проблемы ...",
  "new_state": ["in_work", "closed"],
  "change_responsible": false
}
```

---

### 5) Добавить комментарий к заявке

- **Method:** `POST`
- **Path:** `add_comment?telegram={telegram_user_id}&source={source}&source_type=servicecall&comment_text={comment_text}`
- **Query params:**
  - `telegram`: Telegram ID
  - `source`: номер заявки
  - `source_type`: фиксировано `servicecall`
  - `comment_text`: текст комментария
  - `files`: опционально, список файлов через `;` (добавляется в коде)

Использование: `ItiliumBaseApi.add_comment_to_sc`.

#### Пример запроса
```http
POST {{ITILIUM_URL}}/add_comment?telegram=123456789&source=SC-000123&source_type=servicecall&comment_text=Проверьте%20еще%20раз
```

#### Пример запроса с файлами
```http
POST {{ITILIUM_URL}}/add_comment?telegram=123456789&source=SC-000123&source_type=servicecall&comment_text=Комментарий&files=photos/file_19.jpg;documents/file_21.pdf
```

#### Пример ответа
```json
{
  "status": "ok"
}
```

---

### 6) Оценить заявку (закрытие)

- **Method:** `POST`
- **Path:** `confirm_sc?telegram={telegram_user_id}&incident={incident}&mark={mark}`
- **Query params:**
  - `mark`: от 0 до 5
  - `comment_text`: опционально

Использование: `ItiliumBaseApi.confirm_sc`.

#### Пример запроса
```http
POST {{ITILIUM_URL}}/confirm_sc?telegram=123456789&incident=SC-000123&mark=5&comment_text=Спасибо
```

#### Пример ответа
```json
{
  "status": "ok"
}
```

---

### 7) Список заявок, где пользователь ответственный

- **Method:** `POST`
- **Path:** `list_sc_responsible?telegram={telegram_user_id}`
- **Response:** список номеров заявок (далее запрашиваются детально через `find_sc`)

Использование: `ItiliumBaseApi.scs_responsibility_tasks`.

#### Пример запроса
```http
POST {{ITILIUM_URL}}/list_sc_responsible?telegram=123456789
```

#### Пример ответа
```json
["SC-000101", "SC-000102", "SC-000103"]
```

---

### 8) Сменить статус заявки

- **Method:** `POST`
- **Path:** `change_state_sc?telegram={telegram_user_id}&inc_number={inc_number}&new_state={new_state}`
- **Query params:** `telegram`, `inc_number`, `new_state`

Использование: `ItiliumBaseApi.change_sc_state`.

#### Пример запроса
```http
POST {{ITILIUM_URL}}/change_state_sc?telegram=123456789&inc_number=SC-000123&new_state=in_work
```

#### Пример ответа
```json
{
  "status": "ok"
}
```

---

### 9) Сменить статус с комментарием и датой

- **Method:** `POST`
- **Path:** `change_state_sc?telegram={telegram_user_id}&inc_number={inc_number}&new_state={new_state}&date_inc={date_inc}&comment={comment}`
- **Query params:** `telegram`, `inc_number`, `new_state`, `date_inc`, `comment`

Использование: `ItiliumBaseApi.change_sc_state_with_comment`.

#### Пример запроса
```http
POST {{ITILIUM_URL}}/change_state_sc?telegram=123456789&inc_number=SC-000123&new_state=postponed&date_inc=2026-02-01&comment=Нужны%20доп%20данные
```

#### Пример ответа
```json
{
  "status": "ok"
}
```

---

### 10) Получить список подразделений/ответственных по заявке

- **Method:** `POST`
- **Path:** `responsibles_sc?telegram={telegram_user_id}&sc_number={sc_number}`
- **Response (JSON, используется в коде):**
  - список подразделений с полями:
    - `responsibleTeamId`
    - `responsibleTeamTitle`
    - `responsibles`: список сотрудников
      - `responsibleEmployeeId`
      - `responsibleEmployeeTitle`

Использование: `ItiliumBaseApi.get_responsibles`.

#### JSON схема (используемые поля)
```json
[
  {
    "responsibleTeamId": "string",
    "responsibleTeamTitle": "string",
    "responsibles": [
      {
        "responsibleEmployeeId": "string",
        "responsibleEmployeeTitle": "string"
      }
    ]
  }
]
```

#### Пример запроса
```http
POST {{ITILIUM_URL}}/responsibles_sc?telegram=123456789&sc_number=SC-000123
```

#### Пример ответа
```json
[
  {
    "responsibleTeamId": "TEAM-001",
    "responsibleTeamTitle": "Сервис-деск",
    "responsibles": [
      {
        "responsibleEmployeeId": "EMP-1001",
        "responsibleEmployeeTitle": "Иванов И.И."
      }
    ]
  }
]
```

---

### 11) Назначить нового ответственного

- **Method:** `POST`
- **Path:** `change_responsible_sc?telegram={telegram_user_id}&inc_number={inc_number}&responsibleEmployeeId={responsible_employee_id}`
- **Query params:** `telegram`, `inc_number`, `responsibleEmployeeId`

Использование: `ItiliumBaseApi.change_responsible`.

#### Пример запроса
```http
POST {{ITILIUM_URL}}/change_responsible_sc?telegram=123456789&inc_number=SC-000123&responsibleEmployeeId=EMP-1001
```

#### Пример ответа
```json
{
  "status": "ok"
}
```

---

## Маркетинговые заявки

### 12) Список сервисов маркетинга

- **Method:** `GET`
- **Path:** `/listServicesMarketing?telegram={telegram_id}`
- **Response (JSON, используется в коде):**
  - список услуг, где важны поля:
    - `КомпонентаУслуги` (отображаемое имя)
    - `НомерФормы` (определяет тип формы: 1/2/3)

Использование: `ItiliumBaseApi.get_marketing_services`.

#### JSON схема (используемые поля)
```json
[
  {
    "КомпонентаУслуги": "string",
    "НомерФормы": 1
  }
]
```

#### Пример запроса
```http
GET {{ITILIUM_URL}}/listServicesMarketing?telegram=123456789
```

#### Пример ответа
```json
[
  {
    "КомпонентаУслуги": "Дизайн",
    "НомерФормы": 1
  },
  {
    "КомпонентаУслуги": "Мероприятие",
    "НомерФормы": 2
  }
]
```

---

### 13) Список подразделений маркетинга

- **Method:** `GET`
- **Path:** `/listSubdivisionMarketing?telegram={telegram_id}`
- **Response:** список строк (названия подразделений)

Использование: `ItiliumBaseApi.get_marketing_subdivisions`.

#### Пример запроса
```http
GET {{ITILIUM_URL}}/listSubdivisionMarketing?telegram=123456789
```

#### Пример ответа
```json
[
  "Отдел маркетинга",
  "PR",
  "Digital"
]
```

---

### 14) Создать маркетинговую заявку

- **Method:** `POST`
- **Path:** `create_sc_Marketing`
- **Request (data):**
  - `telegram`: Telegram ID
  - `Services`: название услуги (например, `Дизайн`, `Мероприятие`, `Реклама`, `SMM`, `Акция`, `Иное`)
  - `Subdivision`: название подразделения
  - `ExecutionDate`: дата исполнения (в коде отправляется как `YYYY.MM.DD`)
  - `files`: JSON-строка списка файлов (опционально)
    - формат аналогичен обычной заявке

**Дополнительные поля формы (в зависимости от сервиса):**

**Сервис: `Дизайн` (форма 1)**
- `LayoutName`
- `Size`
- `ForWhat`
- `RequiredText`
- `LayoutFormats`

**Сервис: `Мероприятие` (форма 2)**
- `ThemeEvent`
- `Description`
- `Budget`
- если в `form_data` присутствует `free_text`, он заменяет `Description`

**Сервис: `Реклама` / `SMM` / `Акция` / `Иное` (форма 3)**
- `Description` (берется из `free_text`)

Использование: `ItiliumBaseApi.create_marketing_request`.

#### JSON схема (используемые поля)
```json
{
  "telegram": 123456789,
  "Services": "string",
  "Subdivision": "string",
  "ExecutionDate": "YYYY.MM.DD",
  "files": "[{\"path\":\"string\",\"filename\":\"string\"}]",
  "LayoutName": "string",
  "Size": "string",
  "ForWhat": "string",
  "RequiredText": "string",
  "LayoutFormats": "string",
  "ThemeEvent": "string",
  "Description": "string",
  "Budget": "string"
}
```

#### Пример запроса (Дизайн)
```http
POST {{ITILIUM_URL}}/create_sc_Marketing
Content-Type: application/x-www-form-urlencoded

telegram=123456789&
Services=Дизайн&
Subdivision=Отдел%20маркетинга&
ExecutionDate=2026.02.10&
LayoutName=Листовка&
Size=A4&
ForWhat=Печать&
RequiredText=Текст%20для%20листовки&
LayoutFormats=PDF,AI&
files=[{"path":"photos/file_19.jpg","filename":"file_19.jpg"}]
```

#### Пример запроса (Мероприятие)
```http
POST {{ITILIUM_URL}}/create_sc_Marketing
Content-Type: application/x-www-form-urlencoded

telegram=123456789&
Services=Мероприятие&
Subdivision=PR&
ExecutionDate=2026.02.10&
ThemeEvent=День%20компании&
Description=Описание%20события&
Budget=100000
```

#### Пример запроса (Реклама/SMM/Акция/Иное)
```http
POST {{ITILIUM_URL}}/create_sc_Marketing
Content-Type: application/x-www-form-urlencoded

telegram=123456789&
Services=Реклама&
Subdivision=Digital&
ExecutionDate=2026.02.10&
Description=Текст%20для%20рекламы
```

#### Пример ответа
```json
{
  "status": "ok",
  "number": "MKT-000045"
}
```

---

## Примечания

- Все вызовы собираются в `src/api/itilium_api.py`.
- Валидации форматов на стороне клиента нет, кроме базовых проверок в обработчиках.
- Если `ITILIUM` не отвечает, пользователю показывается общее сообщение об ошибке.

