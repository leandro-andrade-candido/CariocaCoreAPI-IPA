package main

import (
	"database/sql"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/nakagami/firebirdsql"
)

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	response := request.Body
	fmt.Println("Received body: ", response)

	conn, _ := sql.Open("firebirdsql", "SYSDBA:masterkey@moveiscarioca.certaddns.com.br:7525//storage/firebird/dados/DADOSMC.FDB")
	defer conn.Close()

	query := fmt.Sprintf("select EXTRACT(MONTH FROM DATAEMISSAO_NF) mes, EXTRACT(YEAR FROM DATAEMISSAO_NF) ano, SUM(VALORTOTALITENS_NF) total "+
		"FROM %s WHERE DATAEMISSAO_NF BETWEEN '%s' AND '%s' AND ENTRADASAIDA_NF = '%s' AND EMPRESA_NF = %d AND CFOP_NF IN(%s) GROUP BY 2,1;",
		"NOTA_FISCAL", "2018-01-01", "2019-08-01", "S", 1, "5101,6101,6118")

	result, err := conn.Query(query)
	if err != nil {
		response = err.Error()
	} else {
		fmt.Println("Result")
		fmt.Println(result)
	}

	return events.APIGatewayProxyResponse{Body: response, StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handler)
}
