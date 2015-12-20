/*
 * look at costs csv and output report
 */

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Cost struct {
	Name   string
	Amount float64
}

func (c Cost) String() string {
	return fmt.Sprintf("%s: %.2f", c.Name, c.Amount)
}

type ByAmount []Cost

func (c ByAmount) Len() int           { return len(c) }
func (c ByAmount) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByAmount) Less(i, j int) bool { return c[i].Amount > c[j].Amount }

type ByMonth []string

func (m ByMonth) Len() int      { return len(m) }
func (m ByMonth) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
func (m ByMonth) Less(i, j int) bool {
	iParts := strings.Split(m[i], "-")
	iYear64, _ := strconv.ParseInt(iParts[0], 10, 0)
	iMonth64, _ := strconv.ParseInt(iParts[1], 10, 0)

	jParts := strings.Split(m[j], "-")
	jYear64, _ := strconv.ParseInt(jParts[0], 10, 0)
	jMonth64, _ := strconv.ParseInt(jParts[1], 10, 0)

	if iYear64 < jYear64 {
		return true
	}
	if iYear64 > jYear64 {
		return false
	}
	if iMonth64 < jMonth64 {
		return true
	}
	return false
}

func main() {
	// log output format. 0 to be very minimal - no prefix.
	log.SetFlags(0)

	sourceToAmount := make(map[string]float64)
	monthToAmount := make(map[string]float64)
	monthToNameToAmount := make(map[string]map[string]float64)

	file, err := os.Open("costs.csv")
	if err != nil {
		log.Printf("Unable to open: %s", err.Error())
		os.Exit(1)
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		pieces := strings.Split(line, ",")
		if len(pieces) != 4 {
			log.Printf("Invalid line: %s", line)
			os.Exit(1)
		}

		date := pieces[0]
		source := pieces[1]
		amount := pieces[2]
		//desc := pieces[3]

		amountFloat, err := strconv.ParseFloat(amount, 64)
		if err != nil {
			log.Printf("Invalid amount: %s: %s", amount, err.Error())
			os.Exit(1)
		}

		dateParts := strings.Split(date, "-")
		if len(dateParts) != 3 {
			log.Fatalf("Invalid date: %s", date)
		}
		month := dateParts[0] + "-" + dateParts[1]

		// Overall total
		_, ok := sourceToAmount[source]
		if !ok {
			sourceToAmount[source] = float64(0)
		}
		sourceToAmount[source] += amountFloat

		// Month total
		_, ok = monthToAmount[month]
		if !ok {
			monthToAmount[month] = float64(0)
		}
		monthToAmount[month] += amountFloat

		_, ok = monthToNameToAmount[month]
		if !ok {
			monthToNameToAmount[month] = make(map[string]float64)
		}
		_, ok = monthToNameToAmount[month][source]
		if !ok {
			monthToNameToAmount[month][source] = float64(0)
		}
		monthToNameToAmount[month][source] += amountFloat
	}

	// Pull out to slice for sorting
	var totalCosts []Cost
	total := float64(0)
	for key, value := range sourceToAmount {
		cost := Cost{
			Name:   key,
			Amount: value,
		}
		totalCosts = append(totalCosts, cost)
		total += value
	}

	sort.Sort(ByAmount(totalCosts))

	log.Printf("Overall totals:")
	for _, value := range totalCosts {
		log.Print(value)
	}
	log.Printf("")
	log.Printf("Total: %.2f", total)

	log.Printf("")
	log.Printf("----")
	log.Printf("")

	// Pull out to slice for sorting
	var months []string
	for key, _ := range monthToAmount {
		months = append(months, key)
	}

	sort.Sort(ByMonth(months))

	log.Printf("By month:")
	for _, month := range months {
		amount := monthToAmount[month]
		log.Printf("%s: %.2f", month, amount)
	}

	log.Printf("")
	log.Printf("----")
	log.Printf("")

	log.Printf("Source by month:")
	for _, month := range months {
		sourceToAmount := monthToNameToAmount[month]
		log.Printf("%s", month)

		// Pull out costs to slice for sorting
		var costs []Cost
		for source, amount := range sourceToAmount {
			cost := Cost{Name: source, Amount: amount}
			costs = append(costs, cost)
		}
		sort.Sort(ByAmount(costs))

		for _, value := range costs {
			log.Printf("    %s", value)
		}
	}
}
