package main

import (
	"fmt"
	"sync"
)

type mensagem struct {
	tipo  int    // tipo da mensagem para fazer o controle do que fazer (eleição, confirmação da eleição)
	corpo [3]int // conteúdo da mensagem para colocar os IDs (usar um tamanho compatível com o número de processos no anel)
}

var (
	chans = []chan mensagem{
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
	}
	controle = make(chan int)
	wg       sync.WaitGroup
)

func ElectionControler(in chan int) {
	defer wg.Done()
	var temp mensagem

	// Simular falhas e recuperações
	temp.tipo = 2
	chans[3] <- temp
	fmt.Printf("Controle: mudar o processo 0 para falho\n")
	fmt.Printf("Controle: confirmação %d\n", <-in)

	// Iniciar nova eleição se o líder falhou
	temp.tipo = 1
	temp.corpo[0] = -1
	chans[3] <- temp
	// fmt.Printf("Controle: iniciar nova eleição\n")

	temp.tipo = 2
	chans[0] <- temp
	fmt.Printf("Controle: mudar o processo 1 para falho\n")
	fmt.Printf("Controle: confirmação %d\n", <-in)

	temp.tipo = 2
	chans[2] <- temp
	fmt.Printf("Controle: mudar o processo 4 para falho\n")
	fmt.Printf("Controle: confirmação %d\n", <-in)
		
	temp.tipo = 4
	for i := 0; i < len(chans); i++ {
			chans[i] <- temp
	}
	fmt.Println("\nProcesso controlador concluído")
}

func ElectionStage(TaskId int, in chan mensagem, out chan mensagem, leader int) {
	defer wg.Done()
	var bFailed bool = false

	for {
		select {
		case temp := <-in:
			fmt.Printf("%2d: recebi mensagem %d, [ %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2])

			switch temp.tipo {
			case 1: // Realiza a eleição e atualiza o atual líder para o maior id
				if !bFailed {
					//fmt.Printf("Controle: iniciar nova eleição\n")
					if temp.corpo[0] == TaskId {
						fmt.Printf("%2d: Eu sou o novo líder\n", TaskId)
						leader = TaskId
					} else {
						if TaskId > temp.corpo[0] {
							fmt.Printf("%2d: novo líder é %d\n", TaskId, temp.corpo[0])
							temp.corpo[0] = TaskId
						}
						out <- temp
					}
				}
				case 2: // Mensagem indicando falha do processo
					bFailed = true
					fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
					controle <- -5

					// Se o processo falho é o líder, iniciar uma nova eleição
					if leader == TaskId {
						temp.tipo = 1
						temp.corpo[0] = TaskId
						out <- temp
						//fmt.Printf("%2d: Iniciando nova eleição\n", TaskId)
				}
				case 3: // Mensagem do tipo 3 indica que o processo deve se recuperar
					bFailed = false
					fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
					controle <- -5
				case 4: // Finaliza os Processos
					fmt.Printf("Processo %d: finalizando\n", TaskId)
					return
				default: // Erro ao ler o tipo de mensagem
					fmt.Printf("%2d: Não conheço este tipo de mensagem\n", TaskId)
				}
			}
	}
}

func main() {
    wg.Add(5) // Adicione uma contagem de cinco, uma para cada goroutine

    go ElectionStage(0, chans[3], chans[0], 0)
    go ElectionStage(1, chans[0], chans[1], 0)
    go ElectionStage(2, chans[1], chans[2], 0)
    go ElectionStage(3, chans[2], chans[3], 0)

    fmt.Println("\n<----- Anel de processos criado ----->")

    go ElectionControler(controle) // criar o processo controlador

    fmt.Println("\n<----- Processo controlador criado -----> \n")

    wg.Wait() // Aguarde o término das goroutines
}
