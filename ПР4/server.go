package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

var mu sync.Mutex

func isValidUrl(token string) bool {
	_, err := url.ParseRequestURI(token)
	if err != nil {
		return false
	}
	u, err := url.Parse(token)
	if err != nil || u.Host == "" {
		return false
	}
	return true
}

func shortHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Ошибка!", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	orglURL := r.Form.Get("URL")
	if !isValidUrl(orglURL) {
		fmt.Fprintf(w, "Url is wrong: %s", orglURL)
		return
	}

	if orglURL == "" {
		http.Error(w, "URL ошибка адреса", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	var ShortURl string = ""
	conn, err := net.Dial("tcp", ":6379")
	if err != nil {
		fmt.Println("Ошибка при подключении к серверу")
		log.Println(err.Error())
		return
	}

	_, err = conn.Write([]byte("HSET " + orglURL + " " + ShortURl + "\n"))
	if err != nil {
		fmt.Println("Ошибка при передачи данных")
		log.Println(err.Error())
		return
	}
	reader, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("ОШибка при чтении с сервера")
		log.Println(err.Error())
		return

	}
	defer conn.Close()

	ShortURl = reader
	Parts := strings.Split(ShortURl, "\n")
	ShortUrl := Parts[0]

	fmt.Fprintf(w, "Shortened URL: http://localhost:5252/%s", ShortUrl)
}

var idCounter int

func Direction(w http.ResponseWriter, r *http.Request) {
	var mut sync.Mutex

	conn, err := net.Dial("tcp", ":6379")
	if err != nil {
		fmt.Println("Ошибка при подключении к серверу")
		log.Println(err.Error())
		return
	}
	defer conn.Close()
	mut.Lock()
	defer mut.Unlock()

	shortUrl := strings.TrimPrefix(r.URL.Path, "/")
	_, err = conn.Write([]byte("HGET " + shortUrl + "\n"))
	if err != nil {
		fmt.Printf("Ошибка при подключении к серверу: %s\n", err.Error())
		log.Println(err.Error())
		return
	}

	original, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("ОШибка при чтении с сервера")
		log.Println(err.Error())
		return
	}

	if original != "Элемент не найден" {
		http.Redirect(w, r, original, http.StatusFound)
	} else {
		http.NotFound(w, r)
	}

}

func ReportHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := net.Dial("tcp", ":1234")
	if err != nil {
		fmt.Println("Ошибка при подключении к серверу")
		log.Println(err.Error())
		return
	}
	defer conn.Close()
	rep := "Report"
	_, err = conn.Write([]byte(rep))
	if err != nil {
		fmt.Println("Ошибка при подключении к серверу")
		log.Println(err.Error())
		return
	}

}

func main() {
	fmt.Println("Сервис сокращения ссылок запущен...")
	http.HandleFunc("/shorten", shortHandler)
	http.HandleFunc("/", Direction)
	http.ListenAndServe(":5252", nil)
}
