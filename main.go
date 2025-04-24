package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

type WebScraper struct {
	// Add fields and methods for the web scraper
	URl    string
	Output string
}

type Car struct {
	Description string
	Price       string
	Link        string
}

func Save(cars *[]Car) error {

	data, err := json.MarshalIndent(&cars, "", "  ")

	if err != nil {
		fmt.Println("Error marshalling data:", err)
		return err
	}
	err = os.WriteFile("cars.json", data, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}
	fmt.Println("Data saved to output.json")
	return nil
}
func SaveCSV(cars *[]Car) error {

	file, err := os.Create("cars.csv")

	if err != nil {
		fmt.Println("Error marshalling data:", err)
		return err
	}
	defer file.Close()

	// Write CSV header
	_, err = file.WriteString("Description,Price,Link\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}
	// Write CSV data
	for _, car := range *cars {
		_, err = file.WriteString(fmt.Sprintf("%s\t%s\t%s\n", car.Description, car.Price, car.Link))
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return err
		}
	}
	fmt.Println("Data saved to output.csv")

	return nil
}

func (ws *WebScraper) Scrape() (result *http.Response, err error) {
	// Implement the scraping logic here
	fmt.Println("Scraping URL:", ws.URl)

	client := &http.Client{}

	req, err := http.NewRequest("GET", ws.URl, nil)

	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}
	// Set headers to mimic a browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error fetching URL:", err)
		return nil, err
	}

	return resp, nil

}

func (ws *WebScraper) ParseHTML(resp *http.Response) (*html.Node, error) {
	// Parse the HTML response
	defer resp.Body.Close()
	doc, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return nil, err
	}
	// Example: Print the document's root node
	return doc, nil
}

// func (ws *WebScraper) Save() {
// 	os.WriteFile(ws.Output, []byte("Scraped data"), 0644)
// 	fmt.Println("Data saved to", ws.Output)
// }

func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func hasClass(n *html.Node, class string) bool {

	for _, attr := range n.Attr {
		// if attr.Key == "class" {
		// 	fmt.Println("Class found:", attr.Val)
		// }

		if attr.Key == "class" && attr.Val == class {
			return true
		}
	}
	return false
}

// func displayClass(n *html.Node) bool {

// 	for _, attr := range n.Attr {
// 		if attr.Key == "class" {
// 			fmt.Println("Class found:", attr.Val)
// 		}

// 	}
// 	return false
// }

func CleanText(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func ProcessCars(n *html.Node, cars *[]Car) {
	if n.Type == html.ElementNode && n.Data == "div" && hasClass(n, "vehicle-details") {
		fmt.Println("🚗 Found vehicle card")
		car := Car{}
		// Find <a> with car link and title
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "a" {
				href := getAttr(c, "href") // ← Use getAttr here!
				fmt.Println("🔗 Car link:", href)
				car.Link = CleanText(href)
				for cc := c.FirstChild; cc != nil; cc = cc.NextSibling {
					if cc.Type == html.ElementNode && cc.Data == "h2" {
						if cc.FirstChild != nil {
							fmt.Println("📝 Car Title:", cc.FirstChild.Data)
							car.Description = CleanText(cc.FirstChild.Data)
						}
					}
				}
			}

			// Look for price inside price container
			if hasClass(c, "price-mileage-container") {
				// Dive into nested divs
				for cc := c.FirstChild; cc != nil; cc = cc.NextSibling {
					if cc.Type == html.ElementNode && cc.Data == "div" {
						// Now look for <span class="primary-price">
						for span := cc.FirstChild; span != nil; span = span.NextSibling {
							if span.Type == html.ElementNode && span.Data == "span" && hasClass(span, "primary-price") {
								if span.FirstChild != nil && span.FirstChild.Type == html.TextNode {
									fmt.Println("💰 Price:", span.FirstChild.Data)
									car.Price = CleanText(span.FirstChild.Data)
								}
							}
						}
					}
				}
			}
		}
		*cars = append(*cars, car)
	}

	// Recursively process all children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		ProcessCars(c, cars)
	}
}

// /html/body/main/div/div[2]/div[4]/div[2]/div[1]/div[5]
func main() {
	fmt.Println("Basic Web Scraper")
	ws := &WebScraper{
		URl:    "https://www.cars.com/shopping/results/?stock_type=new&makes%5B%5D=lexus&models%5B%5D=lexus-is_350&maximum_distance=30&zip=33178",
		Output: "output.txt",
	}

	resp, err := ws.Scrape()

	if err != nil {
		fmt.Println("Error during scraping:", err)
		return
	}

	// Parse the HTML response
	doc, err := ws.ParseHTML(resp)
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
	}
	// Example: Print the document's root node
	if doc != nil {
		fmt.Println("Parsed HTML document:", doc)
	} else {
		fmt.Println("No document to parse.")
	}

	fmt.Println("Response Status:", resp.Status)
	// ws.Save()
	// fmt.Print("Scraping data...")

	cars := []Car{}

	ProcessCars(doc, &cars)
	// Save the scraped data to a file json formatted
	err = Save(&cars)
	if err != nil {
		fmt.Println("Error saving data:", err)
		return
	}
	err = SaveCSV(&cars)
	// Save the scraped data to a file csv formatted

	if err != nil {
		fmt.Println("Error saving data:", err)
		return
	}

	fmt.Println("Scraping completed successfully.")
	fmt.Println("Scraped data:", len(cars))

}
