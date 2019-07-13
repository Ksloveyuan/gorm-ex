# gorm-ex
Based on my best practice, extend gorm with some frequent used functions, e.g:
- SaveOne()
- Update()
- GetOne()
- GetList()
- GetOrderedList()
- GetPageRangeList()


# Examples
## GetOne
```go
query := Topic{TopicId: topicId}
result := Topic{}

if found, err := dw.GetOne(&result, query); !found {
	//not found
    if err != nil {
    	// has error
        return err
    }
}

```

## GetList
```go
query := InstInfo{
    Valid:1,
}
result:= make([]*InstInfo, 0)

if err := dw.GetList(&result, query); err != nil{
    // error handling
    return err
}
```

```go
tids := []int{1, 2, 3, 4}
result := []*TuChart{}

if err :=dw.GetList(&result, "valid = 1 and tid in (?)", tids); err != nil{
    //error handling
    return err
}
```

## GetOrderedList
It shares the same usage of GetList, except that there is one more order field.

```go
query := InstInfo{
    Valid:1,
}
result:= make([]*InstInfo, 0)

if err := dw.GetOrderedList(&result,"create_time desc", query); err != nil{
    // error handling
    return err
}
```

## GetPageRangeList
```go
result := []*MpFeedInfo{}

if err :=dw.GetPageRangeList(&result, "update_time asc", limit, offset,
        "valid = 1 and update_time > ? and update_time < ?", startTime, endTime);err != nil{
    return err
}
```

## SaveOne
Update All Fields, the object's primary_key, defined in gorm format definition, must have value.

```go
instInfo  := InstInfo{
    InstId:req.InstId,
    Name:req.Name,
    Contact:req.Contact,
    Address:req.Address,
    Email:req.Email,
    Phone:req.Phone,
    OwnerId: req.OwnerId,
    Valid: req.Valid,
}

if err := dw.SaveOne(&instInfo); err != nil{
    // error handling
    return err
}
```

## Update
Update partial Fields, if attrs is an object, it will ignore default value field; if attrs is map, it will ignore unchanged field.

```go
//todo: add later
```

## ExecuteSql

```go
sql := "SELECT ta.*, tac.content FROM tucloud.tu_article ta inner join tucloud.tu_article_content tac on ta.article_id = tac.article_id and ta.valid = 1 and ta.article_id in (?)"
result := []*TuArticleWithContent{}

if err:= dw.ExecSql(&result, sql, tuArticleIds);err != nil{
    // error handling
    return err
}

```
