package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	ErrInvalidCepMessage = "CEP inválido, um cep válido deve ser no formato 12345678"
	TimeoutMessage       = "Timeout"
	ServiceBrasilAPI     = "BrasilApi"
	ServiceViaCep        = "ViaCep"
)

var (
	BaseURLBrasilAPI = os.Getenv("BASE_URI_BRASIL_API")
	BaseURLViaCep    = os.Getenv("BASE_URI_VIA_CEP")
)

type AddressData struct {
	Cep     string `json:"cep"`
	City    string `json:"city"`
	State   string `json:"state"`
	Service string
}

type AddressDataViaCep struct {
	Cep   string `json:"cep"`
	City  string `json:"localidade"`
	State string `json:"uf"`
}

func main() {
	cep := os.Args[1]
	if len(cep) != 8 {
		fmt.Println(ErrInvalidCepMessage)
		os.Exit(1)
	}
	brazilAPIChannel := make(chan AddressData)
	viaCepChannel := make(chan AddressData)

	go fetchBrasilAPIAddress(cep, brazilAPIChannel)
	go fetchViaCepAddress(cep, viaCepChannel)

	select {
	case addressData := <-brazilAPIChannel:
		printAddressData(addressData)
	case addressData := <-viaCepChannel:
		printAddressData(addressData)
	case <-time.After(time.Second):
		fmt.Println(TimeoutMessage)
	}
}

func fetchBrasilAPIAddress(cep string, resultChannel chan AddressData) {
	response, _ := http.Get(BaseURLBrasilAPI + cep)
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)

	var address AddressData
	json.Unmarshal(body, &address)
	address.Service = ServiceBrasilAPI
	resultChannel <- address
}

func fetchViaCepAddress(cep string, resultChannel chan AddressData) {
	response, _ := http.Get(BaseURLViaCep + cep + "/json")
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)

	var viaCepData AddressDataViaCep
	json.Unmarshal(body, &viaCepData)
	resultChannel <- mapViaCepDataToAddress(viaCepData)
}

func mapViaCepDataToAddress(viaCepData AddressDataViaCep) AddressData {
	return AddressData{
		Cep:     viaCepData.Cep,
		City:    viaCepData.City,
		State:   viaCepData.State,
		Service: ServiceViaCep,
	}
}

func printAddressData(data AddressData) {
	fmt.Println("CEP:     ", data.Cep)
	fmt.Println("Cidade:  ", data.City)
	fmt.Println("Estado:  ", data.State)
	fmt.Println("Serviço: ", data.Service)
}
