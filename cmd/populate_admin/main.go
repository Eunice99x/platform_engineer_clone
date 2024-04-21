package main

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"golang.org/x/crypto/bcrypt"
	"log"
	"platform_engineer_clone/dependency_injection/dic"
)

func main() {
	logger := common.GetLogger(context.Background())
	builder, err := dic.NewBuilder()
	if err != nil {
		log.Fatalf("error trying to initialize the builder: %v", err.Error())
	}
	ctn := builder.Build()

	mysqlConnection, err := ctn.SafeGetMysqlConnection()
	if err != nil {
		log.Fatalf("error getting the mysql_connection from the container: %v", err.Error())
	}

	name := "Admin User"
	email := "admin@gmail.com"
	unhashedPassword := "123456"
	password, _ := strings.Encrypt(unhashedPassword)
	if err = bcrypt.CompareHashAndPassword([]byte(password), []byte(unhashedPassword)); err != nil {
		log.Fatalf("Failed to synchronize passwords: %v", err)
	}

	newUser := models_schema.User{
		Name:     name,
		Email:    email,
		Password: password,
	}
	err = newUser.Insert(mysql.BoilCtx, mysqlConnection.DB, boil.Infer())
	if err != nil {
		log.Fatalf("error inserting a dummy admin user: %v", err.Error())
	}

	logger.WithFields(logrus.Fields{
		email:    "admin@gmail.com",
		password: "123456",
	}).Info("add_admin_user")

}
