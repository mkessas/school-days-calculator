package main

import (
	"fmt"
	"sort"
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
var timezone *time.Location

func parseDate(date string, year string) time.Time {
	format := "2 January 2006"
	str := date

	if year != "" {
		str = str + " " + year
	}

	ret, err := time.Parse(format, str)
	if err != nil {
		panic(err.Error())
	}
	return ret.Local()
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
		date := holiday.getDate().Add(time.Hour * 24)
		if start.Before(date) && date.Before(end) {
			names = append(names, holiday)
		}
	}
	return names
}

func getKeyDates(start, end time.Time, year string) []KeyDate {
	names := make([]KeyDate, 0)

	for _, keyDate := range keyDates {
		date := keyDate.getDate()
		if start.Before(date) && date.Before(end) {
			names = append(names, keyDate)
		}
	}
	return names
}

func getWeekday(date time.Time) int {
	val := int(date.Weekday())
	if val == 0 {
		return 6
	}
	return val - 1
}

func calcDates(term Term, year string) map[string]interface{} {

	firstWeekDays := 6 - getWeekday(term.getDates(year)["start"]) - 1
	if firstWeekDays < 0 {
		firstWeekDays = 0
	}
	lastWeekDays := getWeekday(term.getDates(year)["end"]) + 1
	totalDays := int(term.getDates(year)["end"].Sub(term.getDates(year)["start"]).Hours()/24) + 1
	schoolWeeks := (totalDays - firstWeekDays - lastWeekDays) / 7
	termHolidays := getHolidays(term.getDates(year)["start"], term.getDates(year)["end"], year)

	// fmt.Printf("%s -> %s\n%s %d\n%s %d\n%s %d\n%s %d\n%s %d\n%s %d\n\n",
	// 	term.Start, term.End,
	// 	"firstWeekDays", firstWeekDays,
	// 	"totalDays", totalDays,
	// 	"schoolWeeks", schoolWeeks,
	// 	"lastWeekDays", lastWeekDays,
	// 	"termHolidays", len(termHolidays),
	// 	"schoolDays", (schoolWeeks*5 + firstWeekDays + lastWeekDays - len(termHolidays)),
	// )

	return map[string]interface{}{
		"SchoolYear":        year,
		"StartDate":         term.Start,
		"EndDate":           term.End,
		"TotalCalendarDays": totalDays,
		"SchoolWeeks":       schoolWeeks,
		"SchoolDays":        schoolWeeks*5 + firstWeekDays + lastWeekDays - len(termHolidays),
		//"TermHolidays":      termHolidays,
		//"KeyDates":          getKeyDates(term.getDates(year)["start"], term.getDates(year)["end"], year),
		"Weekends": schoolWeeks + 1,
	}
}

func sortEvents(events []KeyDate) []KeyDate {

	sort.Slice(events, func(i, j int) bool {
		date1 := parseDate(events[i].Date, "")
		date2 := parseDate(events[j].Date, "")
		return date1.Before(date2)
	})

	return events[func(events []KeyDate) int {
		for i := range events {
			if parseDate(events[i].Date, "").Add(time.Hour * 24).After(time.Now()) {
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

func getSummary(year string) map[string]interface{} {

	summary := make(map[string]interface{}, 0)

	summary["SchoolDaysTotal"] = int(0)
	summary["SchoolDaysRemaining"] = int(0)

	for i, term := range terms[year] {

		this := calcDates(term, year)

		if parseDate(term.Start, year).Before(time.Now()) && parseDate(term.End, year).After(time.Now()) {
			_, month, day := time.Now().Date()
			today := fmt.Sprintf("%d %s", day, month)
			this["DaysRemaining"] = calcDates(Term{today, term.End}, year)["SchoolDays"].(int)
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
