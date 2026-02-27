package main

import (
	"fmt"
	"log"
	"os"
)

// Number of address registers (a1, a2)
const NUM_REGISTERS = 2 // Константа, определяющая количество регистров адреса (a1 и a2)

// CommandConstructor function type for creating commands
type CommandConstructor func(bb uint8, addr1, addr2 uint16) Command // Определение типа функции для создания команд

// PSW represents the Program Status Word
type PSW struct {
	IP           uint16 // Указатель на текущую инструкцию (Instruction Pointer)
	SignFlag     bool   // Флаг знака (отрицательное/положительное значение)
	CarryFlag    bool   // Флаг переноса (перенос из старшего бита)
	OverflowFlag bool   // Флаг переполнения (переполнение арифметической операции)
	ZeroFlag     bool   // Флаг нуля (результат операции равен нулю)
}

// Processor represents the virtual machine processor
type Processor struct {
	memory       *Memory                       // Указатель на объект памяти виртуальной машины
	psw          PSW                           // Программное слово состояния (Program Status Word)
	registers    [NUM_REGISTERS]int32          // Массив регистров для хранения значений a1 и a2
	error        bool                          // Флаг, указывающий на наличие ошибки
	stop         bool                          // Флаг, указывающий на остановку процессора
	logFile      *os.File                      // Указатель на файл для записи логов выполнения
	errorLogFile *os.File                      // Указатель на файл для записи логов ошибок
	logger       *log.Logger                   // Логгер для записи обычных логов
	errorLogger  *log.Logger                   // Логгер для записи логов ошибок
	commandMap   map[OpCode]CommandConstructor // мапа команд, связывающая коды операций с конструкторами команд
}

// NewProcessor creates a new Processor instance
func NewProcessor() (*Processor, error) {
	// Открываем файл для записи логов выполнения с флагами создания, записи и обрезки файла
	logFile, err := os.OpenFile("vm_execution.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open execution log: %v", err) // Возвращаем ошибку, если не удалось открыть файл лога
	}

	// Открываем файл для записи логов ошибок с флагами создания, записи и добавления в конец файла
	errorLogFile, err := os.OpenFile("vm_error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logFile.Close()                                             // Закрываем файл логов выполнения в случае ошибки
		return nil, fmt.Errorf("failed to open error log: %v", err) // Возвращаем ошибку при неудачном открытии файла лога ошибок
	}

	// Создаем новый экземпляр процессора с инициализацией памяти и логирования
	p := &Processor{
		memory:       NewMemory(65536),                                // Инициализация памяти размером 65536 байт
		logger:       log.New(logFile, "", log.LstdFlags),             // Инициализация логгера для выполнения
		errorLogger:  log.New(errorLogFile, "ERROR: ", log.LstdFlags), // Инициализация логгера для ошибок с префиксом "ERROR: "
		logFile:      logFile,                                         // Сохранение указателя на файл логов выполнения
		errorLogFile: errorLogFile,                                    // Сохранение указателя на файл логов ошибок
		commandMap:   make(map[OpCode]CommandConstructor),             // Инициализация мапы команд
	}

	// Инициализация мапы команд
	p.initializeCommandMap()
	return p, nil // Возвращаем указатель на созданный процессор и nil (без ошибок)
}

func (p *Processor) Run() {
	p.logMessage("Starting program execution") // Логируем начало выполнения программы
	// Цикл выполнения программы до тех пор, пока не будет установлена остановка или ошибка
	for !p.stop && !p.error {
		// Выполняем следующую инструкцию и проверяем на наличие ошибки
		if err := p.executeNextInstruction(); err != nil {
			p.logError(fmt.Sprintf("Error executing instruction: %v", err)) // Логируем ошибку выполнения инструкции
			p.error = true                                                  // Устанавливаем флаг ошибки
			break                                                           // Выходим из цикла
		}
	}
}

func (p *Processor) executeNextInstruction() error {
	currentIP := p.psw.IP // Получаем текущий адрес инструкций

	// Проверяем, является ли текущий адрес допустимым
	if !p.memory.IsValidAddress(int(currentIP)) {
		return fmt.Errorf("invalid instruction pointer: 0x%X", currentIP) // Возвращаем ошибку с недопустимым адресом
	}

	word, err := p.memory.ReadWord(int(currentIP)) // Читаем слово (инструкцию) из памяти по текущему адресу
	if err != nil {
		return fmt.Errorf("failed to read instruction: %v", err) // Возвращаем ошибку при чтении инструкции
	}

	// Проверяем, существует ли конструктор для данной операции в мапе команд
	if constructor, exists := p.commandMap[OpCode(word.Cmd.Opcode)]; exists {
		cmd := constructor(word.Cmd.BB, word.Cmd.Address1, word.Cmd.Address2) // Создаем команду на основе прочитанного слова
		if err := cmd.Execute(p); err != nil {
			return fmt.Errorf("error executing instruction at 0x%X: %v", currentIP, err) // Возвращаем ошибку выполнения команды
		}
	} else {
		return fmt.Errorf("invalid opcode at 0x%X: %d", currentIP, word.Cmd.Opcode) // Возвращаем ошибку недопустимого кода операции
	}

	// Проверяем, была ли выполнена команда STOP
	if word.Cmd.Opcode == uint8(STOP) {
		p.stop = true // Устанавливаем флаг остановки
	} else {
		// Обновляем указатель инструкций для следующей команды с учетом размера памяти
		p.psw.IP = uint16((int(currentIP) + 1) % p.memory.Size())
	}

	return nil // Возвращаем nil, если ошибок не было
}

// извлекает значение регистра по его индексу
func (p *Processor) GetRegister(index uint8) (int32, error) {
	// Проверяем, что индекс находится в допустимом диапазоне
	if index >= NUM_REGISTERS {
		return 0, fmt.Errorf("invalid register index: %d", index) // Возвращаем ошибку, если индекс недействителен
	}
	return p.registers[index], nil // Возвращаем значение регистра и nil (без ошибок)
}

// устанавливает значение регистра по его индексу
func (p *Processor) SetRegister(index uint8, value int32) error {
	// Проверяем, что индекс находится в допустимом диапазоне
	if index >= NUM_REGISTERS {
		return fmt.Errorf("invalid register index: %d", index) // Возвращаем ошибку, если индекс недействителен
	}
	p.registers[index] = value // Устанавливаем значение регистра по указанному индексу
	return nil                 // Возвращаем nil (без ошибок)
}

// флаг знака
func (p *Processor) SetSignFlag(negative bool) {
	p.psw.SignFlag = negative // Устанавливаем флаг знака в соответствии с переданным значением
}

// флаг переноса
func (p *Processor) SetCarryFlag(carry bool) {
	p.psw.CarryFlag = carry // Устанавливаем флаг переноса в соответствии с переданным значением
}

// флаг переполнения
func (p *Processor) SetOverflowFlag(overflow bool) {
	p.psw.OverflowFlag = overflow // Устанавливаем флаг переполнения в соответствии с переданным значением
}

// флаг нуля
func (p *Processor) SetZeroFlag(zero bool) {
	p.psw.ZeroFlag = zero // Устанавливаем флаг нуля в соответствии с переданным значением
}

func (p *Processor) UpdateArithmeticFlags(result int32, hasCarry, hasOverflow bool) {
	p.SetSignFlag(result < 0)      // Устанавливаем флаг знака в зависимости от результата операции
	p.SetZeroFlag(result == 0)     // Устанавливаем флаг нуля в зависимости от результата операции
	p.SetCarryFlag(hasCarry)       // Устанавливаем флаг переноса в зависимости от наличия переноса
	p.SetOverflowFlag(hasOverflow) // Устанавливаем флаг переполнения в зависимости от наличия переполнения
}

func (p *Processor) UpdateFloatFlags(result float32) {
	// Устанавливаем флаг знака в зависимости от того, отрицательный ли результат
	p.SetSignFlag(result < 0)
	// Устанавливаем флаг нуля, если результат равен нулю
	p.SetZeroFlag(result == 0)
	// Для операций с плавающей точкой флаги переноса и переполнения не имеют смысла
	p.SetCarryFlag(false)
	p.SetOverflowFlag(false)
}

func (p *Processor) GetFlags() uint16 {
	var flags uint16 // Объявляем переменную для хранения флагов
	// Проверяем, установлен ли флаг знака, и если да, устанавливаем соответствующий бит в переменной flags
	if p.psw.SignFlag {
		flags |= 0x8000
	}
	// Проверяем, установлен ли флаг переполнения
	if p.psw.OverflowFlag {
		flags |= 0x0800
	}
	// Проверяем, установлен ли флаг нуля
	if p.psw.ZeroFlag {
		flags |= 0x0400
	}
	// Проверяем, установлен ли флаг переноса
	if p.psw.CarryFlag {
		flags |= 0x0001
	}
	return flags // Возвращаем значение переменной flags
}

func (p *Processor) SetFlags(flags uint16) {
	// Устанавливаем флаг знака на основе старшего бита переменной flags
	p.psw.SignFlag = (flags & 0x8000) != 0
	// Устанавливаем флаг переполнения на основе второго старшего бита
	p.psw.OverflowFlag = (flags & 0x0800) != 0
	// Устанавливаем флаг нуля на основе третьего старшего бита
	p.psw.ZeroFlag = (flags & 0x0400) != 0
	// Устанавливаем флаг переноса на основе младшего бита
	p.psw.CarryFlag = (flags & 0x0001) != 0
}

func (p *Processor) initializeCommandMap() {
	// Инициализируем команду STOP в мапе команд
	p.commandMap[STOP] = func(bb uint8, addr1, addr2 uint16) Command { return NewHalt(bb, addr1, addr2) }
	// Инициализируем команду IADD в мапе команд
	p.commandMap[IADD] = func(bb uint8, addr1, addr2 uint16) Command { return NewAddInt(bb, addr1, addr2) }
	// Инициализируем команду ISUB в мапе команд
	p.commandMap[ISUB] = func(bb uint8, addr1, addr2 uint16) Command { return NewSubInt(bb, addr1, addr2) }
	// Инициализируем команду IMUL в мапе команд
	p.commandMap[IMUL] = func(bb uint8, addr1, addr2 uint16) Command { return NewMulInt(bb, addr1, addr2) }
	// Инициализируем команду IDIV в мапе команд
	p.commandMap[IDIV] = func(bb uint8, addr1, addr2 uint16) Command { return NewDivInt(bb, addr1, addr2) }
	// Инициализируем команду IIN в мапе команд
	p.commandMap[IIN] = func(bb uint8, addr1, addr2 uint16) Command { return NewInputInt(bb, addr1, addr2) }
	// Инициализируем команду IOUT в мапе команд
	p.commandMap[IOUT] = func(bb uint8, addr1, addr2 uint16) Command { return NewOutputInt(bb, addr1, addr2) }
	// Инициализируем команду RADD в мапе команд
	p.commandMap[RADD] = func(bb uint8, addr1, addr2 uint16) Command { return NewAddFloat(bb, addr1, addr2) }
	// Инициализируем команду RSUB в мапе команд
	p.commandMap[RSUB] = func(bb uint8, addr1, addr2 uint16) Command { return NewSubFloat(bb, addr1, addr2) }
	// Инициализируем команду RMUL в мапе команд
	p.commandMap[RMUL] = func(bb uint8, addr1, addr2 uint16) Command { return NewMulFloat(bb, addr1, addr2) }
	// Инициализируем команду RDIV в мапе команд
	p.commandMap[RDIV] = func(bb uint8, addr1, addr2 uint16) Command { return NewDivFloat(bb, addr1, addr2) }
	// Инициализируем команду RIN в мапе команд
	p.commandMap[RIN] = func(bb uint8, addr1, addr2 uint16) Command { return NewInputFloat(bb, addr1, addr2) }
	// Инициализируем команду ROUT в мапе команд
	p.commandMap[ROUT] = func(bb uint8, addr1, addr2 uint16) Command { return NewOutputFloat(bb, addr1, addr2) }
	// Инициализируем команду JZ в мапе команд
	p.commandMap[JZ] = func(bb uint8, addr1, addr2 uint16) Command { return NewJumpZero(bb, addr1, addr2) }
	// Инициализируем команду JG в мапе команд
	p.commandMap[JG] = func(bb uint8, addr1, addr2 uint16) Command { return NewJumpGreater(bb, addr1, addr2) }
	// Инициализируем команду JL в мапе команд
	p.commandMap[JL] = func(bb uint8, addr1, addr2 uint16) Command { return NewJumpLess(bb, addr1, addr2) }
	// Инициализируем команду LOAD в мапе команд
	p.commandMap[LOAD] = func(bb uint8, addr1, addr2 uint16) Command { return NewLoadRegister(bb, addr1, addr2) }
	// Инициализируем команду STORE в мапе команд
	p.commandMap[STORE] = func(bb uint8, addr1, addr2 uint16) Command { return NewStoreRegister(bb, addr1, addr2) }
	// Инициализируем команду ADDR в мапе команд
	p.commandMap[ADDR] = func(bb uint8, addr1, addr2 uint16) Command { return NewAddRegisters(bb, addr1, addr2) }
	// Инициализируем команду SUBR в мапе команд
	p.commandMap[SUBR] = func(bb uint8, addr1, addr2 uint16) Command { return NewSubtractRegisters(bb, addr1, addr2) }
	// Инициализируем команду MOVR в мапе команд
	p.commandMap[MOVR] = func(bb uint8, addr1, addr2 uint16) Command { return NewMoveRegister(bb, addr1, addr2) }
}

func (p *Processor) logMessage(message string) {
	// Проверяем наличие логгера перед записью сообщения
	if p.logger != nil {
		p.logger.Printf("%s", message) // Записываем сообщение в лог
	}
}

func (p *Processor) logError(message string) {
	// Проверяем наличие логгера ошибок перед записью сообщения об ошибке
	if p.errorLogger != nil {
		p.errorLogger.Printf("%s", message) // Записываем сообщение об ошибке в лог ошибок
	}
}

func (p *Processor) Reset(initialIP uint16) {
	// Проверяем, является ли начальный адрес допустимым
	if !p.memory.IsValidAddress(int(initialIP)) {
		// Логируем сообщение об ошибке с недопустимым адресом
		p.logMessage(fmt.Sprintf("Invalid initial IP: 0x%X", initialIP))
		p.error = true // Устанавливаем флаг ошибки
		return         // Завершаем выполнение функции
	}

	p.psw.IP = initialIP       // Устанавливаем начальный адрес инструкций
	p.psw.SignFlag = false     // Сбрасываем флаг знака
	p.psw.CarryFlag = false    // Сбрасываем флаг переноса
	p.psw.OverflowFlag = false // Сбрасываем флаг переполнения
	p.psw.ZeroFlag = false     // Сбрасываем флаг нуля
	p.error = false            // Сбрасываем флаг ошибки
	p.stop = false             // Сбрасываем флаг остановки

	// Сбрасываем регистры (a1, a2)
	p.registers[0] = 0 // Регистру a1 присваиваем 0
	p.registers[1] = 0 // Регистру a2 присваиваем 0

	// Логируем сообщение о сбросе процессора с начальным адресом инструкций
	p.logMessage(fmt.Sprintf("Processor reset with initial IP: 0x%X", initialIP))
}
func (p *Processor) Close() {
	if p.logFile != nil {
		p.logFile.Close() // Закрываем файл лога, если он открыт
	}
	if p.errorLogFile != nil {
		p.errorLogFile.Close() // Закрываем файл лога ошибок, если он открыт
	}
	if p.memory != nil {
		p.memory.Close() // Закрываем память, если она инициализирована
	}
}
