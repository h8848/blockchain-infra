package xutil

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func StructToMap(s interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Map {
		return s.(map[string]interface{})
	}
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() { // 判断字段是否可导出
			fieldName := field.Tag.Get("json")
			if fieldName == "" {
				fieldName = field.Name
				if field.Anonymous {
					// 如果是匿名字段，递归将其转换为 map
					anonymousMap := StructToMap(v.Field(i).Interface())
					for k, v := range anonymousMap {
						result[k] = v
					}
					continue
				}
			} else {
				f := strings.Split(fieldName, ",")
				if len(f) > 0 {
					fieldName = f[0]
				}
			}
			fieldValue := v.Field(i).Interface()
			if fieldValue == nil {
				continue
			}
			// 如果字段是指针类型，而指向的值是结构体，则递归将其转换为 map
			if field.Type.Kind() == reflect.Ptr {
				elem := reflect.Indirect(v.Field(i))
				if elem.Kind() == reflect.Struct {
					fieldValue = StructToMap(elem.Interface())
				}
			}
			if field.Type.Kind() == reflect.Struct {
				fieldValue = StructToMap(v.Field(i).Interface())
			}
			result[fieldName] = fieldValue
		}
	}
	return result
}

func ToInt64Must(value interface{}) int64 {
	toInt64, err := ToInt64(value)
	if err != nil {
		return 0
	}
	return toInt64
}

func ToInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		intValue, err := StringToInt64(v)
		if err != nil {
			return 0, err
		}
		return intValue, nil
	default:
		return 0, fmt.Errorf("无法将类型 %T 转换为 int64", value)
	}
}

func StringToInt64Must(str string) int64 {
	toInt64, err := StringToInt64(str)
	if err != nil {
		return 0
	}
	return toInt64
}

func StringToInt64(str string) (int64, error) {
	intValue, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return intValue, nil
}

func Int64ToBytes(n int64) []byte {
	bytes := make([]byte, 8) // int64 占用 8 个字节
	binary.BigEndian.PutUint64(bytes, uint64(n))
	return bytes
}

func BytesToInt64(data []byte) int64 {
	if len(data) != 8 {
		return 0
	}
	return int64(binary.BigEndian.Uint64(data))
}

func Float64ToBytes(f float64) []byte {
	bits := math.Float64bits(f)
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, bits)
	return bytes
}

func BytesToFloat64(bytes []byte) float64 {
	if len(bytes) != 8 {
		return 0
	}
	bits := binary.BigEndian.Uint64(bytes)
	return math.Float64frombits(bits)
}

func ToString(val interface{}) string {
	return fmt.Sprintf("%v", val)
}

// Int64ToTime int64转时间格式
func Int64ToTime(t int64) (timeStr string) {
	if t == 0 {
		return
	}
	ctime := time.Unix(t, 0)
	timeStr = ctime.Format("2006-01-02 15:04:05")
	return
}

func ToLocalDateBySecond(second int64, local string) string {
	if second == 0 {
		return ""
	}
	location, err := time.LoadLocation(local)
	if err != nil {
		return ""
	}
	return time.Unix(second, 0).In(location).Format("2006-01-02 15:04:05")
}

func ToLocalDateByMilliSecond(milliSecond int64, local string) string {
	if milliSecond == 0 {
		return ""
	}
	location, err := time.LoadLocation(local)
	if err != nil {
		return ""
	}
	return time.UnixMilli(milliSecond).In(location).Format("2006-01-02 15:04:05")
}

func ToBool(val interface{}) bool {
	switch val.(type) {
	case bool:
		return val.(bool)
	case string:
		return val.(string) == "true" || val.(string) == "1"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return val.(int) == 1
	default:
		return false
	}

}
