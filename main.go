package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/pricing"
)

//Product is a product
type Product struct {
	ProductFamily string `json:"productFamily"`
	Sku           string `json:"sku"`
	Attributes    map[string]string
	Terms         Terms `json:"terms"`
}

// OnDemand is a thing
type OnDemand struct {
	Price float32 `json:"USD"`
}

//Attributes is product attributes
type Attributes struct {
	ClockSpeed         string `json:"clockSpeed"`
	NetworkPerformance string `json:"networkPerformance"`
	Vcpu               int    `json:"vcpu"`
}

// Terms is the actual pricing information
type Terms struct {
}

func getCost(itype string) float64 {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))
	price := pricing.New(sess)

	pIn := pricing.GetProductsInput{
		ServiceCode: aws.String("AmazonEC2"),
		Filters: []*pricing.Filter{
			&pricing.Filter{
				Type:  aws.String("TERM_MATCH"),
				Field: aws.String("instanceType"),
				Value: aws.String("m5.2xlarge"),
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
				Value: aws.String("USW1-BoxUsage:m5.2xlarge"),
			},
		},
	}

	res, err := price.GetProducts(&pIn)
	if err != nil {
		log.Fatal(err)
	}

	var dollas string

	cost := res.PriceList[0]["terms"].(map[string]interface{})["OnDemand"]
	for _, c := range cost.(map[string]interface{}) {
		for _, d := range c.(map[string]interface{})["priceDimensions"].(map[string]interface{}) {
			dollas = d.(map[string]interface{})["pricePerUnit"].(map[string]interface{})["USD"].(string)
			break
		}
	}
	hourlyCost, err := strconv.ParseFloat(dollas, 64)
	return hourlyCost

}

type instance struct {
	vips       int
	eips       int
	hourlyCost float64
}

func main() {
	fmt.Println(getCost("none"))

	var instanceType map[string]instance

	// b, err := json.MarshalIndent(res.PriceList, "", "  ")
	// fmt.Println(string(b))
	// var product Product

	// err = json.Unmarshal(b, &product)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(product)
	// fmt.Println(len(res.PriceList))
	// for k, v := range res.PriceList[0]["product"].(map[string]interface{}) {

	// 	fmt.Println(k, v)

	// 	if k == "attributes" {
	// 		fmt.Println
	// 	}
	// 	// switch val := v.(type) {
	// 	// case string:
	// 	// 	fmt.Println(k, "is string", val)
	// 	// case int:
	// 	// 	fmt.Println(k, "is int", val)
	// 	// case []interface{}:
	// 	// 	fmt.Println(k, "is an array")
	// 	// 	for i, v := range val {
	// 	// 		fmt.Println(i, v)
	// 	// 	}
	// 	// default:
	// 	// 	fmt.Println(k, "is unknown type")
	// 	// }
	// }
	// fmt.Println()
	// for k, v := range res.PriceList[0]["product"].([]string) {
	// 	fmt.Println(k, v)
	// }
	//var p Product
	// test, err := json.Unmarshal(res.PriceList)

	// fmt.Println(test)

	// aws.jsonutil
	// var f interface{}
	// _ = json.Unmarshal([]byte(res.PriceList[0]["product"]), &f)
	//json.Marshal

	// fmt.Println(len(res.PriceList))
	// fmt.Printf("%T\n", res.PriceList[0]["product"])

	// t := res.PriceList[0]["product"].([]byte)

	// aws.JSONValue

	// _ =

	// for _, num := range res.PriceList {

	// }
	// fmt.Println(len(res.PriceList))

	// c := colly.NewCollector()

	// // Find and visit all links
	// c.OnHTML("tbody", func(e *colly.HTMLElement) {
	// 	e.ForEach("tr", func(i int, tr *colly.HTMLElement) {
	// 		var itype string
	// 		var ifaces, enis int

	// 		tr.ForEach("td", func(j int, td *colly.HTMLElement) {

	// 			switch j {
	// 			case 0:
	// 				itype = strings.TrimSpace(td.Text)
	// 			case 1:
	// 				ifaces, _ = strconv.Atoi(strings.TrimSpace(td.Text))
	// 			case 2:
	// 				enis, _ = strconv.Atoi(strings.TrimSpace(td.Text))
	// 			}
	// 		})
	// 		if i > 0 {
	// 			maxIps := ifaces * enis
	// 			fmt.Println(itype, maxIps)
	// 		}

	// 	})
	// })

	// c.Visit("https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-eni.html")
}
