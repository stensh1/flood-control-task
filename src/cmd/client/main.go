package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	client := &http.Client{}

	c := http.Cookie{
		Name:  "cookie",
		Value: "test",
	}

	req, err := http.NewRequest("GET", "http://localhost:8080/check", nil)
	req.AddCookie(&c)
	resp, err := client.Do(req)

	data, err := io.ReadAll(resp.Body) // читаем ответ
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("%s", data) // печатаем ответ как строку
}
