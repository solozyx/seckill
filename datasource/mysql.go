package datasource

import (
	"database/sql"
	"fmt"
	"reflect"

	_ "github.com/go-sql-driver/mysql"

	"github.com/solozyx/seckill/comm"
	"github.com/solozyx/seckill/conf"
)

// 创建mysql 连接
func NewMysqlConn() (db *sql.DB, err error) {
	sourcename := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8",
		conf.DbMaster.User,
		conf.DbMaster.Pwd,
		conf.DbMaster.Host,
		conf.DbMaster.Port,
		conf.DbMaster.Database)
	db, err = sql.Open(conf.DriverName, sourcename)
	return
}

// 获取返回值，获取一条
func GetResultRow(rows *sql.Rows) map[string]string {
	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([][]byte, len(columns))
	for j := range values {
		scanArgs[j] = &values[j]
	}
	record := make(map[string]string)
	for rows.Next() {
		// 将行数据保存到 record 字典
		rows.Scan(scanArgs...)
		for i, v := range values {
			if v != nil {
				// fmt.Println(reflect.TypeOf(col))
				// record[columns[i]] = string(v.([]byte))
				record[columns[i]] = string(v)
			}
		}
	}
	return record
}

// 获取所有
func GetResultRows(rows *sql.Rows) map[int]map[string]string {
	// 返回所有列
	columns, _ := rows.Columns()
	// 这里表示一行所有列的值 用[]byte表示
	vals := make([][]byte, len(columns))
	// 这里表示一行填充数据
	scans := make([]interface{}, len(columns))
	// 这里scans引用vals 把数据填充到[]byte里
	for k, _ := range vals {
		scans[k] = &vals[k]
	}
	i := 0
	result := make(map[int]map[string]string)
	for rows.Next() {
		//填充数据
		rows.Scan(scans...)
		//每行数据
		row := make(map[string]string)
		//把vals中的数据复制到row中
		for k, v := range vals {
			key := columns[k]
			//这里把[]byte数据转成string
			row[key] = string(v)
		}
		//放入结果集
		result[i] = row
		i++
	}
	return result
}

// 根据结构体中sql标签映射数据到结构体中并且转换类型
func DataToStructByTagSql(data map[string]string, obj interface{}) {
	objValue := reflect.ValueOf(obj).Elem()
	for i := 0; i < objValue.NumField(); i++ {
		//获取sql对应的值
		value := data[objValue.Type().Field(i).Tag.Get("sql")]
		//获取对应字段的名称
		name := objValue.Type().Field(i).Name
		//获取对应字段类型
		structFieldType := objValue.Field(i).Type()
		//获取变量类型，也可以直接写"string类型"
		val := reflect.ValueOf(value)
		var err error
		if structFieldType != val.Type() {
			//类型转换
			val, err = comm.TypeConversion(value, structFieldType.Name()) //类型转换
			if err != nil {

			}
		}
		//设置类型值
		objValue.FieldByName(name).Set(val)
	}
}
