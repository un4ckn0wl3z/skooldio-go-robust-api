package main

import (
	"context"
	"demo/todoapi/auth"
	"demo/todoapi/todo"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	buildcommit = "dev"
	buildtime   = time.Now().String()
)

func main() {

	_, err := os.Create("live")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove("live")

	err = godotenv.Load("local.env")
	if err != nil {
		log.Println("Error loading local.env file")
	}

	db, err := gorm.Open(sqlite.Open(os.Getenv("DB_CONN")), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&todo.Todo{})

	r := gin.Default()

	r.GET("/healthz", func(c *gin.Context) {
		c.Status(200)
	})

	r.GET("/limitz", limitedHandler)

	r.GET("/x", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"buildcommit": buildcommit,
			"buildtime":   buildtime,
		})
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/tokenz", auth.AccessToken([]byte(os.Getenv("SIGN"))))

	protected := r.Group("", auth.Protect([]byte(os.Getenv("SIGN"))))

	handler := todo.NewTodoHandler(db)
	protected.POST("/todos", handler.NewTask)
	protected.GET("/todos", handler.List)
	protected.DELETE("/todos/:id", handler.Remove)

	//r.Run()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	s := &http.Server{
		Addr:           ":" + os.Getenv("PORT"),
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()
	stop()

	fmt.Println("shutting down gracefully, press Ctrl+C again to force")
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(timeoutCtx); err != nil {
		fmt.Println(err)
	}

}

var limiter = rate.NewLimiter(5, 5)

func limitedHandler(c *gin.Context) {
	if !limiter.Allow() {
		c.AbortWithStatus(http.StatusTooManyRequests)
		return
	}
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
