package xgorm

import (
	"context"
	"fmt"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
	"time"
)

type BaseModel struct {
	ID        uint64 `json:"id" gorm:"primaryKey"`
	CreatedAt int64  `json:"created_at" gorm:"column:created_at;autoCreateTime;comment:创建时间"`
	UpdatedAt int64  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime;comment:更新时间"`
	DeletedAt int64  `json:"deleted_at" gorm:"column:deleted_at;comment:删除时间"`
	Remark    string `gorm:"column:remark;type:varchar(255);not null;default:'';" json:"remark"` // 备注
}

var (
	DsnEmptyErr    = fmt.Errorf("xgorm: %s", "addr is empty")
	RequireConfErr = fmt.Errorf("xgorm: %s", "db, addr, user, passwd is required")
)

type ormLog struct {
	LogLevel logger.LogLevel
}

func (o *ormLog) LogMode(level logger.LogLevel) logger.Interface {
	o.LogLevel = level
	return o
}

func (o *ormLog) Info(ctx context.Context, format string, data ...interface{}) {
	if o.LogLevel < logger.Info {
		return
	}
	logx.WithContext(ctx).Infof(format, data...)
}

func (o *ormLog) Warn(ctx context.Context, format string, data ...interface{}) {
	if o.LogLevel < logger.Warn {
		return
	}
	// logx 没有warn级别的日志，用info代替
	logx.WithContext(ctx).Infof(format, data...)
}

func (o *ormLog) Error(ctx context.Context, format string, data ...interface{}) {
	if o.LogLevel < logger.Error {
		return
	}
	logx.WithContext(ctx).Errorf(format, data...)
}

func (o *ormLog) Trace(ctx context.Context, begin time.Time, fn func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fn()
	logx.WithContext(ctx).WithDuration(elapsed).Infof("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
}

const MysqlDefaultCharset = "utf8mb4"

func MustNewMySql(gormConf *GormConf) *gorm.DB {
	// 判断必填参数
	if gormConf.DB == "" || gormConf.Addr == "" || gormConf.User == "" || gormConf.Passwd == "" {
		panic(RequireConfErr)
	}
	if gormConf.MaxOpenConns == 0 {
		gormConf.MaxOpenConns = 10
	}
	if gormConf.MaxIdleConns == 0 {
		gormConf.MaxIdleConns = 10
	}

	dsn := fmt.Sprintf("%s:%s@"+"tcp(%s)/%s?charset=%s",
		gormConf.User, gormConf.Passwd, gormConf.Addr, gormConf.DB, MysqlDefaultCharset)
	if gormConf.TimeoutSec > 0 {
		// timeout in seconds has "s"
		dsn += fmt.Sprintf("&timeout=%ds", gormConf.TimeoutSec)
	}
	dsn += "&parseTime=true"
	if gormConf.Options != "" {
		dsn += "&" + strings.Trim(gormConf.Options, "&")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:      &ormLog{},
		QueryFields: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxOpenConns(gormConf.MaxOpenConns)
	sqlDB.SetMaxIdleConns(gormConf.MaxIdleConns)
	// plugin
	if gormConf.Metric {
		if err = db.Use(&CustomerPlugin{}); err != nil {
			panic(err)
		}
	}
	if gormConf.Trace {
		if err = db.Use(otelgorm.NewPlugin(
			otelgorm.WithDBName(gormConf.DB),
			otelgorm.WithAttributes(semconv.DBSystemMySQL, attribute.String("db.addr", gormConf.Addr)),
		)); err != nil {
			panic(err)
		}
	}
	return db
}

// Paginate 分页
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// 列表过滤筛选
type Page struct {
	Page    int `json:"page"`
	Limit   int `json:"limit"`
	Filters map[string]interface{}
}

func (p *Page) FindPage(tx *gorm.DB) (*gorm.DB, int64) {
	i := int64(0)
	if p != nil {
		for key, value := range p.Filters {
			fieldValue := reflect.ValueOf(value)
			if !fieldValue.IsZero() {
				if fieldValue.Kind() == reflect.Ptr {
					fieldValue = fieldValue.Elem()
				}
				if strings.Contains(key, "_time") || strings.Contains(key, "_at") {
					t := strings.Split(fieldValue.String(), ",")
					if len(t) == 2 {
						tx = tx.Where(fmt.Sprintf("`%s` between ? and ?", key), t[0], t[1])
					}
				} else if strings.Contains(key, "status") || strings.Contains(key, "state") {
					t := strings.Split(fmt.Sprintf("%v", value), ",")
					if len(t) > 1 {
						tx = tx.Where(fmt.Sprintf("`%s` in (?)", key), t)
					} else {
						tx = tx.Where(fmt.Sprintf("`%s` = ?", key), fieldValue.Interface())
					}
				} else {
					if fieldValue.Kind() == reflect.String {
						fv := fieldValue.String()
						if fv != "" {
							tx = tx.Where(fmt.Sprintf("`%s` like ?", key), "%"+fv+"%")
						}
					} else {
						tx = tx.Where(key, fieldValue.Interface())
					}

				}
			}
		}
		tx.Count(&i)
		if p.Limit > 0 {
			tx = tx.Limit(p.Limit)
		}
		if p.Page > 0 {
			tx = tx.Offset((p.Page - 1) * p.Limit)
		}
	}
	return tx, i
}
