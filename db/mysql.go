package db

import (
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"time"
)

func ConnectMysql(dsn string, replicas ...string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.WithField("role", "mysql").Error(err)
		return nil
	}
	var splittingDB []gorm.Dialector
	for _, dsn := range replicas {
		splittingDB = append(splittingDB, mysql.Open(dsn))
	}
	err = db.Use(dbresolver.Register(dbresolver.Config{Replicas: splittingDB}).
		SetConnMaxIdleTime(time.Hour).
		SetConnMaxLifetime(24 * time.Hour).
		SetMaxIdleConns(100).
		SetMaxOpenConns(200))

	if err != nil {
		logrus.WithField("role", "connect-mysql").Error(err)
		return nil
	}
	return db
}
