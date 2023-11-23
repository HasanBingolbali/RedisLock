package main

import (
	"RedisLock/internal/go_redis_concurrency/redis"
	"context"
	"github.com/gofiber/fiber/v2"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const (
	total_clients = 30
)
const companyId = "TestCompanySL"

func main() {
	// Create a Fiber app
	myFiberApp := fiber.New(fiber.Config{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  5 * time.Second,
		BodyLimit:    20 * 1024 * 1024,
		Concurrency:  100000,
	})
	repository := redis.NewRepository("redis-service.default.svc.cluster.local:6379")

	myFiberApp.Post("/add", func(ctx *fiber.Ctx) error {
		shareNumber := ctx.Query("share")
		shareNumberInt, _ := strconv.Atoi(shareNumber)
		err := repository.PublishShares(context.Background(), companyId, shareNumberInt)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "internal Error",
			})
		}
		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "successfully added shares",
		})
	})

	myFiberApp.Post("/buy", func(ctx *fiber.Ctx) error {
		userId := ctx.Query("userId")
		err := repository.BuySharesWithRedisLock(ctx.Context(), userId, companyId, 1)
		if err != nil {
			// Handle the error and respond accordingly
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":  err.Error(),
				"userId": userId,
				"status": "failed to buy shares",
			})
		}
		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"userId":       userId,
			"companyId":    companyId,
			"sharesBought": 1,
			"message":      "successfully bought shares",
		})
	})

	// Start the server
	go func() {
		if err := myFiberApp.Listen(":8080"); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
	if err := myFiberApp.Shutdown(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
	log.Println("Server gracefully stopped")
}

// The script to send multiple request for test purposes.
/*
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

func sendRequest(client *http.Client, userID int, wg *sync.WaitGroup, mutex *sync.Mutex, totalResponseTime *time.Duration, successCount *int) {
	defer wg.Done()

	startTime := time.Now()
	response, err := client.Post(fmt.Sprintf("http://127.0.0.1/buy?userId=%d", userID), "", nil)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer response.Body.Close()

	elapsedTime := time.Since(startTime)
	mutex.Lock()
	*totalResponseTime += elapsedTime
	if response.StatusCode == http.StatusOK {
		*successCount++
	}
	mutex.Unlock()

	// Optional: Read and print response body
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Printf("User ID %d: %s, Response time: %v\n", userID, body, elapsedTime)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run script.go <num_users>")
		return
	}

	numUsers, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Invalid number of users:", err)
		return
	}

	var wg sync.WaitGroup
	var mutex sync.Mutex
	var totalResponseTime time.Duration
	successCount := 0
	client1 := &http.Client{Transport: &http.Transport{MaxConnsPerHost: 200}}
	for i := 1; i <= numUsers; i++ {
		wg.Add(1)
		go sendRequest(client1, i, &wg, &mutex, &totalResponseTime, &successCount)
	}
	wg.Wait()

	averageResponseTime := totalResponseTime / time.Duration(numUsers)
	fmt.Printf("\nAverage Response Time: %v\n", averageResponseTime)
	fmt.Printf("Total Successful Requests: %d out of %d\n", successCount, numUsers)
}

*/
