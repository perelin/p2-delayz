package main

import (
	"net/url"
	"os"
	"p2lab/p2-delayz/pkg/db"
	"p2lab/p2-delayz/pkg/models"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func init() {

	err := godotenv.Load("../../.env")
	if err != nil {
		log.Error("Error loading local .env file")
	}

	logLevel, _ := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	log.SetLevel(logLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	log.Info("starting stop scan")

}

func main() {
	c := colly.NewCollector()
	// c := colly.NewCollector(colly.Debugger(&debug.LogDebugger{}))

	if os.Getenv("DELAY") == "true" {
		c.Limit(&colly.LimitRule{
			DomainGlob:  "*bahn.de*",
			Delay:       3 * time.Second,
			RandomDelay: 5 * time.Second,
		})
	}

	c.OnHTML("div#sqResult h2 strong", func(e *colly.HTMLElement) {
		//log.Debug("from page: " + e.Text)
		queryDateTime := getTimeFromBahnURL(e.Request.URL)
		log.Println("datetime from url: " + queryDateTime.String())
	})

	c.OnHTML("table.result tr[id]", func(e *colly.HTMLElement) {

		queryDateTime := getTimeFromBahnURL(e.Request.URL)
		departureTimeSplit := strings.Split(e.ChildText("td.time"), ":")
		departureTimeSplitHour, _ := strconv.Atoi(departureTimeSplit[0])
		departureTimeSplitMinute, _ := strconv.Atoi(departureTimeSplit[1])
		loc, _ := time.LoadLocation("Europe/Berlin") // 2do refactor this in global constant or something
		departureTime := time.Date(
			queryDateTime.Year(),
			queryDateTime.Month(),
			queryDateTime.Day(),
			departureTimeSplitHour,
			departureTimeSplitMinute,
			0, 0, loc,
		)
		//log.Debug("dep time: ", departureTime)

		stop := models.DzStationSchedule{
			Station:       "HD",
			Train:         e.ChildText("td.train a"),
			Direction:     e.ChildText("td.route span.bold a"),
			TimeDeparture: departureTime,
			Platform:      e.ChildText("td.platform strong"),
			TrainURL:      e.ChildAttr("td.train a", "href"),
			SourceURL:     e.Request.URL.String(),
		}

		db.SaveStopToDB(stop)
	})

	c.OnHTML("a.arrowlinkbottom", func(e *colly.HTMLElement) {
		rawLink := e.Attr("href")
		urlEncodedLink := strings.Replace(rawLink, " ", "%20", -1)
		c.Visit(urlEncodedLink)
	})

	c.OnHTML("p.errormsg", func(e *colly.HTMLElement) {

		log.Debug("found empty list with (p.errormsg): ", e.Text)

		if e.Text == "Im angegebenen Zeitraum verkehren an dieser Haltestelle keine Züge." {

			newURL := forwardTimeInBahnURL(time.Hour*1, e.Request.URL)

			c.Visit(newURL.String())
		}

	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	now := time.Now()
	startTime := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		0,
		0,
		0,
		time.UTC,
	)
	startURL := getStationScheduleURLforTime("Heidelberg Hbf", startTime)

	c.Visit(startURL.String())

	// "Im angegebenen Zeitraum verkehren an dieser Haltestelle keine Züge."
	//c.Visit("https://reiseauskunft.bahn.de/bin/bhftafel.exe/dn?ld=41130&protocol=https:&rt=1&input=Heidelberg%20Hbf%238000156&boardType=dep&time=01:00%2B60&productsFilter=11111&&&date=07.01.19&&selectDate=&maxJourneys=&start=yes")

	// check "Aktuelle Informationen zu Ihrer Reise sind nur 4320 Minuten im Voraus möglich."
	//https://reiseauskunft.bahn.de/bin/bhftafel.exe/dn?ld=41130&protocol=https:&rt=1&input=Heidelberg%20Hbf%238000156&boardType=dep&time=01:00%2B60&productsFilter=11111&&&date=08.04.19&&selectDate=&maxJourneys=&start=yes
}

func getStationScheduleURLforTime(station string, time time.Time) *url.URL {

	urlTemplateRaw := "https://reiseauskunft.bahn.de/bin/bhftafel.exe/dn?ld=41130&protocol=https:&rt=1&input=Heidelberg%20Hbf%238000156&boardType=dep&time=01:00%2B60&productsFilter=11111&&&date=07.01.19&&selectDate=&maxJourneys=&start=yes"

	stationURL, _ := url.Parse(urlTemplateRaw)

	urlValues := stationURL.Query()

	urlValues["input"][0] = station
	urlValues["date"][0] = time.Format("02.01.06")
	urlValues["time"][0] = time.Format("15:04")

	stationURL.RawQuery = urlValues.Encode()

	return stationURL
}

func getTimeFromBahnURL(bahnURL *url.URL) time.Time {

	// get the basic components
	rawURLValues := bahnURL.Query()
	rawDate := rawURLValues["date"][0]
	rawTime := rawURLValues["time"][0]

	// sometimes there is an offset like +60, we remove it and deal with it later
	if strings.Contains(rawURLValues["time"][0], "+") {
		timeParts := strings.Split(rawURLValues["time"][0], "+")
		rawTime = timeParts[0]
	}

	// basic parsing
	datetime := rawDate + " - MEZ - " + rawTime
	parsedTime, err := time.Parse("02.01.06 - MST - 15:04", datetime)
	if err != nil {
		panic(datetime + " not parsable")
	}

	// deal with the offset if necessary
	if strings.Contains(rawURLValues["time"][0], "+") {
		//log.Debug("time sewtoff detected: " + rawURLValues["time"][0])
		timeParts := strings.Split(rawURLValues["time"][0], "+")
		addOnDuration, _ := strconv.Atoi(timeParts[1])
		parsedTime = parsedTime.Add(time.Minute * time.Duration(addOnDuration))
	}

	return parsedTime
}

func forwardTimeInBahnURL(duration time.Duration, bahnURL *url.URL) *url.URL {

	rawURLValues := bahnURL.Query()

	parsedTime := getTimeFromBahnURL(bahnURL)

	newTime := parsedTime.Add(duration)

	rawURLValues["date"][0] = newTime.Format("02.01.06")
	rawURLValues["time"][0] = newTime.Format("15:04")

	newURL := bahnURL

	newURL.RawQuery = rawURLValues.Encode()

	return newURL
}

func deprecatedColly() {

	// Find and visit all links
	// c.OnHTML("table.result", func(e *colly.HTMLElement) {
	// 	e.ForEach("tr", func(i int, e2 *colly.HTMLElement) {
	// 		if strings.Contains(e2.Attr("id"), "journeyRow") {
	// 			fmt.Println(e2.Attr("id"))

	// 			trainURL := e2.ChildAttr("td.train a", "href")
	// 			fmt.Println(trainURL)
	// 		}
	// 	})
	// })

	// c.OnHTML("table.result tr[id]", func(e *colly.HTMLElement) {
	// 	//fmt.Println(e.ChildText("td.time"))
	// 	//fmt.Println(e.ChildText("td.train a"))
	// 	fmt.Println(e.ChildText("td.time"), e.ChildText("td.train a"))
	// })

	// c.OnHTML("table.result tr td.train a", func(e *colly.HTMLElement) {
	// 	fmt.Println(e.Attr("href"))
	// })

	// c.OnHTML("p.lastParagraph", func(e *colly.HTMLElement) {
	// 	fmt.Println(e.Text)
	// })
}
