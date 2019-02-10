package main

import (
	"fmt"
	"net/url"
	"testing"
	"time"
)

func TestForwardTimeInBahnURLWithoutOffset(t *testing.T) {

	initialURL, _ := url.Parse("https://reiseauskunft.bahn.de/bin/bhftafel.exe/dn?ld=41130&protocol=https:&rt=1&input=Heidelberg%20Hbf%238000156&boardType=dep&time=01:00&productsFilter=11111&&&date=07.01.19&&selectDate=&maxJourneys=&start=yes")

	resultingURL, _ := url.Parse("https://reiseauskunft.bahn.de/bin/bhftafel.exe/dn?boardType=dep&date=07.01.19&input=Heidelberg+Hbf%238000156&ld=41130&maxJourneys=&productsFilter=11111&protocol=https%3A&rt=1&selectDate=&start=yes&time=02%3A00")

	addonDuration := time.Hour * 1

	newURL := forwardTimeInBahnURL(addonDuration, initialURL)

	if newURL.String() != resultingURL.String() {
		fmt.Println(newURL.String())
		t.Fail()
	}
}

func TestForwardTimeInBahnURLWithOffset(t *testing.T) {

	initialURL, _ := url.Parse("https://reiseauskunft.bahn.de/bin/bhftafel.exe/dn?ld=41130&protocol=https:&rt=1&input=Heidelberg%20Hbf%238000156&boardType=dep&time=01:00%2B60&productsFilter=11111&&&date=07.01.19&&selectDate=&maxJourneys=&start=yes")

	resultingURL, _ := url.Parse("https://reiseauskunft.bahn.de/bin/bhftafel.exe/dn?boardType=dep&date=07.01.19&input=Heidelberg+Hbf%238000156&ld=41130&maxJourneys=&productsFilter=11111&protocol=https%3A&rt=1&selectDate=&start=yes&time=03%3A00")

	addonDuration := time.Hour * 1

	newURL := forwardTimeInBahnURL(addonDuration, initialURL)

	if newURL.String() != resultingURL.String() {
		//fmt.Println(newURL.String())
		t.Fail()
	}
}

func TestGetTimeFromBahnURL(t *testing.T) {

	bahnURL, _ := url.Parse("https://reiseauskunft.bahn.de/bin/bhftafel.exe/dn?ld=41130&protocol=https:&rt=1&input=Heidelberg%20Hbf%238000156&boardType=dep&time=01:00%2B60&productsFilter=11111&&&date=07.01.19&&selectDate=&maxJourneys=&start=yes")

	extractedTime := getTimeFromBahnURL(bahnURL)

	if extractedTime.String() != "2019-01-07 02:00:00 +0000 MEZ" {
		fmt.Println(extractedTime.String())
		t.Fail()
	}
}
