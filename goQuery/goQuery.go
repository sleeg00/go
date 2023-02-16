package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var baseURL string = "https://www.saramin.co.kr/zf_user/search/recruit?&searchword=python"

type extracteJob struct {
	value string
	title string
}

func main() {
	var jobs []extracteJob
	c := make(chan []extracteJob)

	totalPages := getPages()
	fmt.Println(totalPages)

	for i := 0; i < totalPages; i++ {
		go getPage(i, c)
	}

	for i := 0; i < totalPages; i++ {
		extractJob := <-c
		jobs = append(jobs, extractJob...)
	}
	fmt.Println(jobs)
}

func getPage(page int, mainC chan<- []extracteJob) {

	var jobs []extracteJob
	c := make(chan extracteJob)
	pageUrl := baseURL + "&start=" + strconv.Itoa(page*50)
	//fmt.Println("Requesting", pageUrl)
	res, err := http.Get(pageUrl)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)

	checkErr(err)
	searchCards := doc.Find(".item_recruit")
	searchCards.Each(func(i int, s *goquery.Selection) {
		go extractJob(s, c)
	})
	for i := 0; i < 5; i++ {
		job := <-c
		jobs = append(jobs, job)
	}
	mainC <- jobs
}

func extractJob(card *goquery.Selection, c chan<- extracteJob) {
	id, _ := card.Find(".area_job").Find(".job_tit>a").Attr("title")
	title := cleanString(card.Find(".area_job").Text())

	c <- extracteJob{
		title: id,
		value: title,
	}
}

func getPages() int {

	pages := 0

	res, err := http.Get(baseURL)

	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)

	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})

	return pages
}

func cleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status : ", res.StatusCode)
	}

}
