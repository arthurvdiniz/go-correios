package gocorreios

import (
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/djimenez/iconv-go"
)

func getPackageDocument(code string) (*goquery.Selection, error) {
	trackerURL := "https://www2.correios.com.br/sistemas/rastreamento/resultado.cfm"

	body := strings.NewReader(url.Values{"acao": {"track"}, "objetos": {code}, "btnPesq": {"Buscar"}}.Encode())
	req, err := http.NewRequest("POST", trackerURL, body)
	if err != nil {
		log.Fatalf("Error creating request: %s", err.Error())
		return nil, err
	}

	req.Header.Add("Referer", trackerURL)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := new(http.Client).Do(req)
	if err != nil {
		log.Fatalf("Error requesting: %s", err.Error())
		return nil, err
	}

	defer res.Body.Close()

	utfBody, err := iconv.NewReader(res.Body, "ISO-8859-1", "utf-8")
	if err != nil {
		log.Fatalf("Cannot convert to UTF-8: %s", err.Error())
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(utfBody)

	if err != nil {
		log.Fatalf("Cannot parse document: %s", err.Error())
		return nil, err
	}

	return doc.Find(".ctrlcontent").First(), nil
}

func getPackageContent(content *goquery.Selection) (postDate time.Time, events []Event) {
	postDateRegex := regexp.MustCompile("\\d{2}\\/\\d{2}\\/\\d{4}")
	postDateText := postDateRegex.FindStringSubmatch(content.Find("#EventoPostagem").Text())[0]
	postDate, err := time.Parse("02/01/2006", postDateText)
	if err != nil {
		log.Printf("Error parsing package post date: %s", err.Error())
	}

	r, _ := regexp.Compile("\\s+")
	datetimeRegex := regexp.MustCompile("\\d{2}\\/\\d{2}\\/\\d{4}\\s\\d{2}:\\d{2}")
	locationRegex := regexp.MustCompile("[A-z].+\\s{0,}\\/{0,}[A-Z]+")

	content.Find(".listEvent tr").Each(func(i int, s *goquery.Selection) {
		detailsContent := s.Find(".sroDtEvent").Text()
		textContent := strings.TrimSpace(s.Find(".sroLbEvent").Text())
		detailsContent = r.ReplaceAllString(detailsContent, " ")
		textContent = r.ReplaceAllString(textContent, " ")

		datetimeText := datetimeRegex.FindStringSubmatch(detailsContent)[0]
		date, err := time.Parse("02/01/2006 15:04", datetimeText)
		if err != nil {
			log.Printf("Error parsing package event date: %s", err.Error())
		}

		location := locationRegex.FindStringSubmatch(detailsContent)[0]

		events = append(events, Event{
			Date:     date.UTC().String(),
			Location: location,
			Info:     textContent,
		})
	})
	return
}

func GetTrackerCodeInformation(code string) (*Box, error) {
	document, err := getPackageDocument(code)
	if err != nil {
		return nil, err
	}

	postDate, events := getPackageContent(document)

	box := &Box{
		Code:     code,
		PostDate: postDate.UTC().String(),
		Events:   events,
	}

	return box, nil
}
