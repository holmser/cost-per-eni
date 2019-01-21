package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/gocolly/colly"
)

type test interface {
	cpi() float64
	getCost() float64
}

func (i Instance) calcGPI() float64 {
	//fmt.Println((i.hourlyCost * 750.0) / float64(i.enis) * float64(i.ifaces))
	return (i.hourlyCost * 750.0) / (float64(i.enis) * float64(i.ifaces))
}

func getCost(price *pricing.Pricing, instance Instance, ch chan Instance, writer *csv.Writer) {
	pIn := pricing.GetProductsInput{
		ServiceCode: aws.String("AmazonEC2"),
		Filters: []*pricing.Filter{
			&pricing.Filter{
				Type:  aws.String("TERM_MATCH"),
				Field: aws.String("instanceType"),
				Value: aws.String(instance.itype),
			},
			&pricing.Filter{
				Type:  aws.String("TERM_MATCH"),
				Field: aws.String("operatingSystem"),
				Value: aws.String("Linux"),
			},
			&pricing.Filter{
				Type:  aws.String("TERM_MATCH"),
				Field: aws.String("tenancy"),
				Value: aws.String("Shared"),
			},
			&pricing.Filter{
				Type:  aws.String("TERM_MATCH"),
				Field: aws.String("preInstalledSw"),
				Value: aws.String("NA"),
			},
			&pricing.Filter{
				Type:  aws.String("TERM_MATCH"),
				Field: aws.String("capacityStatus"),
				Value: aws.String("Used"),
			},
			&pricing.Filter{
				Type:  aws.String("TERM_MATCH"),
				Field: aws.String("usagetype"),
				Value: aws.String("USE2-BoxUsage:" + instance.itype),
			},
		},
	}
	// fmt.Println(itype)
	res, err := price.GetProducts(&pIn)

	if err != nil {
		instance.hourlyCost = 0.0
		log.Println(err)
	}
	if len(res.PriceList) > 0 {
		var dollas string

		// so ugly, needs refactoring
		cost := res.PriceList[0]["terms"].(map[string]interface{})["OnDemand"]
		for _, c := range cost.(map[string]interface{}) {
			for _, d := range c.(map[string]interface{})["priceDimensions"].(map[string]interface{}) {
				dollas = d.(map[string]interface{})["pricePerUnit"].(map[string]interface{})["USD"].(string)
				break
			}
		}
		attr := res.PriceList[0]["product"].(map[string]interface{})["attributes"].(map[string]interface{})
		instance.bandwidth = attr["networkPerformance"].(string)
		instance.memory = attr["memory"].(string)
		instance.vcpu, _ = strconv.Atoi(attr["vcpu"].(string))
		hourlyCost, err := strconv.ParseFloat(dollas, 64)
		checkError("failed parsing hourly cost", err)
		instance.hourlyCost = hourlyCost

	} else {
		instance.hourlyCost = 0.0
	}

	fmt.Println(instance.itype, instance.enis, instance.ifaces, instance.hourlyCost, instance.calcGPI(), instance.bandwidth, instance.vcpu, instance.memory)
	//fmt.Println()
	writeCSV(instance, writer)
	ch <- instance
}

// Instance contains all the data needed to identify the instance cost
type Instance struct {
	itype      string
	ifaces     int
	enis       int
	hourlyCost float64
	cpi        float64
	bandwidth  string
	vcpu       int
	memory     string
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

func main() {
	// var wg sync.WaitGroup

	file, err := os.Create("result.csv")
	checkError("Cannot create file", err)
	defer file.Close()
	writer := csv.NewWriter(file)
	err = writer.Write([]string{"Type", "Hourly Cost", "ENI", "EIPs", "Cost/IP/mo", "vCPU", "Memory", "Bandwidth"})

	rate := time.Second / 5
	throttle := time.Tick(rate)

	ch := make(chan Instance)

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))

	price := pricing.New(sess)
	//var instances []Instance

	c := colly.NewCollector()

	// Find and visit all links
	c.OnHTML("tbody", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(i int, tr *colly.HTMLElement) {
			var instance Instance
			tr.ForEach("td", func(j int, td *colly.HTMLElement) {

				switch j {
				case 0:
					instance.itype = strings.TrimSpace(td.Text)
				case 1:
					instance.ifaces, _ = strconv.Atoi(strings.TrimSpace(td.Text))
				case 2:
					instance.enis, _ = strconv.Atoi(strings.TrimSpace(td.Text))
				}
			})
			if i > 0 {

				// wg.Add(1)
				// fmt.Println("calling routines for", instance)
				<-throttle
				go getCost(price, instance, ch, writer)

			}

		})
	})

	c.Visit("https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-eni.html")
}

func writeCSV(i Instance, writer *csv.Writer) {
	defer writer.Flush()

	err := writer.Write([]string{i.itype, fmt.Sprintf("%f", i.hourlyCost), strconv.Itoa(i.ifaces), strconv.Itoa(i.enis), fmt.Sprintf("%f", i.calcGPI()), strconv.Itoa(i.vcpu), i.memory, i.bandwidth})
	checkError("Error writing csv", err)

}

func printChan(ch chan Instance) {
	res := (<-ch)
	fmt.Println(res)
}
