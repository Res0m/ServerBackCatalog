package main

// import (
// 	"fmt"
// 	"net/http"
// 	"os"
// )

// func main() {
// 	http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
// 		file, err := os.ReadFile("products.json")
// 		if err != nil {
// 			http.Error(w, "Не удалось загрузить данные", http.StatusInternalServerError)
// 			return
// 		}
// 		w.Header().Set("Content-Type", "application/json")
// 		w.Write(file)
// 	})

// 	http.Handle("/", http.FileServer(http.Dir("../frontend")))

// 	fmt.Println("Каталог работает на порту 3000")
// 	http.ListenAndServe(":3000", nil)
// }
