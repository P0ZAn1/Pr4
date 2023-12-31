package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type URLInfo struct {
	ID       int    `json:"id"`
	PID      int    `json:"pid"`
	URL      string `json:"URL"`
	ShortURL string `json:"ShortURL"`
	IP       string `json:"SourceIP"`
	Time     string `json:"TimeINterval"`
	Count    int    `json:"Count"`
}

func FuncHash(key string) int { //Преобразование значения в хеш

	hash := 0
	for i := 0; i < len(key); i++ {
		hash += int(key[i])
	}
	return hash % 512

}
func (hashMap *HashTable) insert(key string, value string) error { //Добавление пары ключ значение в Хеш-таблицу

	newKeyValue := &KeyValue{key, value}
	index := FuncHash(key)
	if hashMap.table[index] == nil {
		hashMap.table[index] = newKeyValue
		return nil
	} else {
		if hashMap.table[index].key == key {
			return errors.New("Такой ключ уже сушествует")
		} else {
			for i := index; i < 512; i++ {
				if hashMap.table[i] == nil {
					hashMap.table[i] = newKeyValue
					return nil
				}
			}
		}
	}
	return errors.New("Неудолось добавить элемент")
}

func (hashMap *HashTable) remuve(key string) error { //Удаление пары ключ значение из Хеш-таблицы

	index := FuncHash(key)
	if hashMap.table[index] == nil {
		return errors.New("Элемент не найден")
	} else if hashMap.table[index].key == key {
		hashMap.table[index] = nil
		return nil
	} else {
		for i := index; i < 512; i++ {
			if hashMap.table[i].key == key {
				hashMap.table[i] = nil
				return nil
			}
		}
	}
	return errors.New("Неудалось удалить элемент")
}

func (hashMap *HashTable) HashGet(key string) (string, error) { //Поиск изначения по ключу в Хеш-таблице

	index := FuncHash(key)
	if hashMap.table[index] == nil {
		return "", errors.New("Элемент не найден")
	} else if hashMap.table[index].key == key {
		return hashMap.table[index].value, nil
	} else {
		for i := index; i < 512; i++ {
			if hashMap.table[i].key == key {
				return hashMap.table[index].value, nil
			}
		}
	}
	return "", errors.New("Элемент не найден")
}

func (hash *HashTable) reaflines(filename string) { //Запись Хеш-таблицы из файла

	content, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			file, createErr := os.Create(filename)
			if createErr != nil {
				panic(createErr)
			}
			file.Close()
			return
		}
		panic(err)
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		parts := strings.Split(line, " ")
		if len(parts) >= 2 {
			key := parts[0]
			value := strings.Join(parts[1:], " ")
			err := hash.insert(key, value)
			if err != nil {
				panic(err)
			}
		}
	}
}

func (hash *HashTable) writeslines(filename string) { //Запись Хеш-таблицы в файл

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	for i := 0; i < 512; i++ {
		if hash.table[i] != nil {
			_, err = file.WriteString(hash.table[i].key + " " + hash.table[i].value + "\n")
			if err != nil {
				panic(err)
			}
			er := hash.remuve(hash.table[i].key)
			if er != nil {
				panic(er)
			}
		}
	}
	return
}

type KeyValue struct {
	key   string
	value string
}

type HashTable struct {
	table [512]*KeyValue
}

func main() {
	fmt.Println("Сервер создан")
	ln, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println("Ошибка при запуске сервера:", err)
		return
	}
	defer ln.Close()

	fmt.Println("Введите действие")
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Ошибка при принятии соединения:", err)
		}

		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	if scanner.Scan() {
		input := scanner.Text()
		args := strings.Fields(input)
		actions := args[0]
		var key string
		var value string
		if len(args) == 2 {
			key = args[1]
			value = ""
		} else if len(args) == 3 {
			key = args[1]
			value = args[2]
		} else if len(args) == 1 {
			key = ""
			value = ""
		}
		var mut sync.Mutex
		hashTable := &HashTable{}
		if actions == "HSET" {
			mut.Lock()
			hashTable.reaflines("Url.txt")
			value = generateUrl()
			er := hashTable.insert(key, value)
			if er != nil {
				hash, erro := hashTable.HashGet(key)
				if erro != nil {

				}
				_, err := conn.Write([]byte(hash + "\n"))
				if err != nil {
					fmt.Println("Ошибка при отправке команды на сервер:", err)
					return
				}
			} else {
				hashTable.insert(value, key)
				_, err := conn.Write([]byte(value + "\n"))
				if err != nil {
					fmt.Println("Ошибка при отправке команды на сервер:", err)
					return
				}
			}
			hashTable.writeslines("Url.txt")
			mut.Unlock()
			mut.Lock()
			orgurl := args[1]
			shortUrl, err := hashTable.HashGet(args[1])
			if err != nil {
				shortUrl = value
			}
			var urlInfo URLInfo

			urlInfos, err := readURLInfoFromJSON("urls.json")
			if err != nil {
				fmt.Println("Error reading JSON file:", err)
				fmt.Println("Internal Server Error", http.StatusInternalServerError)
				return
			}
			for i := range urlInfos {
				if urlInfos[i].URL == orgurl {
					fmt.Printf("Url is exist: http://localhost:5252/%s", urlInfos[i].ShortURL)
					return

				}
			}

			urlInfo = URLInfo{
				ID:       getId(urlInfos),
				PID:      0,
				URL:      orgurl,
				ShortURL: shortUrl,
				IP:       "null",
				Time:     "null",
				Count:    0,
			}

			urlInfos = append(urlInfos, urlInfo)

			err = writeURLInfoToJSON("urls.json", urlInfos)
			if err != nil {
				fmt.Println("Error writing JSON file:", err)
				fmt.Println("Internal Server Error", http.StatusInternalServerError)
				return
			}

			fmt.Printf("Shortened URL: http://localhost:5252/%s", shortUrl)
			fmt.Println()
			mut.Unlock()
		} else if actions == "HGET" {
			mut.Lock()
			hashTable.reaflines("Url.txt")
			remove, er := hashTable.HashGet(key)
			if er == nil {
				_, err := conn.Write([]byte(remove + "\n"))
				if err != nil {
					fmt.Println("Ошибка при отправке команды на сервер:", err)
					return
				}
			} else {
				_, err := conn.Write([]byte(er.Error() + "\n"))
				if err != nil {
					fmt.Println("Ошибка при отправке команды на сервер:", err)
					return
				}
			}
			hashTable.writeslines("Url.txt")
			mut.Unlock()
			mut.Lock()
			orUrl := remove
			urlInfos, err := readURLInfoFromJSON("urls.json")
			if err != nil {
				fmt.Println("Error reading JSON file:", err)
				fmt.Println("Internal Server Error", http.StatusInternalServerError)
			}
			for i := range urlInfos {
				if urlInfos[i].URL == orUrl {
					urlInfos[i].Count++

					urlObject := URLInfo{
						ID:       getId(urlInfos),
						PID:      urlInfos[i].ID,
						URL:      "",
						ShortURL: "",
						IP:       getIP(),
						Time:     time.Now().Format("15:04"),
						Count:    0,
					}

					urlObject.Count++

					urlInfos = counterObjects(urlObject, urlInfos)
				}

			}

			err = writeURLInfoToJSON("urls.json", urlInfos)
			if err != nil {
				fmt.Println("Error writing JSON file:", err)
				return
			}

			connST, err := net.Dial("tcp", ":1234")
			if err != nil {
				log.Println(err.Error())
				return
			}
			_, err = connST.Write([]byte("Report"))
			if err != nil {
				log.Println(err.Error())
				return
			}

			data := make([]byte, 1024)
			n, err := connST.Read(data)

			if string(data[:n]) == "GiveData" {
				jsData, err := readURLInfoFromJSON("urls.json")
				if err != nil {
					log.Println(err.Error())
					return
				}

				req, err := json.MarshalIndent(jsData, "", "  ")
				if err != nil {
					log.Println(err.Error())
					return
				}
				connST.Write(req)
			}

			mut.Unlock()
		}
	}

}
func a(connDB net.Conn) {
	data := make([]byte, 1024)
	n, err := connDB.Read(data)
	if err != nil {
		log.Println(err.Error())
		return
	}
	if string(data[:n]) == "GiveData" {
		jsData, err := readURLInfoFromJSON("urls.json")
		if err != nil {
			log.Println(err.Error())
			return
		}

		req, err := json.MarshalIndent(jsData, "", "  ")
		if err != nil {
			log.Println(err.Error())
			return
		}
		connDB.Write(req)
	}
}

func generateUrl() string {
	const alp = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	url := rand.Intn(6)
	if url <= 1 {
		url += 2
	}
	res := make([]byte, url)

	for i := range res {
		res[i] = alp[rand.Intn(len(alp))]
	}
	return string(res)
}
func readURLInfoFromJSON(filename string) ([]URLInfo, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var urlInfos []URLInfo
	if len(content) == 0 {
		return urlInfos, nil
	}

	err = json.Unmarshal(content, &urlInfos)
	if err != nil {
		return nil, err
	}

	return urlInfos, nil
}
func writeURLInfoToJSON(filename string, urlInfos []URLInfo) error {
	data, err := json.MarshalIndent(urlInfos, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
func getId(urls []URLInfo) int {
	if urls == nil {
		return 1
	}

	id := 1
	for i := range urls {
		if urls[i].ID == id {
			id++
		}
	}

	return id
}
func getIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}

	return ""
}
func counterObjects(urlobj URLInfo, urls []URLInfo) []URLInfo {
	for i := range urls {
		if urls[i].IP == urlobj.IP && urls[i].Time == urlobj.Time && urlobj.PID == urls[i].PID {
			urls[i].Count++
			return urls
		}
	}
	urls = append(urls, urlobj)
	return urls
}
