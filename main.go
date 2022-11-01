package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct {
	id       string
	title    string
	location string
	summary  string
}

var baseURL string = "https://www.saramin.co.kr/zf_user/search/recruit?&searchword=python"

func main() {
	mainChannel := make(chan []extractedJob)
	extractedJobs := []extractedJob{}
	totalPages := getNumberOfPages()

	for i := 0; i < totalPages; i++ {
		go getPage(i+1, mainChannel)
	}

	for i := 0; i < totalPages; i++ {
		jobs := <-mainChannel
		extractedJobs = append(extractedJobs, jobs...)
	}

	writeToCSV(extractedJobs)
	fmt.Println("Done, extracted", len(extractedJobs))
}

func getPage(page int, mainChannel chan<- []extractedJob) {
	jobs := []extractedJob{}
	c := make(chan extractedJob)
	pageURL := "https://www.saramin.co.kr/zf_user/search/recruit?=&searchword=python&recruitPage=" + strconv.Itoa(page)
	fmt.Println("Requesting page", page)
	res, err := http.Get(pageURL)
	checkError(err)
	checkStatusCode(res)

	defer res.Body.Close()
	doc, _ := goquery.NewDocumentFromReader(res.Body)
	foundCards := doc.Find(".item_recruit")

	foundCards.Each(func(i int, card *goquery.Selection) {
		go extractJob(card, c)
	})

	for i := 0; i < foundCards.Length(); i++ {
		job := <-c
		jobs = append(jobs, job)
	}

	mainChannel <- jobs
}

func extractJob(card *goquery.Selection, c chan<- extractedJob) {
	id, _ := card.Attr("value")
	title := removeSpace(card.Find(".job_tit>a").Text())
	location := removeSpace(card.Find(".job_condition>span>a").Text())
	summary := removeSpace(card.Find(".job_sector").Text())
	c <- extractedJob{id: id, title: title, location: location, summary: summary}
}

func getNumberOfPages() int {
	res, err := http.Get(baseURL)
	checkError(err)
	checkStatusCode(res)

	defer res.Body.Close()
	doc, _ := goquery.NewDocumentFromReader(res.Body)

	pages := 0
	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})

	return pages
}

func writeToCSV(jobs []extractedJob) {
	file, err := os.Create("results.csv")
	checkError(err)
	writer := csv.NewWriter(file)

	headers := []string{"Link", "Title", "Location", "Summary"}
	writeErr := writer.Write(headers)
	checkError(writeErr)

	for _, job := range jobs {
		jobWriteError := writer.Write([]string{"https://www.saramin.co.kr/zf_user/jobs/relay/view?isMypage=no&rec_idx=" + job.id, job.title, job.location, job.summary})
		checkError(jobWriteError)
	}
	writer.Flush()
}

func checkError(e error) {
	if e != nil {
		log.Fatalln(e)
	}
}

func checkStatusCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request Failed:", res.StatusCode)
	}
}

func removeSpace(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}
