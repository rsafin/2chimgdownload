package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

const baseURL = "https://2ch.hk"

//URLCollection ...
type URLCollection struct {
	mutex sync.Mutex
	list  map[string]bool
}

//Append ...
func (collection *URLCollection) Append(url string) {
	collection.mutex.Lock()
	if _, ok := collection.list[url]; !ok {
		collection.list[url] = false
	}
	collection.mutex.Unlock()
}

//Set ...
func (collection *URLCollection) Set(url string, state bool) {
	collection.mutex.Lock()
	if _, ok := collection.list[url]; ok {
		collection.list[url] = state
	}
	collection.mutex.Unlock()
}

//getCount ...
func (collection *URLCollection) getCount() int {
	return len(collection.list)
}

//Downloader ...
func Downloader(collection *URLCollection, dir string, channel chan int) {
	for url, state := range collection.list {
		if !state {
			collection.Set(url, true)
			parts := strings.Split(url, "/")
			download(url, dir, parts[len(parts)-1])

			channel <- 1
		}
	}
}

//getImageUrls
func getImageUrls(url string) (*URLCollection, error) {
	res, err := http.Get(url)
	urlCollection := URLCollection{list: make(map[string]bool)}

	if err != nil {
		return &urlCollection, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return &urlCollection, err
	}

	re := regexp.MustCompile(`<img.*data-src="(.*)">`)

	imgUrls := re.FindAllSubmatch(body, -1)

	for _, url := range imgUrls {
		urlCollection.Append(baseURL + string(url[1]))
	}

	return &urlCollection, nil
}

//download
func download(url string, dir string, fileName string) error {
	response, err := http.Get(url)

	if err != nil {
		return err
	}

	defer response.Body.Close()

	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}

	file, err := os.Create(dir + fileName)

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}

//clearTerminal ...
func clearTerminal() {
	cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
	cmd.Stdout = os.Stdout
	cmd.Run()
}

//renderProgressbar ...
func renderProgressbar(size int, total int, current int) {
	full := "▰ "
	empty := "▱ "
	progressbar := ""

	procentage := int((100.0 * current) / total)
	fullness := size * procentage / 100

	for i := 0; i < size; i++ {
		if i < fullness {
			progressbar += full
		} else {
			progressbar += empty
		}
	}

	progressbar = fmt.Sprintf("[%s] %d%% | %d/%d ", progressbar, procentage, current, total)
	println(progressbar)
}

func main() {
	urls, err := getImageUrls("https://2ch.hk/b/res/233362968.html")

	if err != nil {
		print(err)
	}

	total := urls.getCount()
	channel := make(chan int)

	go Downloader(urls, "E:/2ch/img/", channel)
	go Downloader(urls, "E:/2ch/img/", channel)
	go Downloader(urls, "E:/2ch/img/", channel)
	go Downloader(urls, "E:/2ch/img/", channel)

	go Downloader(urls, "E:/2ch/img/", channel)
	go Downloader(urls, "E:/2ch/img/", channel)
	go Downloader(urls, "E:/2ch/img/", channel)
	go Downloader(urls, "E:/2ch/img/", channel)

	count := 0
	for i := range channel {
		count += i
		clearTerminal()
		renderProgressbar(50, total, count)
		if total == count {
			close(channel)
		}
	}
}
