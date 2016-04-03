/*
 * Look at bills listed in a CSV and output a report.
 *
 * CSV lines look like:
 * YYYY-MM-DD,Store 123,10.00,Note about this
 *
 * TODO Don't handle money with floats
 */

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Cost is a parsed CSV line.
type Cost struct {
	Time   time.Time
	Source string
	Amount float64
	Note   string
}

func (c Cost) String() string {
	return fmt.Sprintf("%s %s: %.2f", c.Time.Format("2006-01-02"), c.Source,
		c.Amount)
}

type GroupedCost struct {
	Name   string
	Amount float64
}

func (c GroupedCost) String() string {
	return fmt.Sprintf("%s: %.2f", c.Name, c.Amount)
}

type CostByTime []Cost

func (m CostByTime) Len() int           { return len(m) }
func (m CostByTime) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m CostByTime) Less(i, j int) bool { return m[i].Time.Before(m[j].Time) }

type GroupedCostByAmount []GroupedCost

func (c GroupedCostByAmount) Len() int           { return len(c) }
func (c GroupedCostByAmount) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c GroupedCostByAmount) Less(i, j int) bool { return c[i].Amount > c[j].Amount }

// main is the program entry!
func main() {
	// Log output format. 0 to be very minimal - no prefix.
	log.SetFlags(0)

	csv := flag.String("csv", "", "CSV file to read.")
	locationString := flag.String("location", "America/Vancouver", "Time zone location.")
	daysBack := flag.Int("days-back", 30, "Number of days back to include in the report. Entries older than this will be ignored.")

	flag.Parse()

	if len(*csv) == 0 {
		log.Print("You must specify a CSV file.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if len(*locationString) == 0 {
		log.Print("You must specify a location.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *daysBack <= 0 {
		log.Print("You must provide a number of days back >= 0.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	location, err := time.LoadLocation(*locationString)
	if err != nil {
		log.Printf("Invalid location: %s", err.Error())
		os.Exit(1)
	}

	filterDuration := time.Duration(*daysBack*24) * time.Hour

	costs, err := readCostsCSV(*csv, location, filterDuration)
	if err != nil {
		log.Printf("Unable to read costs: %s", err.Error())
		os.Exit(1)
	}

	sourceToAmount := tallyCosts(costs)
	total := getTotal(costs)

	reportCosts(costs, total, sourceToAmount)
}

// readCostsCSV reads in a CSV and parses each line as a Cost.
func readCostsCSV(file string, location *time.Location,
	filterDuration time.Duration) ([]Cost, error) {
	fh, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("Unable to open: %s: %s", file, err.Error())
	}
	defer fh.Close()

	scanner := bufio.NewScanner(fh)

	timeLayout := "2006-01-02"

	filterTime := time.Now().Truncate(24 * time.Hour).Add(-filterDuration)
	log.Printf("Ignoring any entries < %s", filterTime.Format("2006-01-02"))

	var costs []Cost

	for scanner.Scan() {
		line := scanner.Text()

		pieces := strings.Split(line, ",")
		if len(pieces) != 4 {
			return nil, fmt.Errorf("Line missing expected number of fields: %s", line)
		}

		date := pieces[0]
		source := pieces[1]
		amount := pieces[2]
		note := pieces[3]

		costTime, err := time.ParseInLocation(timeLayout, date, location)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse date: %s: %s", date, err.Error())
		}

		if costTime.Before(filterTime) {
			continue
		}

		amountFloat, err := strconv.ParseFloat(amount, 64)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse amount: %s: %s", amount,
				err.Error())
		}

		costs = append(costs, Cost{
			Time:   costTime,
			Source: source,
			Amount: amountFloat,
			Note:   note,
		})
	}

	if scanner.Err() != nil {
		return nil, fmt.Errorf("Scanner error: %s", scanner.Err().Error())
	}

	return costs, nil
}

// getTotal totals up the costs
func getTotal(costs []Cost) float64 {
	total := float64(0)
	for _, cost := range costs {
		total += cost.Amount
	}
	return total
}

// tallyCosts builds some totals from the costs.
func tallyCosts(costs []Cost) map[string]GroupedCost {
	sourceToAmount := make(map[string]GroupedCost)

	for _, cost := range costs {
		_, ok := sourceToAmount[cost.Source]
		if !ok {
			sourceToAmount[cost.Source] = GroupedCost{Name: cost.Source}
		}
		sourceToAmount[cost.Source] = GroupedCost{
			Name:   cost.Source,
			Amount: sourceToAmount[cost.Source].Amount + cost.Amount,
		}
	}

	return sourceToAmount
}

// reportCosts outputs a report.
func reportCosts(costs []Cost, total float64,
	sourceToAmount map[string]GroupedCost) {
	// Output all costs, ordered by amount descending.
	sort.Sort(CostByTime(costs))

	log.Printf("Costs:")
	for _, value := range costs {
		log.Print(value)
	}
	log.Printf("")
	log.Printf("Total: %.2f", total)

	log.Printf("")
	log.Printf("----")
	log.Printf("")

	// Output totals of bill sources, ordered by amount descending.
	var groupedCosts []GroupedCost
	for _, groupedCost := range sourceToAmount {
		groupedCosts = append(groupedCosts, groupedCost)
	}
	sort.Sort(GroupedCostByAmount(groupedCosts))

	log.Print("Grouped costs:")
	for _, value := range groupedCosts {
		log.Print(value)
	}
}
