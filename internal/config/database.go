package config

import (
	"ashno-onepay/internal/model"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

type Database struct {
	Host              string `env:"HOST" json:"host"`
	Port              string `env:"PORT" json:"port"`
	Username          string `env:"USERNAME" json:"username"`
	Driver            string `env:"DRIVER" json:"driver"`
	Password          string `env:"PASSWORD" json:"password"`
	DBName            string `env:"NAME" json:"dBName"`
	ConnectionTimeout string `env:"CONNECTION_TIMEOUT" json:"connectionTimeout"`
	TimeZone          string `env:"TIME_ZONE" json:"timeZone"`
}

func (d *Database) GetDSN() string {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s", d.Host, d.Username, d.Password, d.DBName, d.Port, d.TimeZone)
	fmt.Println(dsn)
	return dsn
}

func (d *Database) GetConnectionTimeout() time.Duration {
	duration, err := time.ParseDuration(d.ConnectionTimeout)
	if err != nil {
		panic(errors.Wrap(err, "Failed to parse connection timeout"))
	}
	return duration
}

var DB *gorm.DB

func InitDatabase() {
	DBCfg := GetConfig().Database
	ctx, cancel := context.WithTimeout(context.Background(), DBCfg.GetConnectionTimeout())
	defer cancel()
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					if DB == nil {
						panic("Timeout to connect database")
					}
				}
				return
			default:
				time.Sleep(time.Second)
			}
		}
	}(ctx)

	var logLevel logger.LogLevel

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logLevel,    // Log level
			Colorful:      true,        // Disable color
		},
	)
	db, err := gorm.Open(postgres.Open(DBCfg.GetDSN()), &gorm.Config{
		Logger:          newLogger,
		TranslateError:  true,
		CreateBatchSize: 100,
	})
	if err != nil {
		panic(errors.Wrap(err, "Failed to connect database"+err.Error()))
	}
	sqlDB, err := db.DB()
	sqlDB.SetMaxIdleConns(500)
	sqlDB.SetMaxOpenConns(2000)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)
	if err != nil {
		panic(errors.Wrap(err, "Failed to connect database"))
	}
	err = db.AutoMigrate(
		model.User{},
	)
	if err != nil {
		panic(errors.Wrap(err, "Failed to migrate database"))
	}

	DB = db
}

func GetDB() *gorm.DB {
	return DB
}
