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
)

type Items []string

func (u Items) Len() int {
	return len(u)
}
func (u Items) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}
func (u Items) Less(i, j int) bool {
	if u[i][0:5] == "Yeste" {
		return true
	}
	if u[j][0:5] == "Yeste" {
		return false
	}
	// first 5 characters are date
	date, err := strconv.Atoi(strings.Replace(u[i][0:5], "/", "", 1))
	date2, err2 := strconv.Atoi(strings.Replace(u[j][0:5], "/", "", 1))
	if err != nil || err2 != nil {
		// fmt.Println("Error in item sort\nItem 1: " + u[i] + "\nItem 2: " + u[j])
		return false
	}

	// convert date and date2 to integers
	return date > date2
}

func main() {
	urls := [3]string{"https://www.toyota-4runner.org/for-sale-t4r-items/",
		"https://www.toyota-4runner.org/free/",
		"https://www.4runners.com/forums/5th-gen-4runner-parts-marketplace-2010-2024.8/"}

	names := [3]string{"t4r.org for sale",
		"t4r.org free",
		"4runners.com"}

	terms := []string{"oem", "rock rails", "oem sliders", "skid plate", "skidplate", "tire hitch", "tire mount", "front valence", "bumper valence", "orp valence", "road valence"}

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
			date := "     "
			// if line includes substring
			if strings.Contains(line, "thread_title") {
				// get first index of >
				firstIndex = strings.Index(line, ">")

				// get second index of <
				secondIndex = strings.Index(line, "</a>")

				// get date
				date = responseLines[j+21]
				date = strings.TrimSpace(date)
				date = strings.Replace(date[0:5], "-", "/", 1)
				fmt.Println(date)
			} else if strings.Contains(line, "/preview") {
				// get first index of >
				firstIndex = strings.Index(line, ">")

				// get second index of <
				secondIndex = strings.Index(line, "<")
			}

			if firstIndex != 0 && secondIndex != 0 {
				line = line[firstIndex+1 : secondIndex]
				line = strings.Replace(line, "&quot;", "\"", 1)

				cost := "For sale:"
				if strings.Contains(names[i], "free") {
					cost = "Free:\t"
				} else if strings.Contains(line, "$") {
					index := strings.Index(line, "$")
					afterString := line[index:]
					if strings.Contains(afterString, " ") {
						afterString = afterString[:strings.Index(afterString, " ")]
					}
					cost = afterString + "\t"
				}

				if strings.Contains(date, "Yeste") {
					date = "Yday"
				}
				if strings.Contains(date, "Today") {
					date = "Today"
				}

				line = date + "\t" + cost + "\t" + line

				for k := 0; k < len(terms); k++ {
					if strings.Contains(line, terms[k]) {
						items = append(items, line)
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
