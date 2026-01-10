# Telegram Export Parser

Микросервис для парсинга JSON экспортов чатов Telegram (Dev2).

## API

```go
// Парсинг одного файла
events, err := parser.ParseSingleFile("result.json")

// Объединение из папок  
merged, err := parser.MergeFromFolders([]string{"export1", "export2"})

// Работа с потоками
events, err := parser.ParseFile(reader, "filename.json")