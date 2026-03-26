# DAX Warehouse Lookup API — полный контракт (для переписывания бота на Go)

API только читает данные из DAX/warehouse (SELECT). Используются те же правила маршрутизации, что и в текущем Telegram-боте.

## Endpoint
`GET /api/warehouse/batches`

### Query params
| Параметр | Тип | Обяз. | Описание |
| --- | --- | --- | --- |
| `code` | string | да | Строка запроса, как в боте. |

### Форматы `code` (маршруты)
- `YY%NNNNN` — партия за год `YY`, без цвета.  
- `YY%NNNNN*` — партия за год `YY`, с цветом/конфигом.  
- `YY%NNNNN@` — партия за год `YY`, с цветом и ФИО.  
- `NNNNN` — партия за **текущий** год (год подставляется), без цвета.  
- `NNNNN*` — текущий год, с цветом/конфигом.  
- `NNNNN@` — текущий год, с цветом и ФИО.  
`YY` — две последние цифры года, `NNNNN` — номер партии (4–10 цифр). `*` — добавить цвет/конфиг, `@` — добавить ФИО.

## Пример ответа
`200 OK`
```json
{
  "success": true,
  "data": [
    {
      "batch": "21%12345",
      "namealias": "Sample product",
      "configid": "CFG-1",         // только для маршрутов с цветом (*)
      "colorid": "RED",            // только для маршрутов с цветом (*)
      "wmslocation": "LOC-01",
      "license": "LIC-001",
      "username": "John Doe"       // только для маршрутов с @
    }
  ],
  "rows": 1
}
```

### Ошибки
- `400` — отсутствует `code` или формат не распознан (no_match).
- `404` — партия не найдена (success=false, rows=0).
- `500` — внутренняя ошибка или недоступность DAX.

## Примеры запросов
- `GET /api/warehouse/batches?code=21%12345`          — без цвета, явный год  
- `GET /api/warehouse/batches?code=21%12345*`         — с цветом  
- `GET /api/warehouse/batches?code=21%12345@`         — с цветом и ФИО  
- `GET /api/warehouse/batches?code=12345`             — текущий год, без цвета  
- `GET /api/warehouse/batches?code=12345*`            — текущий год, с цветом  
- `GET /api/warehouse/batches?code=12345@`            — текущий год, с цветом и ФИО  

## SQL-скрипты (как есть в проекте)
### 1) Базовый поиск (без цвета) — `getSqlQuery()`
```sql
SET NOCOUNT ON;
IF OBJECT_ID('tempdb.dbo.#Initial') IS NOT NULL DROP TABLE #Initial;

SELECT INVENTDIM.INVENTBATCHID as Batch,
      CAST((COALESCE(INVENTTABLE_PT.NAMEALIAS,INVENTTABLE.NAMEALIAS)) as nvarchar(100)) as NAMEALIAS,
      CAST(INVENTDIM.WMSLOCATIONID as nvarchar(50)) as WMSLOCATION,
      CAST(INVENTDIM.LICENSEPLATEID as nvarchar(50)) as LICENSE,
      ROW_NUMBER() over(partition by INVENTDIM.INVENTBATCHID order by INVENTDIM.INVENTBATCHID) as rn
INTO #Initial
FROM INVENTSUM INVENTSUM WITH (READUNCOMMITTED)
LEFT JOIN INVENTDIM INVENTDIM ON INVENTDIM.INVENTDIMID = INVENTSUM.INVENTDIMID
JOIN INVENTTABLE INVENTTABLE ON INVENTSUM.ITEMID = INVENTTABLE.ITEMID
LEFT JOIN ProdTable ProdTable ON INVENTDIM.INVENTBATCHID = ProdTable.ProdID
LEFT JOIN INVENTTABLE INVENTTABLE_PT ON ProdTable.ITEMID = INVENTTABLE_PT.ITEMID
WHERE INVENTSUM.PARTITION = 5637144576 AND INVENTSUM.DATAAREAID = 'rlc'
     AND INVENTDIM.PARTITION = 5637144576 AND INVENTDIM.DATAAREAID = 'rlc'
     AND INVENTSUM.PHYSICALINVENT != 0
     AND INVENTSUM.CLOSEDQTY = 0
     AND INVENTSUM.CLOSED = 0
     AND INVENTDIM.INVENTBATCHID LIKE :number

;WITH RecursiveConcate AS (
    SELECT Batch
        ,CAST(NAMEALIAS AS NVARCHAR(100)) AS NAMEALIAS
        ,CAST(WMSLOCATION AS NVARCHAR(max)) AS WMSLOCATION
        ,CAST(LICENSE AS NVARCHAR(max)) AS LICENSE
        ,2 [rn]
    FROM #Initial AS Initt
    WHERE Initt.rn = 1

    UNION ALL

    SELECT Initt.batch
        ,Initt.NAMEALIAS
        ,IIF(RecCon.WMSLOCATION LIKE '%' + Initt.WMSLOCATION + '%', RecCon.WMSLOCATION, RecCon.WMSLOCATION + ', ' + Initt.WMSLOCATION)
        ,IIF(RecCon.LICENSE LIKE '%' + Initt.LICENSE + '%', RecCon.LICENSE, RecCon.LICENSE + ', ' + Initt.LICENSE)
        ,RecCon.rn + 1
    FROM #Initial AS Initt
    JOIN RecursiveConcate RecCon ON Initt.rn = RecCon.rn
        AND Initt.Batch = RecCon.batch
)
,mRank AS (
    SELECT Batch, NAMEALIAS, WMSLOCATION, LICENSE, MAX(rn) OVER (PARTITION BY batch) AS mrn, rn
    FROM RecursiveConcate
)
SELECT BATCH, NAMEALIAS,
       REPLACE(WMSLOCATION , ' , ', '') as WMSLOCATION,
       REPLACE(LICENSE, ' , ', '') as LICENSE
FROM mRank
WHERE Batch IN (SELECT DISTINCT Batch FROM RecursiveConcate)
      AND rn IN (mrn)
OPTION (MAXRECURSION 32767, RECOMPILE, MAXDOP 1)
DROP TABLE #Initial
```

### 2) Поиск с цветом/конфигом — `getSqlQueryWithColor()`
```sql
SET NOCOUNT ON;
DECLARE @InventBatchId NVARCHAR(10), @Partition BIGINT, @DataAreaId NVARCHAR(5), @ProdInventBatchId NVARCHAR(20);
SET @InventBatchId = :number;
SET @Partition = 5637144576;
SET @DataAreaId = 'rlc';

IF OBJECT_ID('tempdb.dbo.#Initial') IS NOT NULL DROP TABLE #Initial;

SELECT INVENTDIM.INVENTBATCHID as Batch,
    CAST((COALESCE(INVENTTABLE_PT.NAMEALIAS,INVENTTABLE.NAMEALIAS)) as nvarchar(100)) as NAMEALIAS,
    CAST((COALESCE(INVENTDIM_PT.RUK_InventConfigId,INVENTDIM.RUK_InventConfigId)) as nvarchar(50)) as CONFIGID,
    CAST((COALESCE(INVENTDIM_PT.RUK_INVENTCOLORID,INVENTDIM.RUK_INVENTCOLORID)) as nvarchar(100)) as COLORID,
    CAST(INVENTDIM.WMSLOCATIONID as nvarchar(50)) as WMSLOCATION,
    CAST(INVENTDIM.LICENSEPLATEID as nvarchar(50)) as LICENSE,
    ROW_NUMBER() over(partition by INVENTDIM.INVENTBATCHID order by INVENTDIM.INVENTBATCHID) as rn
INTO #Initial
FROM INVENTSUM INVENTSUM WITH (READUNCOMMITTED)
LEFT JOIN INVENTDIM INVENTDIM ON INVENTDIM.INVENTDIMID = INVENTSUM.INVENTDIMID
JOIN INVENTTABLE INVENTTABLE ON INVENTSUM.ITEMID = INVENTTABLE.ITEMID
LEFT JOIN ProdTable ProdTable ON INVENTDIM.INVENTBATCHID = ProdTable.ProdID
LEFT JOIN INVENTDIM INVENTDIM_PT ON INVENTDIM_PT.INVENTDIMID = ProdTable.INVENTDIMID
LEFT JOIN INVENTTABLE INVENTTABLE_PT ON ProdTable.ITEMID = INVENTTABLE_PT.ITEMID
WHERE INVENTSUM.PARTITION = @Partition AND INVENTSUM.DATAAREAID = @DataAreaId
    AND INVENTDIM.PARTITION = @Partition AND INVENTDIM.DATAAREAID = @DataAreaId
    AND INVENTSUM.PHYSICALINVENT != 0
    AND INVENTSUM.CLOSEDQTY = 0
    AND INVENTSUM.CLOSED = 0
    AND INVENTDIM.INVENTBATCHID LIKE @InventBatchId

;WITH RecursiveConcate AS (
    SELECT Batch, CAST(NAMEALIAS AS NVARCHAR(100)) AS NAMEALIAS,
           CAST(CONFIGID AS NVARCHAR(max)) AS CONFIGID,
           CAST(COLORID AS NVARCHAR(max)) AS COLORID,
           CAST(WMSLOCATION AS NVARCHAR(max)) AS WMSLOCATION,
           CAST(LICENSE AS NVARCHAR(max)) AS LICENSE,
           2 [rn]
    FROM #Initial AS Initt WHERE Initt.rn = 1
    UNION ALL
    SELECT Initt.batch, Initt.NAMEALIAS,
           IIF(RecCon.CONFIGID LIKE '%' + Initt.CONFIGID + '%', RecCon.CONFIGID, RecCon.CONFIGID + ', ' + Initt.CONFIGID),
           IIF(RecCon.COLORID LIKE '%' + Initt.COLORID + '%', RecCon.COLORID, RecCon.COLORID + ', ' + Initt.COLORID),
           IIF(RecCon.WMSLOCATION LIKE '%' + Initt.WMSLOCATION + '%', RecCon.WMSLOCATION, RecCon.WMSLOCATION + ', ' + Initt.WMSLOCATION),
           IIF(RecCon.LICENSE LIKE '%' + Initt.LICENSE + '%', RecCon.LICENSE, RecCon.LICENSE + ', ' + Initt.LICENSE),
           RecCon.rn + 1
    FROM #Initial AS Initt
    JOIN RecursiveConcate RecCon ON Initt.rn = RecCon.rn AND Initt.Batch = RecCon.batch
)
,mRank AS (
    SELECT Batch, NAMEALIAS, CONFIGID, COLORID, WMSLOCATION, LICENSE,
           MAX(rn) OVER (PARTITION BY batch) AS mrn, rn
    FROM RecursiveConcate
)
SELECT BATCH, NAMEALIAS,
       REPLACE(CONFIGID , ' , ', '') as CONFIGID,
       REPLACE(COLORID , ' , ', '') as COLORID,
       REPLACE(WMSLOCATION , ' , ', '') as WMSLOCATION,
       REPLACE(LICENSE, ' , ', '') as LICENSE
FROM mRank
WHERE Batch IN (SELECT DISTINCT Batch FROM RecursiveConcate)
      AND rn IN (mrn)
OPTION (MAXRECURSION 32767, RECOMPILE, MAXDOP 1)
DROP TABLE #Initial
```

### 3) Поиск по ячейке (частичный, без цвета) — `getSqlQueryByPartial()`
```sql
SET NOCOUNT ON;
DECLARE @locationId NVARCHAR(10);
SET @locationId = :number;

IF OBJECT_ID('tempdb.dbo.#Initial') IS NOT NULL DROP TABLE #Initial;

SELECT INVENTDIM.INVENTBATCHID as Batch,
  CAST((COALESCE(INVENTTABLE_PT.NAMEALIAS,INVENTTABLE.NAMEALIAS)) as nvarchar(max)) as NAMEALIAS,
  CAST(INVENTDIM.WMSLOCATIONID as nvarchar(max)) as WMSLOCATION,
  CAST(INVENTDIM.LICENSEPLATEID as nvarchar(max)) as LICENSE,
  ROW_NUMBER() over(partition by INVENTDIM.INVENTBATCHID order by INVENTDIM.INVENTBATCHID) as rn
INTO #Initial
FROM INVENTSUM INVENTSUM WITH (READUNCOMMITTED)
LEFT JOIN INVENTDIM INVENTDIM ON INVENTDIM.INVENTDIMID = INVENTSUM.INVENTDIMID
JOIN INVENTTABLE INVENTTABLE ON INVENTSUM.ITEMID = INVENTTABLE.ITEMID
LEFT JOIN ProdTable ProdTable ON INVENTDIM.INVENTBATCHID = ProdTable.ProdID
LEFT JOIN INVENTTABLE INVENTTABLE_PT ON ProdTable.ITEMID = INVENTTABLE_PT.ITEMID
WHERE  INVENTSUM.PARTITION = 5637144576 AND INVENTSUM.DATAAREAID = 'rlc'
      AND INVENTDIM.PARTITION = 5637144576 AND INVENTDIM.DATAAREAID = 'rlc'
      AND INVENTSUM.PHYSICALINVENT != 0
      AND INVENTSUM.CLOSEDQTY = 0
      AND INVENTSUM.CLOSED = 0
      AND INVENTDIM.WMSLOCATIONID LIKE @locationId

;WITH RecursiveConcate AS (
    SELECT Batch, CAST(NAMEALIAS AS NVARCHAR(max)) AS NAMEALIAS,
           CAST(WMSLOCATION AS NVARCHAR(max)) AS WMSLOCATION,
           CAST(LICENSE AS NVARCHAR(max)) AS LICENSE,
           2 [rn]
    FROM #Initial AS Initt WHERE Initt.rn = 1
    UNION ALL
    SELECT Initt.batch, Initt.NAMEALIAS,
           IIF(RecCon.WMSLOCATION LIKE '%' + Initt.WMSLOCATION + '%', RecCon.WMSLOCATION, RecCon.WMSLOCATION + ', ' + Initt.WMSLOCATION),
           IIF(RecCon.LICENSE LIKE '%' + Initt.LICENSE + '%', RecCon.LICENSE, RecCon.LICENSE + ', ' + Initt.LICENSE),
           RecCon.rn + 1
    FROM #Initial AS Initt
    JOIN RecursiveConcate RecCon ON Initt.rn = RecCon.rn AND Initt.Batch = RecCon.batch
)
,mRank AS (
    SELECT Batch, NAMEALIAS, WMSLOCATION, LICENSE,
           MAX(rn) OVER (PARTITION BY batch) AS mrn, rn
    FROM RecursiveConcate
)
SELECT  BATCH, NAMEALIAS,
        REPLACE(WMSLOCATION , ' , ', '') as WMSLOCATION,
        REPLACE(LICENSE, ' , ', '') as LICENSE
FROM mRank
WHERE Batch IN (SELECT DISTINCT Batch FROM RecursiveConcate)
      AND rn IN (mrn)
OPTION (MAXRECURSION 32767, RECOMPILE, MAXDOP 1)
DROP TABLE #Initial
```

### 4) Поиск с цветом и ФИО — `getSqlQueryByPartialWithUsers()`
```sql
SET NOCOUNT ON;
DECLARE @InventBatchId NVARCHAR(10), @Partition BIGINT, @DataAreaId NVARCHAR(5), @ProdInventBatchId NVARCHAR(20);
SET @InventBatchId = :number;
SET @Partition = 5637144576;
SET @DataAreaId = 'rlc';

IF OBJECT_ID('tempdb.dbo.#Initial') IS NOT NULL DROP TABLE #Initial;

WITH PreInitial as (
    SELECT INVENTDIM.INVENTBATCHID AS Batch,
    CAST((COALESCE(INVENTTABLE_PT.NAMEALIAS, INVENTTABLE.NAMEALIAS)) AS NVARCHAR(100)) AS NAMEALIAS,
    CAST((COALESCE(INVENTDIM_PT.RUK_INVENTCOLORID, INVENTDIM.RUK_INVENTCOLORID)) AS NVARCHAR(100)) AS COLORID,
    CAST(INVENTDIM.WMSLOCATIONID AS NVARCHAR(50)) AS WMSLOCATION,
    CAST(INVENTDIM.LICENSEPLATEID AS NVARCHAR(50)) AS LICENSE,
    COALESCE(WWU.USERNAME, '') AS USERNAME,
    IT.DATEPHYSICAL,
    MAX(IT.DatePhysical) OVER (PARTITION BY  INVENTDIM.INVENTBATCHID,INVENTDIM.WMSLOCATIONID) as MaxDate
    FROM INVENTSUM INVENTSUM WITH (READUNCOMMITTED)
    LEFT JOIN INVENTDIM INVENTDIM ON INVENTDIM.INVENTDIMID = INVENTSUM.INVENTDIMID
    JOIN INVENTTABLE INVENTTABLE ON INVENTSUM.ITEMID = INVENTTABLE.ITEMID
    LEFT JOIN ProdTable ProdTable ON INVENTDIM.INVENTBATCHID = ProdTable.ProdID
    LEFT JOIN INVENTDIM INVENTDIM_PT ON INVENTDIM_PT.INVENTDIMID = ProdTable.INVENTDIMID
    LEFT JOIN INVENTTABLE INVENTTABLE_PT ON ProdTable.ITEMID = INVENTTABLE_PT.ITEMID
    LEFT JOIN INVENTTRANS IT ON IT.INVENTDIMID = INVENTSUM.INVENTDIMID AND IT.STATUSRECEIPT IN (1, 2)
    LEFT JOIN INVENTTRANSORIGIN ITO ON ITO.RECID = IT.INVENTTRANSORIGIN
    LEFT JOIN WHSWORKLINE WWL ON WWL.WORKID = ITO.REFERENCEID
    LEFT JOIN WHSWORKUSER WWU ON WWU.USERID = WWL.USERID
    WHERE   ITO.REFERENCECATEGORY = 201
    AND INVENTSUM.PARTITION = @Partition AND INVENTSUM.DATAAREAID = @DataAreaId
    AND INVENTDIM.PARTITION = @Partition AND INVENTDIM.DATAAREAID = @DataAreaId
    AND INVENTDIM_PT.PARTITION = @Partition AND INVENTDIM_PT.DATAAREAID = @DataAreaId
    AND INVENTSUM.PHYSICALINVENT != 0
    AND INVENTSUM.CLOSEDQTY = 0
    AND INVENTSUM.CLOSED = 0
    AND INVENTDIM.INVENTBATCHID LIKE @InventBatchId
)
SELECT *,ROW_NUMBER() OVER (PARTITION BY BATCH ORDER BY BATCH) AS rn
INTO #Initial
FROM PreInitial
WHERE DATEPHYSICAL = MaxDate

;WITH RecursiveConcate AS (
    SELECT Batch, CAST(NAMEALIAS AS NVARCHAR(100)) AS NAMEALIAS,
        CAST(COLORID AS NVARCHAR(max)) AS COLORID,
        CAST(WMSLOCATION AS NVARCHAR(max)) AS WMSLOCATION,
        CAST(LICENSE AS NVARCHAR(max)) AS LICENSE,
        CAST(USERNAME AS NVARCHAR(max)) AS USERNAME,
        2 [rn]
    FROM #Initial AS Initt WHERE Initt.rn = 1

    UNION ALL

    SELECT Initt.batch, Initt.NAMEALIAS,
        IIF(RecCon.COLORID LIKE '%' + Initt.COLORID + '%', RecCon.COLORID, RecCon.COLORID + ', ' + Initt.COLORID),
        IIF(RecCon.WMSLOCATION LIKE '%' + Initt.WMSLOCATION + '%', RecCon.WMSLOCATION, RecCon.WMSLOCATION + ', ' + Initt.WMSLOCATION),
        IIF(RecCon.LICENSE LIKE '%' + Initt.LICENSE + '%', RecCon.LICENSE, RecCon.LICENSE + ', ' + Initt.LICENSE),
        IIF(RecCon.USERNAME LIKE '%' + Initt.USERNAME + '%', RecCon.USERNAME, RecCon.USERNAME + ', ' + Initt.USERNAME),
        RecCon.rn + 1
    FROM #Initial AS Initt
    JOIN RecursiveConcate RecCon ON Initt.rn = RecCon.rn
        AND Initt.Batch = RecCon.batch
)
,mRank AS (
    SELECT Batch, NAMEALIAS, COLORID, WMSLOCATION, LICENSE, USERNAME,
        MAX(rn) OVER (PARTITION BY batch) AS mrn, rn
    FROM RecursiveConcate
)
SELECT BATCH, NAMEALIAS,
    REPLACE(COLORID, ' , ', '') AS COLORID,
    REPLACE(WMSLOCATION, ' , ', '') AS WMSLOCATION,
    REPLACE(LICENSE, ' , ', '') AS LICENSE,
    REPLACE(USERNAME, ' , ', '') AS USERNAME
FROM mRank
WHERE Batch IN (SELECT DISTINCT Batch FROM RecursiveConcate)
    AND rn IN (mrn)
OPTION (MAXRECURSION 0, RECOMPILE, MAXDOP 1)

DROP TABLE #Initial
```

### 5) Простой поиск (старый маршрут, без цвета) — из `WarehouseService::executeCommandFindCell`
```sql
SELECT DISTINCT INVENTDIM.INVENTBATCHID, INVENTTABLE.NAMEALIAS,
                INVENTDIM.RUK_INVENTCOLORID, INVENTDIM.WMSLOCATIONID
FROM INVENTSUM WITH (READUNCOMMITTED)
LEFT JOIN INVENTDIM ON INVENTDIM.INVENTDIMID = INVENTSUM.INVENTDIMID
JOIN INVENTTABLE ON INVENTSUM.ITEMID = INVENTTABLE.ITEMID
WHERE INVENTSUM.PARTITION = 5637144576 AND INVENTSUM.DATAAREAID = 'rlc'
  AND INVENTDIM.PARTITION = 5637144576 AND INVENTDIM.DATAAREAID = 'rlc'
  AND INVENTSUM.PHYSICALINVENT != 0
  AND INVENTSUM.CLOSEDQTY = 0
  AND INVENTSUM.CLOSED = 0
  AND INVENTDIM.INVENTBATCHID LIKE :number
OPTION (RECOMPILE, MAXDOP 1)
```

## Требования к реализации в Go
- Перед любым запросом выполнять «санитарию» сессии (важно для пулла):
  ```
  SET NOCOUNT ON; SET XACT_ABORT ON; SET IMPLICIT_TRANSACTIONS OFF;
  SET LOCK_TIMEOUT 15000; IF @@TRANCOUNT > 0 ROLLBACK;
  ```
- Все запросы — только SELECT; обязательна опция `OPTION (RECOMPILE, MAXDOP 1)`.
- Маршрутизация `code` повторяет правила выше; при отсутствии года подставляется текущий (две последние цифры).
- Рекомендованный ответ: `success`, `data` (массив), `rows`, `message` (в ошибках).

Этого достаточно, чтобы переписать бота на Go и вызывать API Mini App без изменений в БД.
