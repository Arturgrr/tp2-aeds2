package main

import (
	"aeds2-tp1/domain"
	"aeds2-tp1/entity"
	"aeds2-tp1/infrastructure"
	"aeds2-tp1/storage"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const filename = "alunos.dat"

func main() {
	fmt.Println("=== Sistema de Armazenamento de Registros de Alunos ===")
	fmt.Println()

	if _, err := os.Stat(filename); err == nil {
		err := os.Remove(filename)
		if err != nil {
			fmt.Printf("Aviso: não foi possível deletar o arquivo %s existente: %v\n", filename, err)
		} else {
			fmt.Printf("Arquivo %s existente foi deletado. Criando novo arquivo.\n", filename)
		}
	}

	reader := bufio.NewReader(os.Stdin)

	numRecords := readInt(reader, "Digite o número de registros a serem gerados: ")
	blockSize := readInt(reader, "Digite o tamanho máximo do bloco (em bytes): ")

	fmt.Println("\nModo de armazenamento:")
	fmt.Println("1 - Registros de tamanho fixo")
	fmt.Println("2 - Registros de tamanho variável")
	storageMode := readInt(reader, "Escolha o modo (1 ou 2): ")

	var storageImpl storage.Storage
	var err error

	if storageMode == 1 {
		storageImpl, err = storage.NewFixedStorage(blockSize)
		if err != nil {
			fmt.Printf("Erro: %v\n", err)
			return
		}
	} else if storageMode == 2 {
		fmt.Println("\nTipo de armazenamento variável:")
		fmt.Println("1 - Contíguo (sem espalhamento)")
		fmt.Println("2 - Espalhado (com fragmentação entre blocos)")
		fragmentedMode := readInt(reader, "Escolha o tipo (1 ou 2): ")

		if fragmentedMode == 1 {
			storageImpl, err = storage.NewVariableStorage(blockSize)
			if err != nil {
				fmt.Printf("Erro: %v\n", err)
				return
			}
		} else if fragmentedMode == 2 {
			storageImpl, err = storage.NewVariableFragmentedStorage(blockSize)
			if err != nil {
				fmt.Printf("Erro: %v\n", err)
				return
			}
		} else {
			fmt.Println("Tipo inválido, usando contíguo por padrão")
			storageImpl, err = storage.NewVariableStorage(blockSize)
			if err != nil {
				fmt.Printf("Erro: %v\n", err)
				return
			}
		}
	} else {
		fmt.Println("Modo inválido, usando tamanho variável contíguo por padrão")
		storageImpl, err = storage.NewVariableStorage(blockSize)
		if err != nil {
			fmt.Printf("Erro: %v\n", err)
			return
		}
	}

	fmt.Println("\nGerando registros de alunos...")
	generator := domain.NewStudentGenerator()
	students := generator.Generate(numRecords)
	fmt.Printf("Gerados %d registros de alunos\n", len(students))

	fmt.Println("\nGravando registros no arquivo alunos.dat...")
	err = storageImpl.WriteStudents(filename, students)
	if err != nil {
		fmt.Printf("Erro ao gravar arquivo: %v\n", err)
		return
	}
	fmt.Println("Arquivo gravado com sucesso!")

	stats := storageImpl.GetStats(filename)
	reporter := infrastructure.NewReporter(stats)
	reporter.PrintStats()
	reporter.PrintBlockMap()
	reporter.PrintBlockVisualization()

	runQueryMode(reader, storageImpl)
}

func runQueryMode(reader *bufio.Reader, storageImpl storage.Storage) {
	for {
		fmt.Println("\n=== MENU PRINCIPAL ===")
		fmt.Println("1 - Consultar aluno por matrícula")
		fmt.Println("2 - Consultar todos os alunos")
		fmt.Println("3 - Registrar novo aluno (Manual)")
		fmt.Println("4 - Registrar lote de alunos (Gerador)")
		fmt.Println("5 - Atualizar dados de aluno")
		fmt.Println("6 - Remover aluno")
		fmt.Println("7 - Reorganizar arquivo")
		fmt.Println("8 - Ver relatório de armazenamento")
		fmt.Println("9 - Sair")
		
		option := readInt(reader, "Escolha uma opção: ")

		switch option {
		case 1:
			matricula := readInt(reader, "Digite a matrícula do aluno: ")
			student, err := storageImpl.FindStudentByMatricula(filename, matricula)
			if err != nil {
				fmt.Printf("Erro: %v\n", err)
			} else {
				printStudent(student)
			}
		case 2:
			listAllStudents(storageImpl)
		case 3:
			registerManualStudent(reader, storageImpl)
		case 4:
			registerNewStudents(reader, storageImpl)
		case 5:
			updateStudentData(reader, storageImpl)
		case 6:
			removeStudent(reader, storageImpl)
		case 7:
			reorganizeFile(storageImpl)
		case 8:
			showStorageReport(storageImpl)
		case 9:
			return
		default:
			fmt.Println("Opção inválida!")
		}
	}
}

func listAllStudents(storageImpl storage.Storage) {
	fmt.Println("\n=== TODOS OS ALUNOS ===")
	students, err := storageImpl.GetAllStudents(filename)
	if err != nil {
		fmt.Printf("Erro ao listar alunos: %v\n", err)
		return
	}

	if len(students) == 0 {
		fmt.Println("Nenhum aluno encontrado.")
		return
	}

	fmt.Printf("Total de alunos: %d\n\n", len(students))
	for i, student := range students {
		fmt.Printf("%d. Matrícula: %d - %s - Curso: %s - CA: %.2f\n", 
			i+1, student.Matricula, student.Nome, student.Curso, student.CA)
	}
}

func registerManualStudent(reader *bufio.Reader, storageImpl storage.Storage) {
	fmt.Println("\n=== REGISTRAR ALUNO MANUALMENTE ===")
	
	matricula := readInt(reader, "Matrícula: ")
	nome := readString(reader, "Nome: ")
	cpf := readString(reader, "CPF (11 dígitos): ")
	curso := readString(reader, "Curso: ")
	mae := readString(reader, "Filiação Mãe: ")
	pai := readString(reader, "Filiação Pai: ")
	ano := readInt(reader, "Ano Ingresso: ")
	ca := readFloat(reader, "CA (0.0 a 10.0): ")

	student := entity.Student{
		Matricula:   matricula,
		Nome:        nome,
		CPF:         cpf,
		Curso:       curso,
		FiliacaoMae: mae,
		FiliacaoPai: pai,
		AnoIngresso: ano,
		CA:          ca,
	}
	
	student.TruncateFields()
	if err := student.Validate(); err != nil {
		fmt.Printf("Erro de validação: %v\n", err)
		return
	}
	
	err := storageImpl.AddStudents(filename, []entity.Student{student})
	if err != nil {
		fmt.Printf("Erro ao adicionar aluno: %v\n", err)
	} else {
		fmt.Println("Aluno registrado com sucesso!")
	}
}

func updateStudentData(reader *bufio.Reader, storageImpl storage.Storage) {
	fmt.Println("\n=== ATUALIZAR ALUNO ===")
	matricula := readInt(reader, "Digite a matrícula do aluno a atualizar: ")
	
	existing, err := storageImpl.FindStudentByMatricula(filename, matricula)
	if err != nil {
		fmt.Printf("Erro: %v\n", err)
		return
	}
	
	printStudent(existing)
	fmt.Println("\nDigite os novos dados (pressione Enter para manter o atual):")
	
	// Lógica simplificada de atualização campo a campo
	// Como readString retorna vazio se Enter, podemos checar
	
	fmt.Printf("Nome [%s]: ", existing.Nome)
	nome := readStringOptional(reader)
	if nome != "" { existing.Nome = nome }
	
	fmt.Printf("Curso [%s]: ", existing.Curso)
	curso := readStringOptional(reader)
	if curso != "" { existing.Curso = curso }
	
	fmt.Printf("CA [%.2f]: ", existing.CA)
	caStr := readStringOptional(reader)
	if caStr != "" {
		ca, err := strconv.ParseFloat(caStr, 64)
		if err == nil { existing.CA = ca }
	}

	// Outros campos... (simplificando para o exemplo, pode expandir se quiser)
	// Para ser minucioso, deveria pedir todos ou menu de qual campo.
	// Vamos pedir todos os principais que mudam tamanho.
	
	fmt.Printf("Filiação Mãe [%s]: ", existing.FiliacaoMae)
	mae := readStringOptional(reader)
	if mae != "" { existing.FiliacaoMae = mae }
	
	fmt.Printf("Filiação Pai [%s]: ", existing.FiliacaoPai)
	pai := readStringOptional(reader)
	if pai != "" { existing.FiliacaoPai = pai }
	
	// CPF geralmente não muda, mas...
	
	existing.TruncateFields()
	if err := existing.Validate(); err != nil {
		fmt.Printf("Dados inválidos: %v\n", err)
		return
	}
	
	err = storageImpl.UpdateStudent(filename, *existing)
	if err != nil {
		fmt.Printf("Erro ao atualizar: %v\n", err)
	} else {
		fmt.Println("Aluno atualizado com sucesso!")
	}
}

func removeStudent(reader *bufio.Reader, storageImpl storage.Storage) {
	fmt.Println("\n=== REMOVER ALUNO ===")
	matricula := readInt(reader, "Digite a matrícula do aluno a remover: ")
	
	err := storageImpl.DeleteStudent(filename, matricula)
	if err != nil {
		fmt.Printf("Erro ao remover: %v\n", err)
	} else {
		fmt.Println("Aluno removido com sucesso (Exclusão Lógica)!")
	}
}

func reorganizeFile(storageImpl storage.Storage) {
	fmt.Println("\n=== REORGANIZAR ARQUIVO ===")
	fmt.Println("Iniciando compactação...")
	
	report, err := storageImpl.Reorganize(filename)
	if err != nil {
		fmt.Printf("Erro na reorganização: %v\n", err)
		return
	}
	
	fmt.Println("\n===== RELATÓRIO DE REORGANIZAÇÃO =====")
	fmt.Println("Antes:")
	fmt.Printf("Blocos: %d\n", report.BlocksBefore)
	fmt.Printf("Ocupação média: %.1f%%\n", report.OccupancyBefore)
	fmt.Printf("Eficiência total: %.1f%%\n", report.EfficiencyBefore)
	
	fmt.Println("\nDepois:")
	fmt.Printf("Blocos: %d\n", report.BlocksAfter)
	fmt.Printf("Ocupação média: %.1f%%\n", report.OccupancyAfter)
	fmt.Printf("Eficiência total: %.1f%%\n", report.EfficiencyAfter)
	
	fmt.Printf("\nGanho de eficiência: %+.1f%%\n", report.EfficiencyGain)
	fmt.Printf("Blocos liberados: %d\n", report.FreedBlocks)
	fmt.Println("======================================")
	
	fmt.Printf("\nArquivo reorganizado salvo como: %s_reorg.dat\n", strings.TrimSuffix(filename, ".dat"))
}

func registerNewStudents(reader *bufio.Reader, storageImpl storage.Storage) {
	fmt.Println("\n=== REGISTRAR LOTE DE ALUNOS ===")
	numRecords := readInt(reader, "Digite o número de alunos a serem gerados: ")
	
	fmt.Println("\nGerando novos alunos...")
	generator := domain.NewStudentGenerator()
	students := generator.Generate(numRecords)
	fmt.Printf("Gerados %d novos alunos\n", len(students))

	fmt.Println("\nAdicionando alunos ao arquivo...")
	err := storageImpl.AddStudents(filename, students)
	if err != nil {
		fmt.Printf("Erro ao adicionar alunos: %v\n", err)
		return
	}
	fmt.Println("Alunos adicionados com sucesso!")
}

func showStorageReport(storageImpl storage.Storage) {
	stats := storageImpl.GetStats(filename)
	reporter := infrastructure.NewReporter(stats)
	reporter.PrintStats()
	reporter.PrintBlockMap()
	reporter.PrintBlockVisualization()
}


func printStudent(student *entity.Student) {
	fmt.Println("\n=== DADOS DO ALUNO ===")
	fmt.Printf("Matrícula:     %d\n", student.Matricula)
	fmt.Printf("Nome:           %s\n", student.Nome)
	fmt.Printf("CPF:            %s\n", student.CPF)
	fmt.Printf("Curso:          %s\n", student.Curso)
	fmt.Printf("Filiação Mãe:   %s\n", student.FiliacaoMae)
	fmt.Printf("Filiação Pai:   %s\n", student.FiliacaoPai)
	fmt.Printf("Ano de Ingresso: %d\n", student.AnoIngresso)
	fmt.Printf("CA:             %.2f\n", student.CA)
}

func readInt(reader *bufio.Reader, prompt string) int {
	for {
		fmt.Print(prompt)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		value, err := strconv.Atoi(input)
		if err == nil && value >= 0 { // Aceitar 0 em alguns casos, ou só > 0? Melhor >= 0 para flexibilidade
			return value
		}
		fmt.Println("Valor inválido. Digite um número inteiro.")
	}
}

func readString(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func readStringOptional(reader *bufio.Reader) string {
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func readFloat(reader *bufio.Reader, prompt string) float64 {
	for {
		fmt.Print(prompt)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		value, err := strconv.ParseFloat(input, 64)
		if err == nil {
			return value
		}
		fmt.Println("Valor inválido. Digite um número decimal.")
	}
}

