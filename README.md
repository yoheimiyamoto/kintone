## 概要
KintoneのSDK

## 使用方法
```
repo := NewRepository(os.Getenv("KINTONE_DOMAIN"), os.Getenv("KINTONE_ID"), os.Getenv("KINTONE_PASSWORD"), &RepositoryOption{MaxConcurrent: 90})

records = []*Record

fs := kintone.Fields{
    "field1":        kintone.SingleLineTextField("hello"),
}

records = append(records, &Record{Fields: fs})
appID = {YOUR APP ID}

_, err := repo.AddRecords(nil, appID, records)
if err != nil {
    // エラーハンドリング
}
```