package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	ErrHTTPRequest = errors.New("erro na requisição HTTP")
	ErrHTTPStatus  = errors.New("status HTTP inválido")
	ErrReadBody    = errors.New("erro ao ler corpo da resposta")
)

func GetCountries() (string, error) {
	url := "https://restcountries.com/v3.1/all"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrHTTPRequest, err.Error())
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrHTTPRequest, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("%w: código %d", ErrHTTPStatus, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrReadBody, err.Error())
	}

	if len(body) == 0 {
		return "", fmt.Errorf("corpo da resposta vazio")
	}

	var countries []map[string]interface{}
	err = json.Unmarshal(body, &countries)
	if err != nil {
		return "", fmt.Errorf("erro ao decodificar JSON: %s", err.Error())
	}

	limitedCountries := countries
	if len(countries) > 5 {
		limitedCountries = countries[:5]
	}

	limitedBody, err := json.Marshal(limitedCountries)
	if err != nil {
		return "", fmt.Errorf("erro ao codificar JSON: %s", err.Error())
	}

	body = limitedBody

	return string(body), nil
}

func HandleCountriesRequest() {
	data, err := GetCountries()
	if err != nil {
		switch {
		case errors.Is(err, ErrHTTPRequest):
			fmt.Println("Erro de conexão. Verifique sua internet.")
		case errors.Is(err, ErrHTTPStatus):
			fmt.Println("A API retornou um erro. Tente novamente mais tarde.")
		case errors.Is(err, ErrReadBody):
			fmt.Println("Erro ao processar a resposta.")
		default:
			fmt.Println("Erro inesperado:", err)
		}
		return
	}

	fmt.Println("Dados recebidos com sucesso. Tamanho:", len(data))
}
