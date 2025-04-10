package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"strconv"
	"time"

	"github.com/flaviojohansson/goexpert-rate-limiter/internal/limiter"
	"github.com/flaviojohansson/goexpert-rate-limiter/internal/middleware"
	"github.com/flaviojohansson/goexpert-rate-limiter/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func userHandler(c *gin.Context) {

	user := c.Param("userId")
	c.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("Olá %s", user)})

}

func main() {

	godotenv.Load()

	redisAddr := os.Getenv("REDIS_ADDR")
	ipLimit, _ := strconv.Atoi(os.Getenv("DEFAULT_IP_LIMIT"))
	tokenLimit, _ := strconv.Atoi(os.Getenv("DEFAULT_TOKEN_LIMIT"))
	blockTime, _ := time.ParseDuration(os.Getenv("BLOCK_DURATION_SECONDS") + "s")
	windowDuration, _ := time.ParseDuration(os.Getenv("WINDOW_DURATION") + "s")

	log.Printf("Inicalizando Redis: %s, ipLimit: %d, tokenLimit: %d, windowDuration: %d, Blocktime: %d", redisAddr, ipLimit, tokenLimit, windowDuration, blockTime)

	// storage.RedisStorage implementa storage.StorageInterface
	redisStore := storage.NewRedisStorage(redisAddr, "")

	limiter := limiter.NewLimiter(redisStore)

	r := gin.Default()

	r.Use(middleware.RateLimiter(limiter, ipLimit, tokenLimit, windowDuration, blockTime))

	r.GET("/user/:userId", userHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		server.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	log.Println("Shutting down server ...")

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Could not gracefully shutdown server: %v\n", err)
	}

	log.Println("Server stopped")
}
