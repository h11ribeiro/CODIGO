/*
- Pontifícia Universidade Católica do Rio Grande do Sul
- Disciplina: de Fundamentos de Programação Paralelo e Distribuído
- Alunos: Giselle Chaves, Henrique Ribeiro e Gustavo Mesquita
- Professor: Cesar De Rose
*/
package main

import (
	"fmt"
	"sync"
)

type mensagem struct {
	tipo  int    // tipo da mensagem para fazer o controle do que fazer (eleição, confirmacao da eleicao)
	corpo [3]int // conteudo da mensagem para colocar os ids (usar um tamanho ocmpativel com o numero de processos no anel)
}

var (
	chans = []chan mensagem{ // vetor de canias para formar o anel de eleicao - chan[0], chan[1] and chan[2] ...
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
	}
	controle = make(chan int)
	wg       sync.WaitGroup // wg is used to wait for the program to finish
)
//-----------------------------------------------------------------------------------------------------------------------------
func ElectionControler(in chan int) {
	defer wg.Done()
	var temp mensagem
	// comandos para o anel iciam aqui

	// mudar o processo 0 - canal de entrada 3 - para falho (defini mensagem tipo 2 pra isto)
	temp.tipo = 2
	chans[3] <- temp
	fmt.Printf("Controle: mudar o processo 0 para falho\n")
	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

	// Iniciar nova eleição se o líder falhou
	temp.tipo = 1
	temp.corpo[0] = -1
	chans[3] <- temp
	fmt.Printf("Controle: iniciar nova eleição\n")

	// mudar o processo 1 - canal de entrada 3 - para falho (defini mensagem tipo 2 pra isto)
	temp.tipo = 2
	chans[0] <- temp
	fmt.Printf("Controle: mudar o processo 1 para falho\n")
	fmt.Printf("Controle: confirmação %d\n", <-in)
	
	//muda o processo 1 para funcional Recupera os valores falhos 
	temp.tipo = 2
	chans[0] <- temp
	fmt.Printf("Controle: indica que o valor 1 esta recuperado \n")
	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

	//Finaliza o programa 
	temp.tipo = 4
	for i := 0; i < len(chans); i++ {
		chans[i] <- temp
	}
	fmt.Println("\n   Processo controlador concluído")
}
//-----------------------------------------------------------------------------------------------------------------------------
func ElectionStage(TaskId int, in chan mensagem, out chan mensagem, leader int,) {
	defer wg.Done()
    var actualLeader int
	var bFailed bool = false // todos iniciam sem falha

    actualLeader = leader // indicação do lider veio por parâmatro
    
	for{
	// variaveis locais que indicam se este processo é o lider e se esta ativo
		select{
			case temp := <-in: // ler mensagem
			fmt.Printf("%2d: recebi mensagem %d, [ %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2]) 	

			switch temp.tipo {
        
				case 1: // Realiza a eleicao e atualiza o atual lider para o maior id - Erro
				
					if !bFailed {
						if TaskId > temp.corpo[0] {
							temp.corpo[0] = TaskId 
							fmt.Printf("%2d: novo líder é %d\n", TaskId, temp.corpo[0])
						}
						out <- temp
					}
				case 2:// Mensagem indicando falha do processo / adiciona falha no processo da proxima goroutines
					{	
						bFailed = true
						fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
						controle <- -5
					}
				case 3:// Mensagem do tipo 3 indica que o processo deve se recuperar/ recupera o processo falho da proxima goroutines
					{	
						bFailed = false
						fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
						controle <- -5
					}
				case 4: //Finaliza os Processos e o codigo
					{
						fmt.Printf("Processo %d: finalizando\n", TaskId)
						return
									
					}
			
				default: //Erro ao levar o tipo de mensagem 
					{
						fmt.Printf("%2d: não conheço este tipo de mensagem\n", TaskId)
						fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
					}
	
			}
		}
  }
}
//-----------------------------------------------------------------------------------------------------------------------------
func main() {

	wg.Add(5) // Adicione uma contagem de quatro, uma para cada goroutine

	// criar os processo do anel de eleicao

	go ElectionStage(0, chans[3], chans[0], 0) // este é o lider o processo 0 
	go ElectionStage(1, chans[0], chans[1], 0) 
	go ElectionStage(2, chans[1], chans[2], 0) 
	go ElectionStage(3, chans[2], chans[3], 0) 

	fmt.Println("\n  <----- Anel de processos criado ----->")

	go ElectionControler(controle)// criar o processo controlador

	fmt.Println("\n   <----- Processo controlador criado -----> \n")

	wg.Wait() //Aguarde o término das goroutines
}
//-----------------------------------------------------------------------------------------------------------------------------