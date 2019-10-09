package gorm_ex

import (
	"errors"
	"github.com/jinzhu/gorm"
	"reflect"
	"fmt"
)

const table_name =  "$Table_Name$"

type TableNameAble interface {
	TableName() string
}

type DBLogger interface {
	LogInfoc(category, message string)
	LogWarnc(category string, err error, message string)
	LogErrorc(category string, err error, message string)
}

type DBExtension struct {
	*gorm.DB
	logger DBLogger
}

func NewDBWrapper(db *gorm.DB) *DBExtension {
	consoleLogger := newConsoleLogger()
	return &DBExtension{
		DB: db,
		logger: consoleLogger,
	}
}

func (dw *DBExtension) SetDB(db *gorm.DB) {
	dw.DB = db
}

func (dw *DBExtension) SetLogger(logger DBLogger) {
	dw.logger = logger
}

type UpdateAttrs map[string]interface{}

func NewUpdateAttrs(tableName string) UpdateAttrs  {
	attrMap := make(map[string]interface{})
	attrMap[table_name] = tableName
	return attrMap
}

func (dw *DBExtension) GetList(result interface{}, query interface{}, args ...interface{}) error {
	return dw.getListCore(result, "", 0, 0, query, args)
}

func (dw *DBExtension) GetOrderedList(result interface{}, order string, query interface{}, args ...interface{}) error {
	return dw.getListCore(result, order, 0, 0, query, args)
}

func (dw *DBExtension) GetFirstNRecords(result interface{}, order string, limit int, query interface{}, args ...interface{}) error {
	return dw.getListCore(result, order, limit, 0, query, args)
}

func (dw *DBExtension) GetPageRangeList(result interface{}, order string, limit, offset int, query interface{}, args ...interface{}) error {
	return dw.getListCore(result, order, limit, offset, query, args)
}

func (dw *DBExtension) getListCore(result interface{}, order string, limit, offset int, query interface{}, args []interface{}) error {
	var (
		tableNameAble TableNameAble
		ok            bool
	)

	if tableNameAble, ok = query.(TableNameAble); !ok {
		// type Result []*Item{}
		// result := &Result{}
		resultType := reflect.TypeOf(result)
		if resultType.Kind() != reflect.Ptr {
			return errors.New("result is not a pointer")
		}

		sliceType := resultType.Elem()
		if sliceType.Kind() != reflect.Slice {
			return errors.New("result doesn't point to a slice")
		}
		// *Item
		itemPtrType := sliceType.Elem()
		// Item
		itemType := itemPtrType.Elem()

		elemValue := reflect.New(itemType)
		elemValueType := reflect.TypeOf(elemValue)
		tableNameAbleType := reflect.TypeOf((*TableNameAble)(nil)).Elem()

		if elemValueType.Implements(tableNameAbleType) {
			return errors.New("neither the query nor result implement TableNameAble")
		}

		tableNameAble = elemValue.Interface().(TableNameAble)
	}

	db := dw.Table(tableNameAble.TableName()).Where(query, args...)
	if len(order) != 0 {
		db = db.Order(order)
	}

	if offset > 0 {
		db = db.Offset(offset)
	}

	if limit > 0 {
		db = db.Limit(limit)
	}

	if err := db.Find(result).Error; err != nil {
		dw.logger.LogErrorc("mysql", err, fmt.Sprintf("failed to query %s, query is %+v, args are %+v, order is %s, limit is %d", tableNameAble.TableName(), query, args, order, limit))
		return err
	}

	return nil
}


// Update All Fields
func (dw *DBExtension) SaveOne(value TableNameAble) error {
	tableNameAble, ok := value.(TableNameAble)
	if !ok {
		return errors.New("value doesn't implement TableNameAble")
	}

	var err error
	if err = dw.Save(value).Error; err != nil {
		dw.logger.LogErrorc("mysql", err, fmt.Sprintf("Failed to save %s, the value is %+v", tableNameAble.TableName(), value))
	}
	return err
}

// Update selected Fields, if attrs is an object, it will ignore default value field; if attrs is map, it will ignore unchanged field.
func (dw *DBExtension) Update(attrs interface{}, query interface{}, args ...interface{}) error {
	var (
		tableNameAble TableNameAble
		ok            bool
		tableName     string
	)

	if tableNameAble, ok = query.(TableNameAble); ok {
		tableName = tableNameAble.TableName()
	}else if tableNameAble, ok = attrs.(TableNameAble); ok {
		tableName = tableNameAble.TableName()
	} else if attrMap, isUpdateAttrs := attrs.(UpdateAttrs); isUpdateAttrs {
		tableName = attrMap[table_name].(string)
		delete(attrMap, table_name)
	}

	if tableName == "" {
		return errors.New("can't get table name from both attrs and query")
	}

	var err error
	db := dw.Table(tableName).Where(query, args...).Update(attrs)

	if err = db.Error; err != nil {
		dw.logger.LogErrorc("mysql", err, fmt.Sprintf("failed to update %s, query is %+v, args are %+v, attrs is %+v", tableName, query, args, attrs))
	}

	if db.RowsAffected == 0 {
		dw.logger.LogWarnc("mysql",nil, fmt.Sprintf("No rows is updated.For %s, query is %+v, args are %+v, attrs is %+v", tableName, query, args, attrs))
	}

	return err
}

func (dw *DBExtension) GetOne(result interface{}, query interface{}, args ...interface{}) (found bool, err error) {
	var (
		tableNameAble TableNameAble
		ok            bool
	)

	if tableNameAble, ok = query.(TableNameAble); !ok {
		if tableNameAble, ok = result.(TableNameAble); !ok {
			return false, errors.New("neither the query nor result implement TableNameAble")
		}
	}

	err = dw.Table(tableNameAble.TableName()).Where(query, args...).First(result).Error

	if err == gorm.ErrRecordNotFound {
		dw.logger.LogInfoc("mysql", fmt.Sprintf("record not found for query %s, the query is %+v, args are %+v", tableNameAble.TableName(), query, args))
		return false, nil
	}

	if err != nil {
		dw.logger.LogErrorc("mysql", err, fmt.Sprintf("failed to query %s, the query is %+v, args are %+v", tableNameAble.TableName(), query, args))
		return false, err
	}

	return true, nil
}

func (dw *DBExtension) ExecSql(result interface{}, sql string, args ...interface{}) error {
	err := dw.Raw(sql, args...).Scan(result).Error

	if err != nil {
		dw.logger.LogErrorc("mysql", err, fmt.Sprintf("failed to execute sql %s, args are %+v", sql, args))
	}

	return err
}

func (dw *DBExtension) Count(count *int, query interface{}) error {
	return dw.countCore(count, "", query)
}

func (dw *DBExtension) CountBy(count *int, byField string, query interface{}) error {
	return dw.countCore(count, byField, query)
}

func (dw *DBExtension) countCore(count *int, byField string, query interface{}) error {
	tableNameAble, ok := query.(TableNameAble)

	if !ok {
		return errors.New("the query doesn't implement TableNameAble")
	}

	tableName := tableNameAble.TableName()

	db := dw.Table(tableName).Where(query)

	if byField != "" {
		db = db.Select("count(?)", byField)
	}

	if err := db.Count(count).Error; err != nil {
		dw.logger.LogErrorc("mysql", err, fmt.Sprintf("failed to count %s, query is %+v, byField is %s", tableNameAble.TableName(), query, byField))
		return err
	}

	return nil
}
