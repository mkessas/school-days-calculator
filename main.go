package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

type Term struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type KeyDate struct {
	Name     string `json:"name"`
	Date     string `json:"date"`
	Division string `json:"division"`
}

var terms map[string][]Term
var holidays, keyDates []KeyDate

func parseDate(date string, year string) time.Time {
	ret, err := time.Parse("2 January 2006 MST", date+" "+year+" NZT")
	if err != nil {
		panic(err.Error())
	}
	return ret
}

func (term Term) getDates(year string) map[string]time.Time {
	start := parseDate(term.Start, year)
	end := parseDate(term.End, year)
	return map[string]time.Time{"start": start, "end": end}
}

func (holiday KeyDate) getDate(year string) time.Time {
	date := parseDate(holiday.Date, year)
	return date
}

func getHolidays(start, end time.Time, year string) []KeyDate {
	names := make([]KeyDate, 0)

	for _, holiday := range holidays {
		date := holiday.getDate(year)
		if start.Unix() < date.Unix() && date.Unix() < end.Unix() {
			names = append(names, holiday)
		}
	}
	return names
}

func getKeyDates(start, end time.Time, year string) []KeyDate {
	names := make([]KeyDate, 0)

	for _, keyDate := range keyDates {
		date := keyDate.getDate(year)
		if start.Unix() < date.Unix() && date.Unix() < end.Unix() {
			names = append(names, keyDate)
		}
	}
	return names
}

func calcDates(term Term, year string) map[string]interface{} {

	firstWeekDays := 8 - int(term.getDates(year)["start"].Weekday())
	lastWeekDays := int(term.getDates(year)["end"].Weekday())
	totalDays := int(term.getDates(year)["end"].Sub(term.getDates(year)["start"]).Hours()/24) + 1
	schoolWeeks := (totalDays - firstWeekDays - lastWeekDays) / 7
	termHolidays := getHolidays(term.getDates(year)["start"], term.getDates(year)["end"], year)

	return map[string]interface{}{
		"SchoolYear":        year,
		"StartDate":         term.Start,
		"EndDate":           term.End,
		"TotalCalendarDays": totalDays,
		"SchoolWeeks":       schoolWeeks,
		"SchoolDays":        schoolWeeks*5 + (firstWeekDays - 2) + lastWeekDays - len(termHolidays),
		"TermHolidays":      termHolidays,
		"KeyDates":          getKeyDates(term.getDates(year)["start"], term.getDates(year)["end"], year),
		"Weekends":          schoolWeeks + 1,
	}
}

func main() {
	terms = make(map[string][]Term, 0)
	holidays = make([]KeyDate, 0)
	keyDates = make([]KeyDate, 0)
	summary := make(map[string]interface{}, 0)
	Year := strconv.Itoa(time.Now().Year())

	if len(os.Args) > 1 {
		Year = os.Args[1]
	}

	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("ERROR: %s\n", e)
		}
	}()

	load := func(filename string, data interface{}) {
		raw, _ := ioutil.ReadFile(filename)
		if err := json.Unmarshal(raw, &data); err != nil {
			panic(err)
		}
	}

	load("terms.json", &terms)
	load("holidays.json", &holidays)
	load("key-dates.json", &keyDates)

	for year := range terms {

		if year != Year {
			continue
		}

		summary["SchoolDaysTotal"] = int(0)
		summary["Terms"] = make([]map[string]interface{}, 0)

		for _, term := range terms[year] {

			if parseDate(term.Start, year).Unix() < time.Now().Unix() {
				continue
			}

			this := calcDates(term, year)

			summary["Terms"] = append(summary["Terms"].([]map[string]interface{}), this)
			summary["SchoolDaysTotal"] = summary["SchoolDaysTotal"].(int) + this["SchoolDays"].(int)
		}
	}

	if encoded, err := json.MarshalIndent(summary, "", "    "); err == nil {
		fmt.Printf("%s\n", encoded)
	} else {
		panic(err.Error())
	}

}
