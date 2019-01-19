package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"time"
)

const _terms = "data/terms.json"
const _holidays = "data/holidays.json"
const _keyDates = "data/key-dates.json"
const _port = 8080

// Term is a struct representing a school term
type Term struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// KeyDate is a generic school important date or public holiday
type KeyDate struct {
	Name string `json:"name"`
	Date string `json:"date"`
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

func (holiday KeyDate) getDate() time.Time {
	date := parseDate(holiday.Date, "")
	return date
}

func getHolidays(start, end time.Time, year string) []KeyDate {
	names := make([]KeyDate, 0)

	for _, holiday := range holidays {
		date := holiday.getDate()
		if start.Unix() < date.Unix() && date.Unix() < end.Unix() {
			names = append(names, holiday)
		}
	}
	return names
}

func getKeyDates(start, end time.Time, year string) []KeyDate {
	names := make([]KeyDate, 0)

	for _, keyDate := range keyDates {
		date := keyDate.getDate()
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
		//"TermHolidays":      termHolidays,
		//"KeyDates":          getKeyDates(term.getDates(year)["start"], term.getDates(year)["end"], year),
		"Weekends": schoolWeeks + 1,
	}
}

func sortEvents(events []KeyDate) []KeyDate {

	// keys := make([]int64, 0, len(events))
	// for _, event := range events {
	// 	keys = append(keys, time.Time(parseDate(event.Date, "")).Unix())
	// }

	sort.Slice(events, func(i, j int) bool {
		date1 := parseDate(events[i].Date, "").Unix()
		date2 := parseDate(events[j].Date, "").Unix()
		return date1 < date2
	})

	return events[func(events []KeyDate) int {
		for i := range events {
			if parseDate(events[i].Date, "").Unix() > time.Now().Unix() {
				return i
			}
		}
		return len(events)
	}(events):]
}

func getEvents() []KeyDate {

	events := make([]KeyDate, 0)
	events = append(events, holidays...)
	events = append(events, keyDates...)

	for year := range terms {
		for i, term := range terms[year] {
			events = append(events, KeyDate{fmt.Sprintf("Term %d Starts", i+1), term.Start + " " + year})
			events = append(events, KeyDate{fmt.Sprintf("Term %d Ends", i+1), term.End + " " + year})
		}
	}

	return sortEvents(events)
}

func getTerms(year string) []map[string]interface{} {

	ret := make([]map[string]interface{}, 0)

	for _, term := range terms[year] {
		ret = append(ret, calcDates(term, year))
	}

	return ret
}

func school(year string) map[string]interface{} {

	summary := make(map[string]interface{}, 0)

	summary["SchoolDaysTotal"] = int(0)
	summary["SchoolDaysRemaining"] = int(0)

	for i, term := range terms[year] {

		this := calcDates(term, year)

		if parseDate(term.Start, year).Unix() < time.Now().Unix() && parseDate(term.End, year).Unix() > time.Now().Unix() {
			_, month, day := time.Now().Date()
			start := fmt.Sprintf("%d %s", day, month)
			this["DaysRemaining"] = calcDates(Term{start, term.End}, year)["SchoolDays"]
			summary["SchoolDaysRemaining"] = summary["SchoolDaysRemaining"].(int) + this["DaysRemaining"].(int)
			summary["CurrentTerm"] = i + 1
			summary["Term"] = this
		} else {
			summary["SchoolDaysRemaining"] = summary["SchoolDaysRemaining"].(int) + this["SchoolDays"].(int)
		}

		summary["SchoolDaysTotal"] = summary["SchoolDaysTotal"].(int) + this["SchoolDays"].(int)
	}

	if summary["CurrentTerm"] == nil {
		summary["CurrentTerm"] = "Holidays"
	}

	//summary["Events"] = sortEvents(events)

	return summary

}

func main() {

	terms = make(map[string][]Term, 0)
	holidays = make([]KeyDate, 0)
	keyDates = make([]KeyDate, 0)

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

	load(_terms, &terms)
	load(_holidays, &holidays)
	load(_keyDates, &keyDates)

	port := os.Getenv("PORT")
	if port == "" {
		port = strconv.Itoa(_port)
	}

	fmt.Printf("Starting server on port %s...\n", port)
	p, _ := strconv.Atoi(port)
	router(p)
}
