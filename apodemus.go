package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	evdev "github.com/gvalkov/golang-evdev"
	_ "github.com/mattn/go-sqlite3"
)

func trackMouse(device string, wg *sync.WaitGroup, resultChan chan<- string) {
	defer wg.Done()
	dev, err := evdev.Open(device)
	if err != nil {
		log.Fatalf("Failed to open mouse device: %v", err)
	}
	defer dev.Release()

	for {
		used := 0
		events, err := dev.Read()
		if err != nil {
			log.Printf("Error reading mouse events: %v", err)
		} else if len(events) > 0 {
			used = 1
		}

		resultChan <- fmt.Sprintf("mouse,%d", used)

		time.Sleep(1 * time.Second)
	}
}

func trackKeyboard(device string, wg *sync.WaitGroup, resultChan chan<- string) {
	defer wg.Done()
	dev, err := evdev.Open(device)
	if err != nil {
		log.Fatalf("Failed to open keyboard device: %v", err)
	}
	defer dev.Release()

	for {
		used := 0
		events, err := dev.Read()
		if err != nil {
			log.Printf("Error reading keyboard events: %v", err)
		} else if len(events) > 0 {
			used = 1
		}

		resultChan <- fmt.Sprintf("keyboard,%d", used)

		time.Sleep(1 * time.Second)
	}
}

func main() {
	mouseDevice := "/dev/input/event16"    // Replace with your mouse event device
	keyboardDevice := "/dev/input/event19" // Replace with your keyboard event device

	var wg sync.WaitGroup
	resultChan := make(chan string, 2)
	wg.Add(2)
	go trackMouse(mouseDevice, &wg, resultChan)
	go trackKeyboard(keyboardDevice, &wg, resultChan)

	db, err := sql.Open("sqlite3", "./usage_stats.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS usage_stats (
        timestamp DATETIME,
        device TEXT,
        used INTEGER
    )`)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	stmt, err := db.Prepare("INSERT INTO usage_stats(timestamp, device, used) VALUES(?, ?, ?)")
	if err != nil {
		log.Fatalf("Failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	go func() {
		for {
			resultChan <- "active,1"
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		for result := range resultChan {
			fmt.Printf("%s\n", result)
			parts := strings.Split(result, ",")
			_, err := stmt.Exec(time.Now(), parts[0], parts[1])
			if err != nil {
				log.Printf("Failed to insert into database: %v", err)
			}
		}
	}()

	fmt.Println("Tracking mouse and keyboard. Press Ctrl+C to exit.")
	wg.Wait()
}
