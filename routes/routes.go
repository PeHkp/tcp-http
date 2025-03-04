package routes

import (
	"errors"
	"fmt"
	"servidor-tcp/services"
)

var (
	ErrRouteNotFound = errors.New("rota não encontrada")
	ErrServiceFailed = errors.New("falha no serviço")
)

func Routes(route string) (string, error) {
	switch route {
	case "/":
		dados, err := services.GetCountries()
		if err != nil {
			fmt.Printf("Erro ao obter dados dos países: %v\n", err)
			return "", fmt.Errorf("%w: não foi possível obter dados dos países", ErrServiceFailed)
		}
		return dados, nil
		
	case "/search":
		return "Bem-vindo à página sobre", nil
		
		
	default:
		fmt.Printf("Tentativa de acesso a rota inexistente: %s\n", route)
		return "", fmt.Errorf("%w: %s", ErrRouteNotFound, route)
	}
}

func GetAvailableRoutes() []string {
	return []string{
		"/",
		"/search",
	}
}
