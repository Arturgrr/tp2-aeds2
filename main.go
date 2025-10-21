package main

import (
	"fmt"
	"os"
)

func main() {
	file, err := os.Open("data.dat")

	if os.IsNotExist(err) {
		file, err = os.Create("data.dat")
		if err != nil {
			fmt.Println("Erro ao criar o arquivo:", err)
			return
		}
		fmt.Println("Arquivo .dat não existia e foi criado com sucesso!")
	} else if err != nil {
		fmt.Println("Erro ao abrir o arquivo:", err)
		return
	} else {
		fmt.Println("Arquivo .dat aberto com sucesso!")
	}

	defer file.Close()
}

