package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

type Product struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Price       int    `json:"price"`
	Description string `json:"description"`
}

var products []Product

func loadProducts() {
	file, _ := os.ReadFile("../backend-catalog/products.json")
	json.Unmarshal(file, &products)
}

func saveProducts() {
	file, _ := json.MarshalIndent(products, "", "  ")
	os.WriteFile("../backend-catalog/products.json", file, 0644)
}

func main() {
	loadProducts()

	// Обслуживание статических файлов (admin.html и стили)
	fs := http.FileServer(http.Dir("../frontend"))
	http.Handle("/", fs)

	http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var p Product
			json.NewDecoder(r.Body).Decode(&p)
			p.ID = len(products) + 1
			products = append(products, p)
			saveProducts()
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(products)
	})

	http.HandleFunc("/products/", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(r.URL.Path[len("/products/"):])
		if r.Method == "PUT" {
			var p Product
			json.NewDecoder(r.Body).Decode(&p)
			for i, v := range products {
				if v.ID == id {
					products[i] = p
					saveProducts()
				}
			}
		} else if r.Method == "DELETE" {
			for i, v := range products {
				if v.ID == id {
					products = append(products[:i], products[i+1:]...)
					saveProducts()
				}
			}
		}
	})

	fmt.Println("Админ-панель работает на порту 8080")
	http.ListenAndServe(":8080", nil)
}
