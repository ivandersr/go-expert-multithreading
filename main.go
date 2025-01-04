package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type FetchResult struct {
	Data string
	URL  string
}

func main() {
	// informar CEP como argumento (e.g. go run main.go 23013770)
	cep := os.Args[1]
	brasilApi := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)
	viaCepApi := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)

	ch := make(chan FetchResult)
	chErr := make(chan string)

	go fetch(brasilApi, ch, chErr)
	go fetch(viaCepApi, ch, chErr)

	select {
	case msg := <-ch:
		fmt.Printf("Dados do Endereço: %s\nURL da API: %s\n", msg.Data, msg.URL)
	case msg := <-chErr:
		fmt.Println(msg)
	case <-time.After(time.Second):
		fmt.Println("Timeout")
	}
}

func fetch(url string, ch chan FetchResult, chErr chan string) {
	resp, err := http.Get(url)
	if err != nil {
		chErr <- fmt.Sprintf("Erro de requisição à URL %s: %v", url, err.Error())
		return
	}
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			chErr <- fmt.Sprint("CEP não encontrado")
			return
		}
		chErr <- fmt.Sprintf("CEP com formato inválido na URL %s", url)
		return
	}
	body, err := io.ReadAll(resp.Body)
	var bodyMap map[string]interface{}
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		chErr <- fmt.Sprint("Erro no parsing da resposta da requisição")
		return
	}
	if _, ok := bodyMap["erro"]; ok {
		chErr <- fmt.Sprint("CEP não encontrado")
		return
	}
	msg := FetchResult{
		Data: string(body),
		URL:  url,
	}
	ch <- msg
}
