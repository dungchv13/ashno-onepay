package config

import (
	"ashno-onepay/internal/model"
	"context"
	"errors"
	"fmt"
	errs "github.com/pkg/errors"
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
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s", d.Host, d.Username, d.Password, d.DBName, d.Port)
	log.Println(dsn)
	return dsn
}

func (d *Database) GetConnectionTimeout() time.Duration {
	duration, err := time.ParseDuration(d.ConnectionTimeout)
	if err != nil {
		panic(errs.Wrap(err, "Failed to parse connection timeout"))
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
		panic(errs.Wrap(err, "Failed to connect database"+err.Error()))
	}
	sqlDB, err := db.DB()
	sqlDB.SetMaxIdleConns(500)
	sqlDB.SetMaxOpenConns(2000)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)
	if err != nil {
		panic(errs.Wrap(err, "Failed to connect database"))
	}
	err = db.AutoMigrate(
		model.Registration{},
		model.RegistrationOption{},
		model.AccompanyPersonDB{},
	)
	if err != nil {
		panic(errs.Wrap(err, "Failed to migrate database"))
	}
	if err := seedRegistrationOptions(db); err != nil {
		log.Fatalf("failed to seed registration options: %v", err)
	}

	DB = db
}

func seedRegistrationOptions(db *gorm.DB) error {
	options := []model.RegistrationOption{
		{Category: string(model.DoctorCategory), Subtype: string(model.EarlyBird), FeeUSD: 500, FeeVND: 1800000, Active: true},
		{Category: string(model.DoctorCategory), Subtype: string(model.Regular), FeeUSD: 600, FeeVND: 2200000, Active: true},
		{Category: string(model.DoctorCategory), Subtype: string(model.OnSite), FeeUSD: 700, FeeVND: 3000000, Active: true},
		{Category: string(model.StudentCategory), Subtype: "", FeeUSD: 300, FeeVND: 1500000, Active: true},
		{Category: string(model.DoctorAndDinnerCategory), Subtype: string(model.EarlyBird), FeeUSD: 600, FeeVND: 2800000, Active: true},
		{Category: string(model.DoctorAndDinnerCategory), Subtype: string(model.Regular), FeeUSD: 700, FeeVND: 3200000, Active: true},
		{Category: string(model.DoctorAndDinnerCategory), Subtype: string(model.OnSite), FeeUSD: 800, FeeVND: 4000000, Active: true},
		{Category: string(model.StudentAndDinnerCategory), Subtype: "", FeeUSD: 400, FeeVND: 2500000, Active: true},
	}

	for _, opt := range options {
		var existing model.RegistrationOption
		query := db.Model(&model.RegistrationOption{})
		if opt.Category != "" {
			query = query.Where("category = ?", opt.Category)
		}
		if opt.Subtype != "" {
			query = query.Where("subtype = ?", opt.Subtype)
		}
		if err := query.First(&existing).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.Create(&opt).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func GetDB() *gorm.DB {
	return DB
}
