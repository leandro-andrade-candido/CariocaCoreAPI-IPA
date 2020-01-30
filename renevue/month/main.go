package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jmoiron/sqlx"
	_ "github.com/nakagami/firebirdsql"
	"ipa/pkg/cache"
	"log"
	"os"
)

const (
	cachePort   = "6379"
	cacheName   = "renevue"
	cacheExpire = "5" //minutes
)

var cacheEndpoint = os.Getenv("REDIS_ENDPOINT")
var cacheKey = fmt.Sprintf("%s-%s", os.Getenv("ENV"), cacheName)

type Renevue struct {
	Mes   string `db:"MES"`
	Ano   string `db:"ANO"`
	Total int32  `db:"TOTAL"`
}

type Response events.APIGatewayProxyResponse

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	response := request.Body

	log.Println("Received body::", response)

	log.Println("ENV::", cacheEndpoint, cachePort)

	// @todo controlar se der erro no init cache
	var ch = cache.Init(cacheEndpoint, cachePort)

	var renevue []Renevue

	cached, err := ch.GetKey(cacheKey)
	log.Printf("CACHED %s:: %#v %v\n", cacheKey, cached, err)

	if cached == nil {

		log.Printf("FIREBIRD CONNECT::\n")

		db, err := sqlx.Connect("firebirdsql", "SYSDBA:masterkey@moveiscarioca.certaddns.com.br:7525//storage/firebird/dados/DADOSMC.FDB")
		if err != nil {
			log.Fatalln(err)
		}
		defer db.Close()

		err = db.Select(
			&renevue,
			"select EXTRACT(MONTH FROM DATAEMISSAO_NF) MES, EXTRACT(YEAR FROM DATAEMISSAO_NF) ANO, CAST(SUM(VALORTOTALITENS_NF) AS INTEGER) TOTAL FROM NOTA_FISCAL WHERE DATAEMISSAO_NF BETWEEN ? AND ? AND ENTRADASAIDA_NF = ? AND EMPRESA_NF = ? AND CFOP_NF IN(5101,6101,6118) GROUP BY 2,1;",
			"2018-01-01", "2019-08-01", "S", 1)
		if err != nil {
			log.Fatal("Query failed:", err.Error())
		}

		log.Println("Result")
		log.Println(renevue)

		log.Printf("CACHING %s:: %#v %v\n", cacheKey, renevue, cacheExpire)

		_, _ = ch.SetKey(cacheKey, renevue, cacheExpire)

	}

	log.Printf("RESPONSING::\n")

	var buf bytes.Buffer

	body, err := json.Marshal(renevue)
	if err != nil {
		return events.APIGatewayProxyResponse(Response{StatusCode: 404}), err
	}
	json.HTMLEscape(&buf, body)

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type":         "application/json",
			"X-Carioca-Func-Reply": "renevue",
		},
	}

	return events.APIGatewayProxyResponse(resp), nil
}

func main() {
	lambda.Start(Handler)
}
