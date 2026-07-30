SELECT
s.name AS SchemaName,
t.name AS TableName,
FROM sys.tables t
INNER JOIN sys.shemas s
ON t.shema_id = s.schema_id
ORDER BY
s.name,
t.name;