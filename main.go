package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

// array of month abbreviations
var months = [...]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

// convert month abbreviation to integer
func monthToInt(month string) int {
	for i, m := range months {
		if m == month {
			return i
		}
	}
	return 0
}

type Items []string

func (u Items) Len() int {
	return len(u)
}
func (u Items) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}
func (u Items) Less(i, j int) bool {
	if u[i][0:3] == "UNK" {
		return true
	}
	if u[j][0:3] == "UNK" {
		return false
	}
	if u[i][0:5] == "Today" {
		return false
	}
	if u[j][0:5] == "Today" {
		return true
	}
	if u[i][0:5] == "Yeste" {
		return false
	}
	if u[j][0:5] == "Yeste" {
		return true
	}
	// first 5 characters are date
	date, err := strconv.Atoi(strings.Replace(u[i][0:5], "/", "", 1))
	date2, err2 := strconv.Atoi(strings.Replace(u[j][0:5], "/", "", 1))
	if err != nil || err2 != nil {
		return false
	}

	// convert date and date2 to integers
	return date < date2
}

func main() {
	currentMonth := int(time.Now().Month())
	urls := [3]string{"https://www.toyota-4runner.org/for-sale-t4r-items/",
		"https://www.toyota-4runner.org/free/",
		"https://www.4runners.com/forums/5th-gen-4runner-parts-marketplace-2010-2024.8/"}

	names := [3]string{"t4r.org for sale",
		"t4r.org free",
		"4runners.com"}

	terms := []string{"rock", "rail", "slider", "skid", "skidplate", "valence", "valance", "parts", "takeoff", "take off"}
	ignore := []string{"3rd", "4th"}

	var items []string
	for i := 0; i < len(urls); i++ {
		client := &http.Client{}
		req, err := http.NewRequest("GET", urls[i], nil)
		if err != nil {
			fmt.Print(err.Error())
		}
		resp, err := client.Do(req)
		defer resp.Body.Close()
		if err != nil {
			fmt.Print(err.Error())
		}
		if resp.StatusCode == 404 {
			fmt.Println("Page not found at url: " + urls[i])
			os.Exit(1)
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		responseString := buf.String()

		w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
		var responseLines []string = strings.Split(responseString, "\n")
		for j := 0; j < len(responseLines); j++ {
			line := strings.ToLower(responseLines[j])
			firstIndex := 0
			secondIndex := 0
			date := "UNK"
			site := "UNK"
			// if line includes substring
			if strings.Contains(line, "thread_title") {
				site = "t4r"
				// get first index of >
				firstIndex = strings.Index(line, ">")

				// get second index of <
				secondIndex = strings.Index(line, "</a>")
				nextLine := ""
				if secondIndex == -1 {
					nextLine = strings.TrimSpace(responseLines[j+1])
					secondIndex = strings.Index(nextLine, "<")
					line = line[firstIndex+1:] + " " + nextLine[0:secondIndex]
				} else {
					line = line[firstIndex+1 : secondIndex]
				}

				// get date
				date = responseLines[j+21]
				date = strings.TrimSpace(date)
				// if date is less than 10 characters
				if len(date) >= 5 {
					date = strings.Replace(date[0:5], "-", "/", 1)
				} else {
					date = "UNK"
				}

				if strings.Contains(date, "div") {
					date = "UNK"
				}
			} else if strings.Contains(line, "/preview") {
				site = "4rs"
				// get first index of >
				firstIndex = strings.Index(line, ">")

				// get second index of <
				secondIndex = strings.Index(line, "<")
				nextLine := ""
				if secondIndex == -1 {
					nextLine = strings.TrimSpace(responseLines[j+1])
					secondIndex = strings.Index(nextLine, "<")
					line = line[firstIndex+1:] + " " + nextLine[0:secondIndex]
				} else {
					line = line[firstIndex+1 : secondIndex]
				}

				// get date
				date = responseLines[j+7]

				dateFirstIndex := strings.Index(date, "title") + 9
				month := monthToInt(date[dateFirstIndex : dateFirstIndex+3])
				day := strings.Replace(date[dateFirstIndex+4:dateFirstIndex+6], ",", "", 1)
				dayInt, err := strconv.Atoi(day)
				if dayInt < 10 && err == nil {
					day = "0" + day
				}

				if month < 10 {
					date = "0" + strconv.Itoa(month) + "/" + day
				} else {
					date = strconv.Itoa(month+1) + "/" + day
				}
			}

			if firstIndex != 0 && secondIndex != 0 {
				line = strings.Replace(line, "&quot;", "\"", 1)

				cost := "For sale"
				if strings.Contains(names[i], "free") {
					cost = "Free\t"
				} else if strings.Contains(line, "$") {
					index := strings.Index(line, "$")
					afterString := line[index:]
					if strings.Contains(afterString, " ") {
						afterString = afterString[:strings.Index(afterString, " ")]
					}
					cost = afterString + "\t"
				}

				keepLooking := true
				for k := 0; k < len(terms); k++ {
					if keepLooking && strings.Contains(line, terms[k]) {
						ignoreItem := false
						for l := 0; l < len(ignore); l++ {
							if !ignoreItem && strings.Contains(line, ignore[l]) {
								ignoreItem = true
							}
						}
						if !ignoreItem {
							month, err := strconv.Atoi(date[0:2])
							if err == nil && month > currentMonth {
								continue
							}

							if strings.Contains(date, "00") {
								date = "UNK"
							}

							if len(date) <= 5 {
								date = date + "\t"
							}

							line = date + "\t" + site + "\t" + cost + "\t" + line
							items = append(items, line)
							keepLooking = false
						}
					}
				}
			}
		}
		w.Flush()
	}
	fmt.Println("\nResults ~\n")
	// items = Items(items)
	sort.Stable(Items(items))
	for i := 0; i < len(items); i++ {
		fmt.Println(items[i])
	}
	return
}
