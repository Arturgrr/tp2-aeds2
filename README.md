# Trabalho Prático 02: Manipulação e Reorganização de Arquivos de Dados

**Disciplina:** Algoritmos e Estruturas de Dados II  
**Professor:** Rafael Alexandre  
**Aluno:** Artur Assis Guerra  
**Matrícula:** 23.1.8006

---

## 1. Especificação

O programa expande o TP1 para suportar manipulação dinâmica de registros de alunos (CRUD) e reorganização física para otimização de espaço.

O arquivo de dados simula blocos de tamanho fixo em disco, onde registros podem ser inseridos, atualizados, removidos e compactados.

### Campos do Registro
| Campo | Tipo / Tamanho | Descrição |
|-------|----------------|-----------|
| **Status** | Byte (1 byte) | Flag de controle (0=Ativo, 1=Removido) |
| **Tamanho** | Inteiro (4 bytes) | Tamanho total do registro em bytes (Header + Payload) |
| Matrícula | Inteiro (4 bytes) | Identificador único |
| Nome | String Var | Nome do aluno |
| CPF | String Fixa | 11 caracteres |
| Curso | String Var | Curso |
| Filiação (mãe) | String Var | Nome da mãe |
| Filiação (pai) | String Var | Nome do pai |
| Ano de Ingresso | Inteiro (4 bytes) | Ano |
| CA | Float (8 bytes) | Coeficiente Acadêmico |

---

## 2. Funcionalidades Implementadas (TP2)

### 2.1. Inserção Inteligente (Create)
- **Estratégia First Fit**: Ao inserir um novo aluno, o sistema varre os blocos existentes procurando espaço livre no final (fragmentação interna) ou reutilizando blocos que foram esvaziados.
- **Expansão Dinâmica**: Se não houver espaço em nenhum bloco existente, um novo bloco é alocado no final do arquivo.

### 2.2. Leitura e Consulta (Read)
- **Filtragem de Excluídos**: Registros marcados logicamente como removidos são ignorados nas consultas e listagens.
- **Navegação Segura**: O sistema utiliza o header de tamanho para navegar sequencialmente pelos registros dentro de cada bloco.

### 2.3. Atualização (Update)
- **In-Place (Ideal)**: Se os novos dados do aluno (após edição) couberem no espaço original ou houver espaço sobrando no mesmo bloco, a atualização é feita no local, movendo os registros subsequentes se necessário.
- **Relocação**: Se o registro aumentar de tamanho e não couber no bloco original, ele é marcado como removido e re-inserido no final do arquivo ou em outro bloco com espaço (como uma nova inserção).

### 2.4. Remoção Lógica (Delete)
- **Tombstone**: A exclusão não apaga fisicamente os dados imediatamente. Apenas altera o byte de **Status** para `1` (Removido).
- **Recuperação de Espaço**: O espaço ocupado por registros removidos continua alocado no arquivo ("buraco") até que ocorra uma reorganização, mas a inserção inteligente pode reutilizar blocos se a fragmentação interna permitir.

### 2.5. Reorganização Física (Defragmentation)
- **Compactação**: Cria um novo arquivo (`_reorg.dat`), lendo apenas os registros ativos e gravando-os sequencialmente, eliminando buracos de exclusão e minimizando a fragmentação interna.
- **Relatório de Eficiência**: Ao final, exibe um comparativo de "Antes e Depois", mostrando o ganho de eficiência e redução de blocos.

---

## 3. Arquitetura e Estrutura de Pastas

O projeto segue princípios de **DDD (Domain-Driven Design)** e **Clean Code**.

```
tp1-aeds2/
├── domain/                    # Regras de negócio e Gerador de dados
│   └── generator.go
├── entity/                    # Entidades do domínio (Student)
│   └── student.go
├── infrastructure/            # Implementações concretas (Reporter)
│   └── reporter.go
├── storage/                   # Persistência em Arquivo
│   ├── interface.go          # Contrato Storage
│   ├── variable.go           # Implementação Principal (TP2)
│   └── fixed.go              # (Legado TP1)
└── main.go                    # CLI e Ponto de Entrada
```

---

## 4. Como Executar

### Pré-requisitos
- Go 1.25+ instalado.

### Execução
```bash
go run .
```

### Menu Principal
Ao iniciar, configure o tamanho do bloco (ex: 4096 bytes) e escolha o modo (Variável). O sistema apresentará o menu:

1. **Consultar aluno por matrícula**: Busca rápida.
2. **Consultar todos os alunos**: Lista ativos.
3. **Registrar novo aluno (Manual)**: Inserção unitária.
4. **Registrar lote de alunos**: Gera massa de dados.
5. **Atualizar dados de aluno**: Edição de campos.
6. **Remover aluno**: Exclusão lógica.
7. **Reorganizar arquivo**: Otimização física.
8. **Ver relatório**: Estatísticas de ocupação.
9. **Sair**

---

## 5. Análise de Eficiência

A implementação da reorganização demonstra claramente o impacto da fragmentação externa (buracos deixados por Deletes) e interna (espaço não usado no fim dos blocos).

**Exemplo de Ganho:**
Após excluir 50% dos alunos aleatoriamente e rodar a Reorganização, observa-se tipicamente uma redução de ~50% no número de blocos e um aumento na densidade de ocupação (Eficiência), pois os "buracos" são eliminados.

---
**Algoritmos e Estruturas de Dados II - 2025**
