package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// Estruturas para mapear as respostas das APIs
type BrasilAPIResponse struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
}

type ViaCEPResponse struct {
	Cep        string `json:"cep"`
	Logradouro string `json:"logradouro"`
	Bairro     string `json:"bairro"`
	Localidade string `json:"localidade"`
	Uf         string `json:"uf"`
}

// Estrutura comum para exibir o endereço
type Address struct {
	Cep          string
	State        string
	City         string
	Neighborhood string
	Street       string
	Source       string // API que forneceu os dados
}

func main() {
	// Verifica se o CEP foi passado como argumento
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run main.go <CEP>")
		return
	}
	cep := os.Args[1]

	resultChan := make(chan Address)

	// Inicia goroutines para as duas APIs
	go func() {
		address, err := fetchFromBrasilAPI(cep)
		if err != nil {
			return
		}
		address.Source = "BrasilAPI"
		resultChan <- address
	}()

	go func() {
		address, err := fetchFromViaCEP(cep)
		if err != nil {
			return
		}
		address.Source = "ViaCEP"
		resultChan <- address
	}()

	// Aguarda o primeiro resultado ou timeout de 1 segundo
	select {
	case address := <-resultChan:
		fmt.Printf("Resposta recebida da %s:\n", address.Source)
		fmt.Printf("CEP: %s\n", address.Cep)
		fmt.Printf("Estado: %s\n", address.State)
		fmt.Printf("Cidade: %s\n", address.City)
		fmt.Printf("Bairro: %s\n", address.Neighborhood)
		fmt.Printf("Rua: %s\n", address.Street)
	case <-time.After(1 * time.Second):
		fmt.Println("Erro de timeout")
	}
}

func fetchFromBrasilAPI(cep string) (Address, error) {
	var address Address
	client := &http.Client{
		Timeout: 1 * time.Second,
	}
	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)
	resp, err := client.Get(url)
	if err != nil {
		return address, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return address, fmt.Errorf("BrasilAPI retornou status %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return address, err
	}
	var apiResp BrasilAPIResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		return address, err
	}
	address = Address{
		Cep:          apiResp.Cep,
		State:        apiResp.State,
		City:         apiResp.City,
		Neighborhood: apiResp.Neighborhood,
		Street:       apiResp.Street,
	}
	return address, nil
}

func fetchFromViaCEP(cep string) (Address, error) {
	var address Address
	client := &http.Client{
		Timeout: 1 * time.Second,
	}
	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)
	resp, err := client.Get(url)
	if err != nil {
		return address, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return address, fmt.Errorf("ViaCEP retornou status %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return address, err
	}
	var apiResp ViaCEPResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		return address, err
	}
	address = Address{
		Cep:          apiResp.Cep,
		State:        apiResp.Uf,
		City:         apiResp.Localidade,
		Neighborhood: apiResp.Bairro,
		Street:       apiResp.Logradouro,
	}
	return address, nil
}
