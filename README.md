# Trabalho Prático 01: Manipulação e Organização de Arquivos de Dados

**Disciplina:** Algoritmos e Estruturas de Dados II  
**Professor:** Rafael Alexandre  
**Aluno:** Artur Assis Guerra  
**Matrícula:** 23.1.8006

---

## 1. Especificação

O programa simula o armazenamento de um conjunto de registros de alunos, com os seguintes campos:

| Campo | Tipo / Tamanho | Descrição |
|-------|----------------|-----------|
| Matrícula | Inteiro (9 dígitos) | Identificador único do aluno |
| Nome | String (até 50 caracteres) | Nome completo do aluno |
| CPF | String (11 caracteres) | CPF do aluno |
| Curso | String (até 30 caracteres) | Curso em que o aluno está matriculado |
| Filiação (mãe) | String (até 30 caracteres) | Nome completo da mãe |
| Filiação (pai) | String (até 30 caracteres) | Nome completo do pai |
| Ano de Ingresso | Inteiro (4 dígitos) | Ano de entrada na instituição |
| CA | Float (2 casas decimais) | Coeficiente Acadêmico |

O conjunto de registros é gerado automaticamente utilizando um gerador de dados fictícios implementado no pacote `domain`.

O usuário informa:
1. O tamanho máximo do bloco (em bytes)
2. O modo de armazenamento:
   - (a) Registros de tamanho fixo
   - (b) Registros de tamanho variável
3. Caso o modo 2 seja selecionado, escolhe:
   - (a) Registros contíguos
   - (b) Registros espalhados

---

## 2. Regras de Armazenamento Implementadas

### 2.1. Tamanho Fixo

**Características:**
- Todos os registros ocupam o mesmo número de bytes (162 bytes)
- O tamanho é definido a partir do maior tamanho possível de um registro:
  - Matrícula: 4 bytes (int32)
  - Nome: 50 bytes (string fixa)
  - CPF: 11 bytes (string fixa)
  - Curso: 30 bytes (string fixa)
  - Filiação Mãe: 30 bytes (string fixa)
  - Filiação Pai: 30 bytes (string fixa)
  - Ano de Ingresso: 4 bytes (int32)
  - CA: 8 bytes (float64 armazenado como int64 * 100)

**Preenchimento:**
- Campos menores são preenchidos com o caractere `#`
- CPF é preenchido com `0` à esquerda se necessário

**Restrição:**
- Cada registro deve ser armazenado integralmente dentro de um único bloco
- Se um registro não couber no bloco atual, ele é movido para o próximo bloco

**Validação:**
- O sistema valida se o tamanho do bloco é suficiente para armazenar pelo menos um registro completo
- Tamanho mínimo: 162 bytes

### 2.2. Tamanho Variável Contíguo

**Características:**
- O tamanho de cada registro depende do conteúdo real dos campos
- Campos de string variável são armazenados com:
  - 4 bytes para o tamanho (int32)
  - N bytes para o conteúdo real

**Estrutura do Registro Variável:**
- Matrícula: 4 bytes (int32)
- Nome: 4 bytes (tamanho) + N bytes (conteúdo)
- CPF: 11 bytes (fixo)
- Curso: 4 bytes (tamanho) + N bytes (conteúdo)
- Filiação Mãe: 4 bytes (tamanho) + N bytes (conteúdo)
- Filiação Pai: 4 bytes (tamanho) + N bytes (conteúdo)
- Ano de Ingresso: 4 bytes (int32)
- CA: 8 bytes (float64 como int64 * 100)

**Regra de Espalhamento:**
- **SEM espalhamento**: Se um registro não couber completamente no bloco atual, ele é movido integralmente para o próximo bloco
- Registros nunca são divididos entre blocos

**Validação:**
- O sistema valida se cada registro individual cabe em um bloco
- Se um registro exceder o tamanho do bloco, retorna erro informativo
- Tamanho mínimo do bloco: ~162 bytes (considerando campos no tamanho máximo)

### 2.3. Tamanho Variável Espalhado

**Características:**
- Registros podem ser fragmentados entre múltiplos blocos
- Permite melhor aproveitamento do espaço disponível

**Estrutura de Fragmentação:**
- Cada fragmento possui um header de 5 bytes:
  - 1 byte: Flag de continuação (0 = último fragmento, 1 = há continuação)
  - 4 bytes: Tamanho do fragmento (int32)
- Os dados do registro seguem após o header

**Algoritmo de Fragmentação:**
1. Se o registro cabe no espaço disponível do bloco atual, é gravado completamente
2. Caso contrário:
   - Grava o máximo possível no bloco atual
   - Cria um novo bloco e continua gravando o restante
   - Repete até completar o registro

**Reconstrução:**
- Para ler um registro fragmentado, o sistema:
  1. Lê o primeiro fragmento
  2. Verifica a flag de continuação
  3. Se houver continuação, lê os blocos subsequentes
  4. Concatena todos os fragmentos
  5. Deserializa o registro completo

**Validação:**
- Tamanho mínimo do bloco: ~162 bytes
- Cada fragmento deve ter pelo menos 5 bytes (header) + 1 byte de dados

---

## 3. Arquitetura e Decisões de Projeto

### 3.1. Estrutura de Pastas

```
tp1-aeds2/
├── domain/                    # Camada de Domínio
│   └── generator.go          # Lógica de geração de dados fictícios
├── entity/                    # Camada de Entidades
│   └── student.go            # Entidade Student com validação
├── infrastructure/            # Camada de Infraestrutura
│   └── reporter.go           # Geração de relatórios e estatísticas
├── storage/                   # Camada de Armazenamento
│   ├── interface.go          # Interface Storage e tipos
│   ├── fixed.go              # Implementação tamanho fixo
│   ├── variable.go           # Implementação variável contíguo
│   └── variable_fragmented.go # Implementação variável espalhado
└── main.go                    # Interface principal
```

### 3.2. Validação de Dados

A entidade `Student` utiliza a biblioteca `github.com/go-playground/validator/v10` para validação declarativa:

```go
type Student struct {
    Matricula   int     `validate:"required,min=1,matricula_digits"`
    Nome        string  `validate:"required,min=1,max=50"`
    CPF         string  `validate:"required,cpf_format"`
    Curso       string  `validate:"required,min=1,max=30"`
    FiliacaoMae string  `validate:"required,min=1,max=30"`
    FiliacaoPai string  `validate:"required,min=1,max=30"`
    AnoIngresso int     `validate:"required,min=1000,max=9999"`
    CA          float64 `validate:"required,min=0.0,max=10.0"`
}
```

**Validações Customizadas:**
- `matricula_digits`: Verifica se a matrícula tem no máximo 9 dígitos
- `cpf_format`: Verifica se o CPF tem exatamente 11 caracteres e apenas dígitos

### 3.3. Serialização

**Formato de Dados:**
- **Tamanho Fixo**: Campos com tamanho fixo, preenchidos quando necessário
- **Tamanho Variável**: Campos precedidos por 4 bytes indicando o tamanho
- **CA (Coeficiente Acadêmico)**: Armazenado como `int64(CA * 100)` para preservar 2 casas decimais

### 3.4. Interface de Armazenamento

A interface `Storage` define o contrato comum para todas as implementações:

```go
type Storage interface {
    WriteStudents(filename string, students []entity.Student) error
    FindStudentByMatricula(filename string, matricula int) (*entity.Student, error)
    GetAllStudents(filename string) ([]*entity.Student, error)
    AddStudents(filename string, students []entity.Student) error
    GetStats(filename string) StorageStats
    ValidateBlockSize(blockSize int) error
    GetBlockSize() int
}
```
---

## 4. Implementação das Estratégias

### 4.1. Armazenamento de Tamanho Fixo (`storage/fixed.go`)

**Algoritmo de Escrita:**
1. Calcula o tamanho fixo do registro (162 bytes)
2. Para cada estudante:
   - Serializa o registro com padding (`#` para strings, `0` para CPF)
   - Verifica se cabe no bloco atual
   - Se não couber, finaliza o bloco atual e cria novo
   - Adiciona o registro ao bloco
3. Finaliza o último bloco com padding

**Algoritmo de Leitura:**
1. Calcula número de blocos: `tamanho_arquivo / tamanho_bloco`
2. Para cada bloco:
   - Lê o bloco completo
   - Percorre em incrementos de `fixedRecordSize`
   - Deserializa cada registro
   - Remove padding dos campos string

### 4.2. Armazenamento Variável Contíguo (`storage/variable.go`)

**Algoritmo de Escrita:**
1. Para cada estudante:
   - Serializa o registro com tamanhos variáveis
   - Verifica se o registro cabe no bloco atual
   - Se não couber, finaliza o bloco atual e cria novo
   - Adiciona o registro ao bloco (inteiro)
2. Finaliza o último bloco

**Algoritmo de Leitura:**
1. Para cada bloco:
   - Lê o bloco completo
   - Percorre sequencialmente:
     - Lê matrícula (4 bytes)
     - Lê tamanho do nome (4 bytes) + nome
     - Lê CPF (11 bytes)
     - Lê tamanho do curso (4 bytes) + curso
     - Lê tamanho da filiação mãe (4 bytes) + filiação mãe
     - Lê tamanho da filiação pai (4 bytes) + filiação pai
     - Lê ano de ingresso (4 bytes)
     - Lê CA (8 bytes)
   - Avança para o próximo registro

### 4.3. Armazenamento Variável Espalhado (`storage/variable_fragmented.go`)

**Algoritmo de Escrita:**
1. Para cada estudante:
   - Serializa o registro
   - Se cabe no espaço disponível: grava completo
   - Se não cabe:
     - Grava fragmento no bloco atual (com header)
     - Cria novo bloco
     - Continua gravando fragmentos até completar
     - Último fragmento tem flag = 0

**Estrutura do Fragmento:**
```
[Flag: 1 byte][Tamanho: 4 bytes][Dados: N bytes]
```

**Algoritmo de Leitura:**
1. Para cada bloco:
   - Lê header (flag + tamanho)
   - Lê dados do fragmento
   - Se flag = 1: continua no próximo bloco
   - Concatena todos os fragmentos
   - Deserializa o registro completo

---

## 5. Funcionalidades Implementadas

### 5.1. Geração de Dados

O gerador (`domain/generator.go`) cria registros fictícios com:
- **Matrículas**: Sequenciais a partir de 100000001
- **Nomes**: Selecionados aleatoriamente de uma lista de 20 nomes
- **CPFs**: Gerados aleatoriamente com 11 dígitos
- **Cursos**: Selecionados de uma lista de 10 cursos
- **Filiações**: Nomes de mãe e pai selecionados aleatoriamente
- **Ano de Ingresso**: Entre 2015 e 2024 (aleatório)
- **CA**: Entre 5.0 e 10.0 com 2 casas decimais

### 5.2. Menu Interativo

O programa oferece um menu completo com as seguintes opções:

1. **Consultar aluno por matrícula**: Busca um aluno específico no arquivo
2. **Consultar todos os alunos**: Lista todos os alunos registrados
3. **Registrar novos alunos**: Adiciona novos alunos ao arquivo existente
4. **Ver relatório de armazenamento**: Exibe estatísticas detalhadas
5. **Sair**: Encerra o programa

### 5.3. Estatísticas e Relatórios

O sistema calcula e exibe:

**Estatísticas Gerais:**
- Número total de blocos utilizados
- Total de bytes utilizados
- Total de bytes disponíveis
- Eficiência total de armazenamento (%)
- Número de blocos parcialmente utilizados
- Percentual médio de ocupação

**Mapa de Ocupação:**
- Lista detalhada de cada bloco:
  - Número do bloco
  - Bytes utilizados
  - Percentual de ocupação
  - Número de registros

**Visualização Gráfica:**
- Barras de ocupação usando caracteres Unicode (█ e ░)
- Representação visual da ocupação de cada bloco

### 5.4. Validações e Tratamento de Erros

**Validação de Tamanho de Bloco:**
- Verifica se o bloco é grande o suficiente para armazenar pelo menos um registro
- Retorna erro descritivo se o tamanho for insuficiente

**Validação de Registros:**
- Verifica se registros variáveis não excedem o tamanho do bloco
- Retorna erro informativo com detalhes do problema

**Tratamento de Arquivo:**
- Se `alunos.dat` existir ao iniciar, é deletado automaticamente
- Mensagens claras de erro e sucesso

---

## 6. Como Executar

### 6.1. Usando Binários Pré-compilados (Recomendado)

O projeto possui releases automáticas com binários compilados para todas as plataformas. A cada commit em qualquer branch, uma nova release é criada automaticamente.

**Passos:**

1. Acesse a página de [Releases](https://github.com/Arturgrr/tp1-aeds2/releases) do repositório
2. Baixe o binário correspondente à sua plataforma:
   - **Windows**: `tp1-aeds2-windows-amd64.exe`
   - **Linux**: `tp1-aeds2-linux-amd64`
   - **macOS (Intel)**: `tp1-aeds2-darwin-amd64`
   - **macOS (Apple Silicon)**: `tp1-aeds2-darwin-arm64`
3. Execute o binário:
   - **Windows**: Clique duas vezes no arquivo `.exe` ou execute no terminal
   - **Linux/macOS**: 
     ```bash
     chmod +x tp1-aeds2-linux-amd64
     ./tp1-aeds2-linux-amd64
     ```

### 6.2. Compilação Local

Se preferir compilar localmente:

**Pré-requisitos:**
- Go 1.25.3 ou superior
- Terminal/Console

**Passos:**

```bash
# Navegar até o diretório do projeto
cd tp1-aeds2

# Baixar dependências
go mod download

# Compilar o programa
go build -o tp1-aeds2 .

# Executar
./tp1-aeds2
```

Ou executar diretamente sem compilar:

```bash
go run .
```

### 6.3. Compilação para Múltiplas Plataformas

Para compilar para diferentes plataformas:

```bash
# Windows (amd64)
GOOS=windows GOARCH=amd64 go build -o tp1-aeds2-windows-amd64.exe .

# Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o tp1-aeds2-linux-amd64 .

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o tp1-aeds2-darwin-amd64 .

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o tp1-aeds2-darwin-arm64 .
```

**Tamanho Variável Espalhado:**
- **Vantagem**: Máximo aproveitamento (pode usar quase 100% do espaço)
- **Desvantagem**: Overhead de 5 bytes por fragmento, complexidade maior
- **Eficiência típica**: 90-98%

## 7. Conclusão

Este trabalho implementa com sucesso um sistema de armazenamento de registros em arquivos binários, demonstrando três estratégias distintas de organização de dados. A implementação segue princípios de Clean Code e Domain-Driven Design, resultando em código limpo, organizado e fácil de manter.

**Principais Conquistas:**
- ✅ Implementação completa das três estratégias de armazenamento
- ✅ Validação robusta de dados usando biblioteca validator
- ✅ Interface polimórfica permitindo extensibilidade
- ✅ Relatórios detalhados de eficiência e ocupação
- ✅ Código organizado seguindo DDD e Clean Code
- ✅ Tratamento adequado de limites de bloco
- ✅ Suporte a registros fragmentados entre blocos

O sistema está pronto para uso e demonstra compreensão dos conceitos de organização de arquivos, armazenamento em blocos e diferentes estratégias de alocação de dados em disco.
