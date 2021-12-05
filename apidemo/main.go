package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name string
}

func main() {

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("faild to connect database")
	}

	db.AutoMigrate(&User{})
	db.Create(&User{Name: "Anuwat"})

	userHandler := UserHandler{db: db}

	r := gin.Default()
	r.GET("/users", userHandler.User)

	r.Run()

}

type UserHandler struct {
	db *gorm.DB
}

func (h *UserHandler) User(c *gin.Context) {
	var u User
	h.db.First(&u)
	c.JSON(200, u)
}
