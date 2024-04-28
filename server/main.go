package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Exchange struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	}
}

var db *sql.DB

func initDB() {
	database, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	db = database
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS exchanges (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            exchange TEXT
        )
    `)
	if err != nil {
		fmt.Println("Error creating table:", err)
		return
	}
}

func main() {
	initDB()
	defer db.Close()

	mux := http.NewServeMux()
	mux.Handle("/cotacao", Exchange{})
	http.ListenAndServe(":8080", mux)

}

func (exchange Exchange) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := GetExchangeUsdToBrl()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = insertExchange(data.USDBRL.Bid)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data.USDBRL.Bid)
}

func insertExchange(bid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	stmt, err := db.PrepareContext(ctx, "INSERT INTO exchanges (exchange) VALUES ($1)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, bid)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		fmt.Println("Time to insert into the database has expired")
		return ctx.Err()
	default:
		return nil
	}
}

func GetExchangeUsdToBrl() (*Exchange, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Time to retrieve BID from the API has expired.")
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var exchange Exchange
	err = json.Unmarshal(body, &exchange)
	if err != nil {
		return nil, err
	}
	return &exchange, nil
}
