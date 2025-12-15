package storage

import (
	"aeds2-tp1/entity"
	"encoding/binary"
	"fmt"
	"os"
)

const (
	StatusActive  = 0
	StatusDeleted = 1
)

type VariableStorage struct {
	blockSize int
	stats      StorageStats
}

func NewVariableStorage(blockSize int) (*VariableStorage, error) {
	vs := &VariableStorage{
		blockSize: blockSize,
		stats: StorageStats{
			BlockStatsList: make([]BlockStats, 0),
		},
	}
	
	if err := vs.ValidateBlockSize(blockSize); err != nil {
		return nil, err
	}
	
	return vs, nil
}

func (vs *VariableStorage) ValidateBlockSize(blockSize int) error {
	minSize := 4 + 1 +
		4 + entity.MaxNomeLength +
		entity.CPFLength +
		4 + entity.MaxCursoLength +
		4 + entity.MaxFiliacaoLength +
		4 + entity.MaxFiliacaoLength +
		4 +
		8
	
	if blockSize < minSize {
		return fmt.Errorf("tamanho do bloco (%d bytes) é menor que o tamanho mínimo necessário para um registro variável (%d bytes)", blockSize, minSize)
	}
	
	return nil
}

func (vs *VariableStorage) WriteStudents(filename string, students []entity.Student) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo: %w", err)
	}
	defer file.Close()

	currentBlock := make([]byte, 0, vs.blockSize)
	currentBlockNumber := 0
	blockStats := BlockStats{
		BlockNumber: currentBlockNumber,
		BytesUsed:   0,
		BytesTotal:  vs.blockSize,
	}

	for i, student := range students {
		recordData := vs.serializeStudent(student)
		
		if len(recordData) > vs.blockSize {
			return fmt.Errorf("registro do aluno %d (matrícula: %d) excede o tamanho do bloco (%d bytes > %d bytes). Aumente o tamanho do bloco", i+1, student.Matricula, len(recordData), vs.blockSize)
		}
		
		vs.writeContiguousRecord(&currentBlock, &currentBlockNumber, &blockStats, recordData, file, i == len(students)-1)
	}

	if len(currentBlock) > 0 {
		blockStats.OccupancyRate = float64(blockStats.BytesUsed) / float64(blockStats.BytesTotal) * 100
		vs.stats.BlockStatsList = append(vs.stats.BlockStatsList, blockStats)
		vs.writeBlock(file, currentBlock)
		vs.stats.TotalBlocks++
		vs.stats.TotalBytesUsed += blockStats.BytesUsed
		vs.stats.TotalBytesTotal += blockStats.BytesTotal
	}

	vs.calculateFinalStats()
	return nil
}

func (vs *VariableStorage) serializeStudent(student entity.Student) []byte {
	payload := make([]byte, 0)
	
	matriculaBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(matriculaBytes, uint32(student.Matricula))
	payload = append(payload, matriculaBytes...)
	
	nomeBytes := []byte(student.Nome)
	nomeLen := make([]byte, 4)
	binary.LittleEndian.PutUint32(nomeLen, uint32(len(nomeBytes)))
	payload = append(payload, nomeLen...)
	payload = append(payload, nomeBytes...)
	
	cpfBytes := []byte(student.CPF)
	payload = append(payload, cpfBytes...)
	
	cursoBytes := []byte(student.Curso)
	cursoLen := make([]byte, 4)
	binary.LittleEndian.PutUint32(cursoLen, uint32(len(cursoBytes)))
	payload = append(payload, cursoLen...)
	payload = append(payload, cursoBytes...)
	
	maeBytes := []byte(student.FiliacaoMae)
	maeLen := make([]byte, 4)
	binary.LittleEndian.PutUint32(maeLen, uint32(len(maeBytes)))
	payload = append(payload, maeLen...)
	payload = append(payload, maeBytes...)
	
	paiBytes := []byte(student.FiliacaoPai)
	paiLen := make([]byte, 4)
	binary.LittleEndian.PutUint32(paiLen, uint32(len(paiBytes)))
	payload = append(payload, paiLen...)
	payload = append(payload, paiBytes...)
	
	anoBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(anoBytes, uint32(student.AnoIngresso))
	payload = append(payload, anoBytes...)
	
	caBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(caBytes, uint64(student.CA*100))
	payload = append(payload, caBytes...)
	
	totalSize := uint32(1 + len(payload))
	
	finalData := make([]byte, 0, 4+totalSize)
	
	sizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBytes, totalSize)
	finalData = append(finalData, sizeBytes...)
	
	finalData = append(finalData, byte(StatusActive))
	finalData = append(finalData, payload...)
	
	return finalData
}

func (vs *VariableStorage) writeContiguousRecord(currentBlock *[]byte, currentBlockNumber *int, blockStats *BlockStats, recordData []byte, file *os.File, isLast bool) {
	recordSize := len(recordData)
	
	if len(*currentBlock)+recordSize > vs.blockSize {
		if len(*currentBlock) > 0 {
			blockStats.OccupancyRate = float64(blockStats.BytesUsed) / float64(blockStats.BytesTotal) * 100
			if blockStats.OccupancyRate < 100 {
				vs.stats.PartialBlocks++
			}
			vs.stats.BlockStatsList = append(vs.stats.BlockStatsList, *blockStats)
			vs.writeBlock(file, *currentBlock)
			vs.stats.TotalBlocks++
			vs.stats.TotalBytesUsed += blockStats.BytesUsed
			vs.stats.TotalBytesTotal += blockStats.BytesTotal
		}
		
		*currentBlock = make([]byte, 0, vs.blockSize)
		*currentBlockNumber++
		*blockStats = BlockStats{
			BlockNumber: *currentBlockNumber,
			BytesUsed:   0,
			BytesTotal:  vs.blockSize,
		}
	}
	
	*currentBlock = append(*currentBlock, recordData...)
	blockStats.BytesUsed += recordSize
	blockStats.RecordsCount++
}

func (vs *VariableStorage) writeBlock(file *os.File, block []byte) {
	paddedBlock := make([]byte, vs.blockSize)
	copy(paddedBlock, block)
	file.Write(paddedBlock)
}

func (vs *VariableStorage) calculateFinalStats() {
	if vs.stats.TotalBytesTotal > 0 {
		vs.stats.EfficiencyRate = float64(vs.stats.TotalBytesUsed) / float64(vs.stats.TotalBytesTotal) * 100
	}
}

func (vs *VariableStorage) GetStats(filename string) StorageStats {
	vs.recalculateStatsFromFile(filename)
	return vs.stats
}

func (vs *VariableStorage) recalculateStatsFromFile(filename string) {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return
	}

	totalBlocks := int(fileInfo.Size()) / vs.blockSize
	vs.stats = StorageStats{
		TotalBlocks:     totalBlocks,
		TotalBytesTotal: totalBlocks * vs.blockSize,
		BlockStatsList:  make([]BlockStats, 0),
	}

	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	totalUsed := 0
	for blockNum := 0; blockNum < totalBlocks; blockNum++ {
		block := make([]byte, vs.blockSize)
		_, err := file.ReadAt(block, int64(blockNum*vs.blockSize))
		if err != nil {
			continue
		}

		bytesUsed := 0
		recordsCount := 0
		offset := 0
		for offset < vs.blockSize {
			if offset+4 > vs.blockSize {
				break
			}
			
			if block[offset] == 0 && block[offset+1] == 0 && block[offset+2] == 0 && block[offset+3] == 0 {
				break
			}

			student, bytesConsumed, err := vs.deserializeStudentFromBlock(block, offset)
			if err != nil {
				if bytesConsumed > 0 {
					offset += bytesConsumed
					continue
				}
				break
			}

			if student != nil {
				recordsCount++
			}
			
			bytesUsed += bytesConsumed
			offset += bytesConsumed
		}

		totalUsed += bytesUsed

		occupancyRate := float64(bytesUsed) / float64(vs.blockSize) * 100
		blockStats := BlockStats{
			BlockNumber:   blockNum,
			BytesUsed:     bytesUsed,
			BytesTotal:    vs.blockSize,
			OccupancyRate: occupancyRate,
			RecordsCount:  recordsCount,
		}

		if occupancyRate < 100 && occupancyRate > 0 {
			vs.stats.PartialBlocks++
		}

		vs.stats.BlockStatsList = append(vs.stats.BlockStatsList, blockStats)
	}

	vs.stats.TotalBytesUsed = totalUsed
	if vs.stats.TotalBytesTotal > 0 {
		vs.stats.EfficiencyRate = float64(vs.stats.TotalBytesUsed) / float64(vs.stats.TotalBytesTotal) * 100
	}
}

func (vs *VariableStorage) FindStudentByMatricula(filename string, matricula int) (*entity.Student, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter informações do arquivo: %w", err)
	}

	totalBlocks := int(fileInfo.Size()) / vs.blockSize
	return vs.findStudentContiguous(file, totalBlocks, matricula)
}

func (vs *VariableStorage) findStudentContiguous(file *os.File, totalBlocks int, matricula int) (*entity.Student, error) {
	for blockNum := 0; blockNum < totalBlocks; blockNum++ {
		block := make([]byte, vs.blockSize)
		_, err := file.ReadAt(block, int64(blockNum*vs.blockSize))
		if err != nil {
			continue
		}

		offset := 0
		for offset < vs.blockSize {
			if offset+4 > vs.blockSize {
				break
			}

			if block[offset] == 0 && block[offset+1] == 0 && block[offset+2] == 0 && block[offset+3] == 0 {
				break
			}

			student, bytesConsumed, err := vs.deserializeStudentFromBlock(block, offset)
			if err != nil {
				if bytesConsumed > 0 {
					offset += bytesConsumed
					continue
				}
				break
			}
			
			if student != nil {
				if student.Matricula == matricula {
					return student, nil
				}
			}

			offset += bytesConsumed
		}
	}

	return nil, fmt.Errorf("aluno com matrícula %d não encontrado", matricula)
}

func (vs *VariableStorage) deserializeStudentFromBlock(block []byte, offset int) (*entity.Student, int, error) {
	if offset+4 > len(block) {
		return nil, 0, fmt.Errorf("offset fora dos limites")
	}

	totalSize := int(binary.LittleEndian.Uint32(block[offset : offset+4]))
	
	if totalSize == 0 {
		return nil, 0, fmt.Errorf("tamanho de registro zero encontrado")
	}

	if offset+4+totalSize > len(block) {
		return nil, 0, fmt.Errorf("registro excede limites do bloco")
	}

	status := block[offset+4]
	
	bytesConsumed := 4 + totalSize

	if status == StatusDeleted {
		return nil, bytesConsumed, nil 
	}

	payloadStart := offset + 5
	payloadEnd := offset + 4 + totalSize

	recordData := block[payloadStart:payloadEnd]
	student, err := vs.deserializeStudent(recordData)
	if err != nil {
		return nil, bytesConsumed, err
	}

	return student, bytesConsumed, nil
}

func (vs *VariableStorage) getRecordSize(student *entity.Student) int {
	size := 4 + 1
	size += 4 + len(student.Nome)
	size += 11
	size += 4 + len(student.Curso)
	size += 4 + len(student.FiliacaoMae)
	size += 4 + len(student.FiliacaoPai)
	size += 4
	size += 8
	return size
}

func (vs *VariableStorage) deserializeStudent(data []byte) (*entity.Student, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("dados insuficientes")
	}

	offset := 0

	matricula := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	if offset+4 > len(data) {
		return nil, fmt.Errorf("dados insuficientes para nome")
	}
	nomeLen := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	if offset+nomeLen > len(data) {
		return nil, fmt.Errorf("dados insuficientes para nome")
	}
	nome := string(data[offset : offset+nomeLen])
	offset += nomeLen

	if offset+11 > len(data) {
		return nil, fmt.Errorf("dados insuficientes para CPF")
	}
	cpf := string(data[offset : offset+11])
	offset += 11

	if offset+4 > len(data) {
		return nil, fmt.Errorf("dados insuficientes para curso")
	}
	cursoLen := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	if offset+cursoLen > len(data) {
		return nil, fmt.Errorf("dados insuficientes para curso")
	}
	curso := string(data[offset : offset+cursoLen])
	offset += cursoLen

	if offset+4 > len(data) {
		return nil, fmt.Errorf("dados insuficientes para filiação mãe")
	}
	maeLen := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	if offset+maeLen > len(data) {
		return nil, fmt.Errorf("dados insuficientes para filiação mãe")
	}
	filiacaoMae := string(data[offset : offset+maeLen])
	offset += maeLen

	if offset+4 > len(data) {
		return nil, fmt.Errorf("dados insuficientes para filiação pai")
	}
	paiLen := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	if offset+paiLen > len(data) {
		return nil, fmt.Errorf("dados insuficientes para filiação pai")
	}
	filiacaoPai := string(data[offset : offset+paiLen])
	offset += paiLen

	if offset+4 > len(data) {
		return nil, fmt.Errorf("dados insuficientes para ano de ingresso")
	}
	anoIngresso := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	if offset+8 > len(data) {
		return nil, fmt.Errorf("dados insuficientes para CA")
	}
	caValue := binary.LittleEndian.Uint64(data[offset : offset+8])
	ca := float64(caValue) / 100.0

	student := &entity.Student{
		Matricula:   matricula,
		Nome:        nome,
		CPF:         cpf,
		Curso:       curso,
		FiliacaoMae: filiacaoMae,
		FiliacaoPai: filiacaoPai,
		AnoIngresso: anoIngresso,
		CA:          ca,
	}

	student.TruncateFields()

	if err := student.Validate(); err != nil {
		return nil, fmt.Errorf("estudante deserializado inválido: %w", err)
	}

	return student, nil
}

func (vs *VariableStorage) GetBlockSize() int {
	return vs.blockSize
}

func (vs *VariableStorage) GetAllStudents(filename string) ([]*entity.Student, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter informações do arquivo: %w", err)
	}

	totalBlocks := int(fileInfo.Size()) / vs.blockSize
	students := make([]*entity.Student, 0)

	for blockNum := 0; blockNum < totalBlocks; blockNum++ {
		block := make([]byte, vs.blockSize)
		_, err := file.ReadAt(block, int64(blockNum*vs.blockSize))
		if err != nil {
			continue
		}

		offset := 0
		for offset < vs.blockSize {
			if offset+4 > vs.blockSize {
				break
			}

			if block[offset] == 0 && block[offset+1] == 0 && block[offset+2] == 0 && block[offset+3] == 0 {
				break
			}

			student, bytesConsumed, err := vs.deserializeStudentFromBlock(block, offset)
			if err != nil {
				if bytesConsumed > 0 {
					offset += bytesConsumed
					continue
				}
				break
			}
			
			if student != nil && student.Matricula > 0 {
				students = append(students, student)
			}
			
			offset += bytesConsumed
		}
	}

	return students, nil
}

// AddStudents com inserção inteligente (Best/First Fit no final dos blocos)
func (vs *VariableStorage) AddStudents(filename string, students []entity.Student) error {
	vs.recalculateStatsFromFile(filename)

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer file.Close()

	for _, student := range students {
		recordData := vs.serializeStudent(student)
		recordSize := len(recordData)

		if recordSize > vs.blockSize {
			return fmt.Errorf("registro muito grande para o bloco")
		}

		inserted := false

		for i, blockStats := range vs.stats.BlockStatsList {
			if blockStats.BytesUsed + recordSize <= vs.blockSize {
				block := make([]byte, vs.blockSize)
				_, err := file.ReadAt(block, int64(i * vs.blockSize))
				if err == nil {
					offset := 0
					for offset < vs.blockSize {
						if offset+4 > vs.blockSize { break }
						sz := int(binary.LittleEndian.Uint32(block[offset:offset+4]))
						if sz == 0 { break }
						offset += 4 + sz
					}
					
					if offset + recordSize <= vs.blockSize {
						copy(block[offset:], recordData)
						vs.writeBlockAt(file, block, int64(i))
						
						vs.stats.BlockStatsList[i].BytesUsed += recordSize
						vs.stats.TotalBytesUsed += recordSize
						inserted = true
						break
					}
				}
			}
		}

		if inserted {
			continue
		}

		newBlock := make([]byte, vs.blockSize)
		copy(newBlock, recordData)
		
		offset := int64(vs.stats.TotalBlocks * vs.blockSize)
		_, err = file.WriteAt(newBlock, offset)
		if err != nil {
			return err
		}

		vs.stats.TotalBlocks++
		vs.stats.TotalBytesTotal += vs.blockSize
		vs.stats.TotalBytesUsed += recordSize
		
		newStat := BlockStats{
			BlockNumber: vs.stats.TotalBlocks - 1,
			BytesUsed: recordSize,
			BytesTotal: vs.blockSize,
		}
		vs.stats.BlockStatsList = append(vs.stats.BlockStatsList, newStat)
	}

	vs.calculateFinalStats()
	return nil
}

// Reorganize: Compactação física
func (vs *VariableStorage) Reorganize(filename string) (*ReorganizationReport, error) {
	statsBefore := vs.GetStats(filename)

	students, err := vs.GetAllStudents(filename)
	if err != nil {
		return nil, err
	}

	reorgFilename := filename
	if len(filename) > 4 && filename[len(filename)-4:] == ".dat" {
		reorgFilename = filename[:len(filename)-4] + "_reorg.dat"
	} else {
		reorgFilename = filename + "_reorg.dat"
	}

	tempStorage, err := NewVariableStorage(vs.blockSize)
	if err != nil {
		return nil, err
	}
	
	studentsValue := make([]entity.Student, len(students))
	for i, s := range students {
		studentsValue[i] = *s
	}
	
	err = tempStorage.WriteStudents(reorgFilename, studentsValue)
	if err != nil {
		return nil, err
	}
	
	statsAfter := tempStorage.GetStats(reorgFilename)
	
	occBefore := 0.0
	if statsBefore.TotalBlocks > 0 {
		sum := 0.0
		for _, b := range statsBefore.BlockStatsList {
			sum += b.OccupancyRate
		}
		occBefore = sum / float64(statsBefore.TotalBlocks)
	}
	
	occAfter := 0.0
	if statsAfter.TotalBlocks > 0 {
		sum := 0.0
		for _, b := range statsAfter.BlockStatsList {
			sum += b.OccupancyRate
		}
		occAfter = sum / float64(statsAfter.TotalBlocks)
	}

	report := &ReorganizationReport{
		BlocksBefore:     statsBefore.TotalBlocks,
		BlocksAfter:      statsAfter.TotalBlocks,
		OccupancyBefore:  occBefore,
		OccupancyAfter:   occAfter,
		EfficiencyBefore: statsBefore.EfficiencyRate,
		EfficiencyAfter:  statsAfter.EfficiencyRate,
		EfficiencyGain:   statsAfter.EfficiencyRate - statsBefore.EfficiencyRate,
		FreedBlocks:      statsBefore.TotalBlocks - statsAfter.TotalBlocks,
	}
	
	return report, nil
}

// FindStudentLocation e helpers
func (vs *VariableStorage) findStudentLocation(file *os.File, totalBlocks int, matricula int) (int, int, int, error) {
	for blockNum := 0; blockNum < totalBlocks; blockNum++ {
		block := make([]byte, vs.blockSize)
		_, err := file.ReadAt(block, int64(blockNum*vs.blockSize))
		if err != nil {
			continue
		}

		offset := 0
		for offset < vs.blockSize {
			if offset+4 > vs.blockSize {
				break
			}
			
			if block[offset] == 0 && block[offset+1] == 0 && block[offset+2] == 0 && block[offset+3] == 0 {
				break
			}

			totalSize := int(binary.LittleEndian.Uint32(block[offset : offset+4]))
			if totalSize == 0 || offset+4+totalSize > len(block) {
				break
			}
			
			status := block[offset+4]
			bytesConsumed := 4 + totalSize

			if status == StatusDeleted {
				offset += bytesConsumed
				continue
			}

			if offset+9 <= len(block) {
				matrBytes := block[offset+5 : offset+9]
				readMatricula := int(binary.LittleEndian.Uint32(matrBytes))
				
				if readMatricula == matricula {
					return blockNum, offset, bytesConsumed, nil
				}
			}

			offset += bytesConsumed
		}
	}
	return -1, -1, 0, fmt.Errorf("aluno não encontrado")
}

func (vs *VariableStorage) UpdateStudent(filename string, updatedStudent entity.Student) error {
	file, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("erro ao ler stats: %w", err)
	}
	totalBlocks := int(fileInfo.Size()) / vs.blockSize

	blockNum, offset, oldTotalSize, err := vs.findStudentLocation(file, totalBlocks, updatedStudent.Matricula)
	if err != nil {
		return err
	}

	newRecord := vs.serializeStudent(updatedStudent)
	newTotalSize := len(newRecord)

	block := make([]byte, vs.blockSize)
	blockStart := int64(blockNum * vs.blockSize)
	_, err = file.ReadAt(block, blockStart)
	if err != nil {
		return err
	}

	realBytesUsed := 0
	checkOffset := 0
	for checkOffset < vs.blockSize {
		if checkOffset+4 > vs.blockSize { break }
		ts := int(binary.LittleEndian.Uint32(block[checkOffset:checkOffset+4]))
		if ts == 0 { break }
		realBytesUsed += 4 + ts
		checkOffset += 4 + ts
	}

	spaceFree := vs.blockSize - realBytesUsed
	sizeDiff := newTotalSize - oldTotalSize

	canFitInBlock := (spaceFree >= sizeDiff)

	if canFitInBlock {
		newBlock := make([]byte, 0, vs.blockSize)
		newBlock = append(newBlock, block[:offset]...)
		newBlock = append(newBlock, newRecord...)
		remainingStart := offset + oldTotalSize
		if remainingStart < realBytesUsed {
			newBlock = append(newBlock, block[remainingStart:realBytesUsed]...)
		}
		
		vs.writeBlockAt(file, newBlock, int64(blockNum))
		vs.recalculateStatsFromFile(filename)
		return nil
	}

	_, err = file.WriteAt([]byte{StatusDeleted}, blockStart+int64(offset)+4)
	if err != nil {
		return err
	}
	
	file.Close() 
	
	return vs.AddStudents(filename, []entity.Student{updatedStudent})
}

func (vs *VariableStorage) DeleteStudent(filename string, matricula int) error {
	file, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("erro ao obter status do arquivo: %w", err)
	}

	totalBlocks := int(fileInfo.Size()) / vs.blockSize
	blockNum, offset, _, err := vs.findStudentLocation(file, totalBlocks, matricula)
	if err != nil {
		return fmt.Errorf("aluno não encontrado")
	}

	statusOffset := int64(blockNum*vs.blockSize) + int64(offset) + 4
	_, err = file.WriteAt([]byte{StatusDeleted}, statusOffset)
	
	if err == nil {
		vs.recalculateStatsFromFile(filename)
	}
	return err
}

func (vs *VariableStorage) writeBlockAt(file *os.File, block []byte, blockIndex int64) error {
	padded := make([]byte, vs.blockSize)
	copy(padded, block)
	_, err := file.WriteAt(padded, blockIndex*int64(vs.blockSize))
	return err
}
