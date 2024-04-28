package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	bid, err := getExchange()
	if err != nil {
		fmt.Println("Time to retrieve BID from the API has expired.")
		panic(err)
	}
	saveBidIntoTxt(bid)
}

func saveBidIntoTxt(bid string) {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer file.Close()
	fileSize, err := file.WriteString("Dolar: " + bid)
	fmt.Printf("File created successfully! Size:%d bytes\n", fileSize)
}

func getExchange() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	return string(body), nil
}
