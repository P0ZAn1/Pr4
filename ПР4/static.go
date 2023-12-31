package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strings"
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

func main() {
	fmt.Println("Сервер создан")

	ln, err := net.Listen("tcp", ":1234")
	if err != nil {
		fmt.Println("Ошибка при запуске сервера:", err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Ошибка при принятии соединения:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	rep := make([]byte, 1024)
	n, err := conn.Read(rep)
	fmt.Println(string(rep[:n]))
	if err != nil {
		log.Println(err.Error())
		return
	}
	if string(rep[:n]) == "Report" {
		conn.Write([]byte("GiveData"))
		data := make([]byte, 1024*10)
		n, err := conn.Read(data)
		if err != nil {
			log.Println(err.Error())
			return
		}
		var jsonData []URLInfo
		err = json.Unmarshal(data[:n], &jsonData)
		if err != nil {
			fmt.Println(err)
			return
		}
		var info string
		_ = info

		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			info = scanner.Text()
		}
		pattern := strings.Split(info, " ")

		var one One
		var two Two
		if len(pattern) == 1 {
			OneElement(jsonData, one, pattern)
		} else if len(pattern) == 2 {
			TwoElements(jsonData, two, pattern)
		} else {
			fmt.Println("Wrong input")
		}
	}
}

type One struct {
	one map[string]interface{}
}
type Two struct {
	two map[string]map[string]interface{}
}

func OneElement(jsdata []URLInfo, mapa One, request []string) {
	mapa.one = make(map[string]interface{})

	if len(request) != 1 {
		return
	}

	switch request[0] {
	case "SourceIP":
		for _, i := range jsdata {
			if i.IP == "null" {
				continue
			}
			value, found := mapa.one[i.IP]
			if !found {
				value = i.Count
				mapa.one[i.IP] = value
			} else {
				value = i.Count + value.(int)
				mapa.one[i.IP] = value
			}
		}

		for ip, count := range mapa.one {
			fmt.Printf("%s : %d", ip, count)
			fmt.Println()
		}
	case "Time":
		for _, i := range jsdata {
			if i.Time == "null" {
				continue
			}
			value, found := mapa.one[i.Time]
			if !found {
				value = i.Count
				mapa.one[i.Time] = value
			} else {
				value = i.Count + value.(int)
				mapa.one[i.Time] = value
			}
		}

		for timing, count := range mapa.one {
			fmt.Printf("%s : %d", timing, count)
			fmt.Println()
		}
	case "URL":
		for _, i := range jsdata {
			if i.URL == "" {
				continue
			}
			value, found := mapa.one[i.URL]
			if !found {
				value = i.Count
				mapa.one[i.URL] = value
			} else {
				value = i.Count + value.(int)
				mapa.one[i.URL] = value
			}
		}

		for urls, count := range mapa.one {
			fmt.Printf("%s : %d", urls, count)
			fmt.Println()
		}
	}
}

func TwoElements(jsdata []URLInfo, mapa Two, request []string) {
	mapa.two = make(map[string]map[string]interface{})

	if len(request) != 2 {
		return
	}

	if request[0] == "SourceIP" && request[1] == "Time" {
		for _, i := range jsdata {
			if i.IP == "null" && i.Time == "null" {
				continue
			}
			_, found := mapa.two[i.IP]
			if !found {
				mapa.two[i.IP] = make(map[string]interface{})
			}
			value, found := mapa.two[i.IP][i.Time]
			if !found {
				mapa.two[i.IP][i.Time] = i.Count
			} else {
				value = i.Count + value.(int)
				mapa.two[i.IP][i.Time] = value
			}
		}

		for ip, times := range mapa.two {
			fmt.Println(ip)
			keys := make([]string, 0, len(times))
			for k := range times {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			var totalTransitions int // add a variable to keep track of total transitions for each IP
			for _, timing := range keys {
				count := times[timing].(int)
				totalTransitions += count // increment the total transitions for each timing
				timep, err := time.Parse("15:04", timing)
				if err != nil {
					return
				}
				timep = timep.Add(time.Minute)
				finaltime := timep.Format("15:04")
				fmt.Printf("\t%s - %s : %d\n", timing, finaltime, count)
			}
			fmt.Printf("\tTotal Transitions: %d\n", totalTransitions) // print the total transitions for each IP
		}
	} else if request[0] == "Time" && request[1] == "SourceIP" {
		for _, i := range jsdata {
			if i.IP == "null" && i.Time == "null" {
				continue
			}
			_, found := mapa.two[i.Time]
			if !found {
				mapa.two[i.Time] = make(map[string]interface{})
			}
			value, found := mapa.two[i.Time][i.IP]
			if !found {
				mapa.two[i.Time][i.IP] = i.Count
			} else {
				value = i.Count + value.(int)
				mapa.two[i.Time][i.IP] = value
			}

		}

		keys := make([]string, 0, len(mapa.two))
		for k := range mapa.two {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, timing := range keys {
			timep, err := time.Parse("15:04", timing)
			if err != nil {
				return
			}
			timep = timep.Add(time.Minute)
			finaltime := timep.Format("15:04")
			fmt.Println(timing + "-" + finaltime)
			times := mapa.two[timing]
			subKeys := make([]string, 0, len(times))
			for k := range times {
				subKeys = append(subKeys, k)
			}
			sort.Strings(subKeys)

			var totalTransitions int // add a variable to keep track of total transitions for each time
			for _, ip := range subKeys {
				count := times[ip].(int)
				totalTransitions += count // increment the total transitions for each IP
				fmt.Printf("\t%s : %d\n", ip, count)
			}
			fmt.Printf("\tTotal Transitions: %d\n", totalTransitions) // print the total transitions for each time
		}

	} else if request[0] == "SourceIP" && request[1] == "URL" { // не работает
		var url string
		var count int
		for _, i := range jsdata {
			if i.IP == "null" && i.URL != "" {
				continue
			}

			_, found := mapa.two[i.IP]
			if !found {
				mapa.two[i.IP] = make(map[string]interface{})
			}
			url = urlfind(jsdata, i, mapa.two[i.IP])
			value, found := mapa.two[i.IP][url]
			count = urlCount(jsdata, i, mapa.two[i.IP][url])
			if !found {
				value = count
				mapa.two[i.IP][url] = value
			} else {
				value = count + value.(int)
				mapa.two[i.IP][url] = value
			}
		}

		for ip, urls := range mapa.two {
			fmt.Println(ip)
			for url, count := range urls {
				fmt.Printf("\t%s : %d\n", url, count.(int))
			}
		}
	} else if request[0] == "URL" && request[1] == "SourceIP" {
		var ip string
		var count int
		for _, i := range jsdata {
			if i.URL == "" && i.IP != "null" {
				continue
			}

			_, found := mapa.two[i.URL]
			if !found {
				mapa.two[i.URL] = make(map[string]interface{})
				ip = ipfind(jsdata, i, mapa.two[i.URL])
			}
			value, found := mapa.two[i.URL][ip]
			count = ipcount(jsdata, i, mapa.two[i.URL][ip])
			if !found {
				value = count
				mapa.two[i.URL][ip] = value
			} else {
				value = count + value.(int)
				mapa.two[i.URL][ip] = value
			}
		}

		for urls, ips := range mapa.two {
			fmt.Println(urls)
			for ip, count := range ips {
				fmt.Printf("\t%s : %d\n", ip, count.(int))
			}
		}
	} else if request[0] == "Time" && request[1] == "URL" {
		var url string
		var count int
		totalVisits := make(map[string]int) // Создаем карту для хранения общего количества посещений

		for _, i := range jsdata {
			if i.URL != "" && i.Time == "null" {
				continue
			}

			_, found := mapa.two[i.Time]
			if !found {
				mapa.two[i.Time] = make(map[string]interface{})
			}
			url = urlfind(jsdata, i, mapa.two[i.Time])
			value, found := mapa.two[i.Time][url]
			count = urlCount(jsdata, i, mapa.two[i.Time])
			if !found {
				value = count
				mapa.two[i.Time][url] = value
			} else {
				value = count + value.(int)
				mapa.two[i.Time][url] = value
			}

			// Обновляем общее количество посещений для каждого интервала времени
			totalVisits[i.Time] += count
		}

		var timings []string
		for timing := range mapa.two {
			timings = append(timings, timing)
		}

		// Отсортируйте срез ключей времени
		sort.Strings(timings)

		// Выведите отсортированные данные в консоль
		for _, timing := range timings {
			urls := mapa.two[timing]

			// Преобразуйте время из строки в формат time.Time
			times, err := time.Parse("15:04", timing)
			if err != nil {
				log.Println(err.Error())
				return
			}
			times = times.Add(time.Minute)
			finaltime := times.Format("15:04")

			fmt.Println(timing + "-" + finaltime)
			for url, count := range urls {
				fmt.Printf("\t%s : %d\n", url, count.(int))
			}

			// Выводим общее количество посещений для текущего интервала времени
			fmt.Printf("\tОбщее количество посещений: %d\n", totalVisits[timing])
		}

	} else if request[0] == "URL" && request[1] == "Time" {

	}

}

func ipfind(jsdata []URLInfo, obj URLInfo, mapa map[string]interface{}) string {
	for _, i := range jsdata {
		if i.PID == obj.ID {
			return i.IP
		}
	}
	return ""
}

func ipcount(jsdata []URLInfo, obj URLInfo, mapa interface{}) int {
	for _, i := range jsdata {
		if obj.ID == i.PID {
			return i.Count
		}
	}
	return 0
}

func urlfind(jsdata []URLInfo, obj URLInfo, mapa map[string]interface{}) string {
	for _, i := range jsdata {
		if obj.PID == i.ID {
			return i.URL
		}
	}
	return ""
}
func urlCount(jsdata []URLInfo, obj URLInfo, mapa interface{}) int {
	for _, i := range jsdata {
		if obj.PID == i.ID {
			return obj.Count
		}
	}
	return 0
}

func urlfind3(jsdata []URLInfo, obj URLInfo, mapa interface{}) string {
	for _, i := range jsdata {
		if obj.PID == i.ID {
			return i.URL
		}
	}
	return ""
}
func urlCount3(jsdata []URLInfo, obj URLInfo, mapa interface{}) int {
	for _, i := range jsdata {
		if obj.PID == i.ID {
			return obj.Count
		}
	}
	return 0
}

func timefind(jsdata []URLInfo, obj URLInfo, mapa interface{}) string {
	for _, i := range jsdata {
		if obj.PID == i.ID {
			return obj.Time
		}
	}
	return ""
}

func timeCount(jsdata []URLInfo, obj URLInfo, mapa interface{}) int {
	var count int
	for _, i := range jsdata {
		if obj.PID == i.ID {
			count += i.Count
		}
	}
	return count
}
