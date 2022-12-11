package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
)

func main() {
	urls := [3]string{"https://www.toyota-4runner.org/for-sale-t4r-items/",
		"https://www.toyota-4runner.org/free/",
		"https://www.4runners.com/forums/5th-gen-4runner-parts-marketplace-2010-2024.8/"}

	names := [3]string{"t4r.org for sale",
		"t4r.org free",
		"4runners.com"}

	terms := []string{"rock rails", "oem sliders", "skid plate", "skidplate", "tire hitch", "tire mount", "front valence", "bumper valence", "orp valence", "road valence"}

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
			// if line includes substring
			if strings.Contains(line, "thread_title") {
				// get first index of >
				firstIndex = strings.Index(line, ">")

				// get second index of <
				secondIndex = strings.Index(line, "</a>")
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

				line = cost + "\t" + line

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
	for i := 0; i < len(items); i++ {
		fmt.Println(items[i])
	}
	return
}
