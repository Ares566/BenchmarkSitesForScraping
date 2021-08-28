package benchmark

import (
	"BenchmarkSitesForScraping/internal/serp"
	"encoding/json"
	"fmt"
	cassowary "github.com/rogerwelin/cassowary/pkg/client"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Benchmark struct {
	error error
	query string
	items []serp.ResponseItem
}

func NewBenchmark(sQuery string) *Benchmark {
	return &Benchmark{query: sQuery}
}
func (bm *Benchmark) Process() {

	bm.prepareURLs()

	if bm.error != nil {
		log.Print(bm.error)
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(len(bm.items))
	for _, item := range bm.items {
		go bm.abURLs(item.Url, &wg)
	}
	wg.Wait()
}

func (bm *Benchmark) abURLs(sUrl string, wg *sync.WaitGroup) {


	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Caught:", r)
			wg.Done()
			return
		}
	}()

	//fmt.Printf("\nProcessing %s\n", sUrl)
	timeout := time.Duration(3 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(sUrl)
	if err != nil {
		log.Println("Site unreachable, error: ", err)
		wg.Done()
		return
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {

	} else {
		fmt.Println("Non-OK HTTP status:", resp.StatusCode)
		wg.Done()
		return
	}

	cass := &cassowary.Cassowary{
		BaseURL:               sUrl,
		ConcurrencyLevel:      1, // TODO подбор количества одновременных запросов
		Requests:              1,
		DisableTerminalOutput: true,
	}

	metrics, err := cass.Coordinate()
	if err != nil {
		log.Print(err)
	}

	// or print as json
	jsonMetrics, err := json.Marshal(metrics)
	if err != nil {
		log.Print(err)
	}

	fmt.Println(string(jsonMetrics))
	wg.Done()


}

func (bm *Benchmark) prepareURLs() {

	compiledUrl := fmt.Sprintf(serp.BaseYandexURL, url.QueryEscape(bm.query))
	res, err := http.Get(compiledUrl)
	if err != nil {
		bm.error = err
		return
	}

	if res.StatusCode != 200 {
		bm.error = fmt.Errorf("error code: %d %s\n", res.StatusCode, res.Status)
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		bm.error = err
		return
	}
	closeErr := res.Body.Close()
	if closeErr != nil {
		bm.error = closeErr
	}

	serpAnswer := serp.ParseYandexResponse(body)

	if serpAnswer.Error != nil {
		bm.error = serpAnswer.Error
		return
	}

	bm.items = serpAnswer.Items
}
