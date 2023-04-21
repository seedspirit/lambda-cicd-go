package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

type extractedInfo struct {
	lineNum   string
	stationNm string
	naverCode int
}

var INFO = map[string][]map[string]interface{}{}
var wg = new(sync.WaitGroup)

func main() {
	start := time.Now()
	var batchNum = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	for _, num := range batchNum {
		fmt.Println(num, "번째 배치 돌기 시작")
		run(num)
		fmt.Println(num, "번째 배치 돌고, 15초 쉬기 시작")
		time.Sleep(time.Second * 15)

	}

	fmt.Println(INFO)
	end := time.Since(start)
	fmt.Println("총 실행 시간 : ", end)
}

func run(num int) {
	var baseURL string = "https://pts.map.naver.com/end-subway/ends/web/"

	// 100~501을 돌며 네이버 코드 스크래핑 (feat. 고루틴)
	c := make(chan extractedInfo)
	for i := (num - 1) * 1000; i < (num * 1000); i++ {
		wg.Add(1)
		go scrapeNavercode(i, baseURL, c)
		wg.Done()

	}
	wg.Wait()

	// go 루틴에서 채널로 값 받아오기 & 받은 값을 이후에 처리하기 쉬운 형태로 가공
	for i := (num - 1) * 1000; i < (num * 1000); i++ {
		result := <-c
		lineNum := result.lineNum
		block := make(map[string]interface{})

		// 만약 받아온 값이 없다면 무시 하고 아니라면 값 정리
		if result.stationNm == "" || result.lineNum == "" {
			continue
		} else {
			block["stationNm"] = result.stationNm
			block["naverCode"] = result.naverCode
		}

		_, ok := INFO[lineNum]
		if ok == false {
			INFO[lineNum] = []map[string]interface{}{}
		}
		INFO[lineNum] = append(INFO[lineNum], block)

		sort.Slice(INFO[lineNum], func(i, j int) bool {
			return INFO[lineNum][i]["naverCode"].(int) < INFO[lineNum][j]["naverCode"].(int)
		})
	}

	writeFile(INFO)

}

func scrapeNavercode(code int, baseURL string, c chan<- extractedInfo) {
	pageURL := baseURL + strconv.Itoa(code) + "/home"

	// pageURL로 접속하기
	res, err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)

	// 작업 끝나면 res.Body 닫아주는 명령 예약
	defer res.Body.Close()

	// html 읽기
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	// 지하철 정보 파싱 후 채널로 보내기
	lineNum := doc.Find(".line_no").Text()
	stationNm := doc.Find(".place_name").Text()
	fmt.Println(code, " 확인완료")

	c <- extractedInfo{
		lineNum:   lineNum,
		stationNm: stationNm,
		naverCode: code}
}

// map을 json 형태로 변환 후 "연_월_일_subwayinformation.json" 이름 형식으로 저장
func writeFile(INFO map[string][]map[string]interface{}) {
	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		panic(err)
	}
	now := time.Now()
	t := now.In(loc)
	fileTime := t.Format("2006_01_02")

	fileName := fileTime + "_subway_information.json"
	content, _ := json.MarshalIndent(INFO, "", " ")
	_ = os.WriteFile(fileName, content, 0644)
}

// 에러 체킹용 함수
func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status:", res.StatusCode)
	}
}
