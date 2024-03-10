package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

const baseURL = "https://2ch.hk"

// URLCollection ...
type URLCollection struct {
	mutex sync.Mutex
	list  map[string]bool
}

// Append ...
func (collection *URLCollection) Append(url string) {
	collection.mutex.Lock()
	if _, ok := collection.list[url]; !ok {
		collection.list[url] = false
	}
	collection.mutex.Unlock()
}

// Set ...
func (collection *URLCollection) Set(url string, state bool) {
	collection.mutex.Lock()
	if _, ok := collection.list[url]; ok {
		collection.list[url] = state
	}
	collection.mutex.Unlock()
}

// getCount ...
func (collection *URLCollection) getCount() int {
	return len(collection.list)
}

// runDownload ...
func runDownload(threadURL string, savePathDir string, threadCount int) {
	urls, err := getImageUrls(threadURL)

	if err != nil {
		print(err)
	}

	total := urls.getCount()
	channel := make(chan int)
	start := time.Now()
	for i := 0; i < threadCount; i++ {
		go Downloader(urls, savePathDir, channel)
	}

	count := 0
	for i := range channel {
		count += i
		clearTerminal()
		renderProgressbar(25, total, count)
		if total == count {
			close(channel)
		}
	}

	fmt.Println("Duratin:", time.Now().Unix()-start.Unix(), "sec")
}

// Downloader ...
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

// getImageUrls
func getImageUrls(url string) (*URLCollection, error) {
	res, err := http.Get(url)
	urlCollection := URLCollection{list: make(map[string]bool)}

	if err != nil {
		return &urlCollection, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

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

// download
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

// clearTerminal ...
func clearTerminal() error {
	clear := make(map[string]func())

	clear["linux"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

	clearFunction, ok := clear[runtime.GOOS]
	if ok {
		clearFunction()
		return nil
	}

	return errors.New("your platform is unsupported! I can't clear terminal screen :(")
}

// renderProgressbar ...
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

	threadURL := flag.String("url", "", "Thread url")
	savePathDir := flag.String("path", "", "Save directory path")
	thredCount := flag.Int("thread", 1, "Thread count")
	flag.Parse()

	if *threadURL == "" {
		println("Thread url should be specified")
		return
	}

	if *savePathDir == "" {
		println("Save directory path  should be specified")
		return
	}

	runDownload(*threadURL, *savePathDir, *thredCount)
}
