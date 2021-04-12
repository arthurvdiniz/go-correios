package gateway

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/djimenez/iconv-go"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/arthurvdiniz/go-correios/entity"
)

type CorreiosScraperGateway struct {
	TrackerURL string
}

func (g *CorreiosScraperGateway) GetTrackerCodeContent(box *entity.Box) error {
	document, err := g.getParcelDocument(box.Code)
	if err != nil {
		return err
	}

	err = g.getParcelContent(box, document)
	if err != nil {
		return err
	}

	return nil

}

func (g *CorreiosScraperGateway) getParcelDocument(code string) (*goquery.Selection, error) {
	body := strings.NewReader(url.Values{"acao": {"track"}, "objetos": {code}, "btnPesq": {"Buscar"}}.Encode())
	req, err := http.NewRequest("POST", g.TrackerURL, body)
	if err != nil {
		log.Printf("Error creating request: %s", err.Error())
		return nil, err
	}

	req.Header.Add("Referer", g.TrackerURL)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := new(http.Client).Do(req)
	if err != nil {
		log.Printf("Error requesting: %s", err.Error())
		return nil, err
	}

	defer res.Body.Close()

	utfBody, err := iconv.NewReader(res.Body, "ISO-8859-1", "utf-8")
	if err != nil {
		log.Printf("Cannot convert to UTF-8: %s", err.Error())
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(utfBody)

	if err != nil {
		log.Printf("Cannot parse document: %s", err.Error())
		return nil, err
	}

	return doc.Find(".ctrlcontent").First(), nil
}

func (g *CorreiosScraperGateway) getParcelContent(box *entity.Box, selection *goquery.Selection) error {
	postDateRegex := regexp.MustCompile("\\d{2}\\/\\d{2}\\/\\d{4}")
	postDateArr := postDateRegex.FindStringSubmatch(selection.Find("#EventoPostagem").Text())
	if len(postDateArr) == 0 {
		return errors.New("Parcel Not Found")
	}

	postDate, err := time.Parse("02/01/2006", postDateArr[0])
	if err != nil {
		return errors.New(fmt.Sprintf("Error parsing package post date: %s", err.Error()))
	}

	r, _ := regexp.Compile("\\s+")
	datetimeRegex := regexp.MustCompile("\\d{2}\\/\\d{2}\\/\\d{4}\\s\\d{2}:\\d{2}")
	locationRegex := regexp.MustCompile("[A-z].+\\s{0,}\\/{0,}[A-Z]+")

	var events []entity.Event
	selection.Find(".listEvent tr").Each(func(i int, s *goquery.Selection) {
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

		events = append(events, entity.Event{
			Date:     date.UTC().String(),
			Location: location,
			Info:     textContent,
		})
	})

	box.PostDate = postDate.UTC().String()
	box.Events = events

	return nil
}