package scrapper

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
	company  string
	location string
	summary  string
}

func Scrape(term string) {
	baseURL := "https://www.saramin.co.kr/zf_user/search/recruit?&searchword=" + term
	mainChannel := make(chan []extractedJob)
	extractedJobs := []extractedJob{}
	totalPages := getNumberOfPages(baseURL)

	for i := 0; i < totalPages; i++ {
		go getPage(term, i+1, mainChannel)
	}

	for i := 0; i < totalPages; i++ {
		jobs := <-mainChannel
		extractedJobs = append(extractedJobs, jobs...)
	}

	writeToCSV(extractedJobs)
	fmt.Println("Done, extracted", len(extractedJobs))
}

func getPage(term string, page int, mainChannel chan<- []extractedJob) {
	jobs := []extractedJob{}
	c := make(chan extractedJob)
	pageURL := "https://www.saramin.co.kr/zf_user/search/recruit?=&searchword=" + term + "&recruitPage=" + strconv.Itoa(page)
	fmt.Println("Requesting", pageURL)
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
	title := TrimSpace(card.Find(".job_tit>a").Text())
	company := TrimSpace(card.Find(".area_corp>.corp_name>a").Text())
	location := TrimSpace(card.Find(".job_condition>span>a").Text())
	summary := TrimSpace(card.Find(".job_sector").Text())
	c <- extractedJob{id: id, title: title, company: company, location: location, summary: summary}
}

func getNumberOfPages(url string) int {
	res, err := http.Get(url)
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
	file.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(file)

	headers := []string{"Link", "Title", "Company", "Location", "Summary"}
	writeErr := writer.Write(headers)
	checkError(writeErr)

	for _, job := range jobs {
		jobWriteError := writer.Write([]string{"https://www.saramin.co.kr/zf_user/jobs/relay/view?isMypage=no&rec_idx=" + job.id, job.title, job.company, job.location, job.summary})
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

func TrimSpace(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}
