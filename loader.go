package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// CommandError представляет ошибку в парсинге команды
type CommandError struct {
	LineNumber int    // Номер строки, в которой произошла ошибка
	Line       string // Содержимое строки, в которой произошла ошибка
	Message    string // Сообщение об ошибке
}

// Error реализует интерфейс error для CommandError, возвращая строку с информацией об ошибке
func (e *CommandError) Error() string {
	return fmt.Sprintf("Line %d: %snContent: %s", e.LineNumber, e.Message, e.Line) // Форматирует сообщение об ошибке с указанием номера строки, сообщения и содержимого строки
}

// isValidOpcode проверяет, является ли опкод допустимым
func isValidOpcode(opcode uint64) bool {
	return opcode <= 0x45 // Возвращает true, если опкод меньше или равен 0x45 (максимально допустимый опкод)
}

// isValidBB проверяет, является ли значение BB допустимым (2 бита)
func isValidBB(bb uint64) bool {
	return bb <= 0x03 // Возвращает true, если значение BB меньше или равно 0x03 (максимальное значение для 2 бит)
}

// isValidAddress проверяет, является ли адрес допустимым
func isValidAddress(addr uint64, memory *Memory) bool {
	return int(addr) < memory.Size() // Возвращает true, если адрес меньше размера памяти (проверка на допустимость адреса)
}

// readProgramFromFile читает программу из файла и загружает ее в память
func readProgramFromFile(file *os.File, memory *Memory) (uint16, error) {
	scanner := bufio.NewScanner(file) // Создает новый сканер для чтения из файла
	var address int                   // Переменная для хранения текущего адреса
	var initialIP uint16              // Переменная для хранения начального значения IP (индикатор программы)
	var entryPointSet bool            // Флаг, указывающий, установлен ли начальный адрес
	lineNumber := 0                   // Инициализация счетчика строк

	// Чтение файла построчно
	for scanner.Scan() {
		lineNumber++           // Увеличиваем номер строки
		line := scanner.Text() // Читаем текущую строку

		// Удаляем встроенные комментарии
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = line[:idx] // Обрезаем строку до комментария
		}

		// Убираем пробелы и пропускаем пустые строки
		line = strings.TrimSpace(line)
		if line == "" {
			continue // Переходим к следующей итерации цикла, если строка пустая
		}

		fields := strings.Fields(line) // Разделяем строку на поля по пробелам
		if len(fields) < 1 {
			continue // Пропускаем строки без команд
		}

		command := strings.ToLower(fields[0]) // Приводим команду к нижнему регистру для нечувствительности к регистру
		switch command {
		case "a": // Обработка команды установки адреса
			if len(fields) < 2 {
				return 0, &CommandError{ // Если не указано значение адреса, возвращаем ошибку
					LineNumber: lineNumber,
					Line:       line,
					Message:    "address command requires a value",
				}
			}
			addr, err := strconv.ParseInt(fields[1], 16, 32) // Парсим значение адреса из шестнадцатеричного формата
			if err != nil {
				return 0, &CommandError{ // Если произошла ошибка парсинга, возвращаем ошибку
					LineNumber: lineNumber,
					Line:       line,
					Message:    fmt.Sprintf("invalid address format: %v", err),
				}
			}
			if !memory.IsValidAddress(int(addr)) { // Проверяем, является ли адрес допустимым в пределах памяти
				return 0, &CommandError{ // Если адрес вне допустимого диапазона, возвращаем ошибку
					LineNumber: lineNumber,
					Line:       line,
					Message:    fmt.Sprintf("address 0x%X is out of valid range [0-%d]", addr, memory.Size()-1),
				}
			}
			address = int(addr) // Устанавливаем текущий адрес
		case "e": // Устанавливаем начальный IP (индикатор программы)
			if len(fields) < 2 { // Проверяем, указано ли значение для начального IP
				return 0, &CommandError{ // Если нет, возвращаем ошибку
					LineNumber: lineNumber,                             // Номер строки с ошибкой
					Line:       line,                                   // Содержимое строки
					Message:    "entry point command requires a value", // Сообщение об ошибке
				}
			}
			ip, err := strconv.ParseInt(fields[1], 16, 16) // Парсим значение начального IP из шестнадцатеричного формата
			if err != nil {                                // Проверяем, произошла ли ошибка при парсинге
				return 0, &CommandError{ // Если да, возвращаем ошибку
					LineNumber: lineNumber,                                        // Номер строки с ошибкой
					Line:       line,                                              // Содержимое строки
					Message:    fmt.Sprintf("invalid initial IP format: %v", err), // Сообщение об ошибке
				}
			}
			if !memory.IsValidAddress(int(ip)) { // Проверяем, является ли адрес начального IP допустимым в пределах памяти
				return 0, &CommandError{ // Если нет, возвращаем ошибку
					LineNumber: lineNumber,                                                                        // Номер строки с ошибкой
					Line:       line,                                                                              // Содержимое строки
					Message:    fmt.Sprintf("entry point 0x%X is out of valid range [0-%d]", ip, memory.Size()-1), // Сообщение об ошибке с указанием диапазона
				}
			}
			initialIP = uint16(ip) // Устанавливаем начальный IP как значение переменной initialIP
			entryPointSet = true   // Устанавливаем флаг entryPointSet в true, указывая, что начальный IP установлен

		case "i": // Обработка команды установки целочисленного значения
			if len(fields) < 2 { // Проверяем, указано ли значение для целочисленной команды
				return 0, &CommandError{ // Если нет, возвращаем ошибку
					LineNumber: lineNumber,                         // Номер строки с ошибкой
					Line:       line,                               // Содержимое строки
					Message:    "integer command requires a value", // Сообщение об ошибке
				}
			}
			value, err := strconv.ParseInt(fields[1], 10, 32) // Парсим значение как целое число в десятичном формате
			if err != nil {                                   // Проверяем, произошла ли ошибка при парсинге
				return 0, &CommandError{ // Если да, возвращаем ошибку
					LineNumber: lineNumber,                                     // Номер строки с ошибкой
					Line:       line,                                           // Содержимое строки
					Message:    fmt.Sprintf("invalid integer format: %v", err), // Сообщение об ошибке
				}
			}
			word := Word{D: Data{I: int32(value)}}                  // Создаем объект Word с целочисленным значением
			if err := memory.WriteWord(address, word); err != nil { // Пытаемся записать слово в память по текущему адресу
				return 0, &CommandError{ // Если произошла ошибка записи, возвращаем ошибку
					LineNumber: lineNumber,                                                // Номер строки с ошибкой
					Line:       line,                                                      // Содержимое строки
					Message:    fmt.Sprintf("failed to write integer to memory: %v", err), // Сообщение об ошибке
				}
			}
			address++ // Увеличиваем адрес для следующей записи в памяти
		case "r": // Обработка команды для записи значения с плавающей запятой
			if len(fields) < 2 { // Проверяем, указано ли значение для команды с плавающей запятой
				return 0, &CommandError{ // Если значение отсутствует, возвращаем ошибку
					LineNumber: lineNumber,                       // Номер строки с ошибкой
					Line:       line,                             // Содержимое строки
					Message:    "float command requires a value", // Сообщение об ошибке
				}
			}
			value, err := strconv.ParseFloat(fields[1], 32) // Парсим значение как число с плавающей запятой (32 бита)
			if err != nil {                                 // Проверяем, произошла ли ошибка при парсинге
				return 0, &CommandError{ // Если ошибка есть, возвращаем её
					LineNumber: lineNumber,                                   // Номер строки с ошибкой
					Line:       line,                                         // Содержимое строки
					Message:    fmt.Sprintf("invalid float format: %v", err), // Сообщение об ошибке с описанием проблемы
				}
			}
			word := Word{D: Data{F: float32(value)}}                // Создаем объект Word с плавающим значением, преобразованным в float32
			if err := memory.WriteWord(address, word); err != nil { // Пытаемся записать слово в память по текущему адресу
				return 0, &CommandError{ // Если произошла ошибка записи, возвращаем её
					LineNumber: lineNumber,                                              // Номер строки с ошибкой
					Line:       line,                                                    // Содержимое строки
					Message:    fmt.Sprintf("failed to write float to memory: %v", err), // Сообщение об ошибке с описанием проблемы
				}
			}
			address++ // Увеличиваем адрес для следующей записи в памяти
		case "k": // Обработка команды "k"
			if len(fields) < 5 { // Проверяем, достаточно ли параметров (минимум 4 параметра)
				return 0, &CommandError{ // Если параметров недостаточно, возвращаем ошибку
					LineNumber: lineNumber,                                                                                     // Номер строки с ошибкой
					Line:       line,                                                                                           // Содержимое строки
					Message:    fmt.Sprintf("command requires 4 parameters (opcode, bb, addr1, addr2), got %d", len(fields)-1), // Сообщение об ошибке с количеством переданных параметров
				}
			}

			// Парсинг операционного кода (opcode)
			opcode, err := strconv.ParseUint(fields[1], 16, 8) // Преобразуем второй параметр из шестнадцатеричного формата в 8-битное целое число
			if err != nil {                                    // Проверяем, произошла ли ошибка при парсинге
				return 0, &CommandError{ // Если ошибка есть, возвращаем её
					LineNumber: lineNumber,                                    // Номер строки с ошибкой
					Line:       line,                                          // Содержимое строки
					Message:    fmt.Sprintf("invalid opcode format: %v", err), // Сообщение об ошибке с описанием проблемы
				}
			}

			if !isValidOpcode(opcode) { // Проверяем, является ли код операции допустимым
				return 0, &CommandError{ // Если код недопустим, возвращаем ошибку
					LineNumber: lineNumber,                                                                 // Номер строки с ошибкой
					Line:       line,                                                                       // Содержимое строки
					Message:    fmt.Sprintf("opcode value 0x%X is out of valid range [0x00-0x45]", opcode), // Сообщение об ошибке с диапазоном допустимых значений
				}
			}

			// Парсинг значения BB
			bb, err := strconv.ParseUint(fields[2], 16, 8) // Преобразуем третий параметр из шестнадцатеричного формата в 8-битное целое число
			if err != nil {                                // Проверяем, произошла ли ошибка при парсинге
				return 0, &CommandError{ // Если ошибка есть, возвращаем её
					LineNumber: lineNumber,                                // Номер строки с ошибкой
					Line:       line,                                      // Содержимое строки
					Message:    fmt.Sprintf("invalid bb format: %v", err), // Сообщение об ошибке с описанием проблемы
				}
			}

			if !isValidBB(bb) { // Проверяем, является ли значение BB допустимым
				return 0, &CommandError{ // Если значение недопустимо, возвращаем ошибку
					LineNumber: lineNumber,                                                       // Номер строки с ошибкой
					Line:       line,                                                             // Содержимое строки
					Message:    fmt.Sprintf("BB value 0x%X exceeds 2-bit range [0x00-0x03]", bb), // Сообщение об ошибке с диапазоном допустимых значений
				}
			}

			// Парсинг адресов
			addr1, err := strconv.ParseUint(fields[3], 16, 16) // Преобразуем четвертый параметр из шестнадцатеричного формата в 16-битное целое число
			if err != nil {                                    // Проверяем, произошла ли ошибка при парсинге
				return 0, &CommandError{ // Если ошибка есть, возвращаем её
					LineNumber: lineNumber,                                   // Номер строки с ошибкой
					Line:       line,                                         // Содержимое строки
					Message:    fmt.Sprintf("invalid addr1 format: %v", err), // Сообщение об ошибке с описанием проблемы
				}
			}

			if !isValidAddress(addr1, memory) { // Проверяем, является ли адрес addr1 допустимым в пределах памяти
				return 0, &CommandError{ // Если адрес недопустим, возвращаем ошибку
					LineNumber: lineNumber,                                                                     // Номер строки с ошибкой
					Line:       line,                                                                           // Содержимое строки
					Message:    fmt.Sprintf("addr1 0x%X is out of valid range [0-%d]", addr1, memory.Size()-1), // Сообщение об ошибке с диапазоном допустимых значений
				}
			}

			addr2, err := strconv.ParseUint(fields[4], 16, 16) // Преобразуем пятый параметр из шестнадцатеричного формата в 16-битное целое число
			if err != nil {                                    // Проверяем, произошла ли ошибка при парсинге
				return 0, &CommandError{ // Если ошибка есть, возвращаем её
					LineNumber: lineNumber,                                   // Номер строки с ошибкой
					Line:       line,                                         // Содержимое строки
					Message:    fmt.Sprintf("invalid addr2 format: %v", err), // Сообщение об ошибке с описанием проблемы
				}
			}

			if !isValidAddress(addr2, memory) { // Проверяем, является ли адрес addr2 допустимым в пределах памяти
				return 0, &CommandError{ // Если адрес недопустим, возвращаем ошибку
					LineNumber: lineNumber,                                                                     // Номер строки с ошибкой
					Line:       line,                                                                           // Содержимое строки
					Message:    fmt.Sprintf("addr2 0x%X is out of valid range [0-%d]", addr2, memory.Size()-1), // Сообщение об ошибке с диапазоном допустимых значений
				}
			}

			word := Word{ // Создаем объект Word для записи в память
				Cmd: CommandData{ // Заполняем данные команды
					Opcode:   uint8(opcode), // Устанавливаем код операции как uint8
					BB:       uint8(bb),     // Устанавливаем значение BB как uint8
					Address1: uint16(addr1), // Устанавливаем первый адрес как uint16
					Address2: uint16(addr2), // Устанавливаем второй адрес как uint16
				},
			}

			if err := memory.WriteWord(address, word); err != nil { // Пытаемся записать слово в память по текущему адресу
				return 0, &CommandError{ // Если произошла ошибка записи, возвращаем её
					LineNumber: lineNumber,                                                // Номер строки с ошибкой
					Line:       line,                                                      // Содержимое строки
					Message:    fmt.Sprintf("failed to write command to memory: %v", err), // Сообщение об ошибке с описанием проблемы записи в память
				}
			}
			address++ // Увеличиваем адрес для следующей записи в памяти
		case "s": // Обработка команды "s", которая обозначает конец программы
			if !entryPointSet {
				return 0, &CommandError{
					LineNumber: lineNumber,
					Line:       line,
					Message:    "program ended without setting entry point (e command)",
				}
			}
			return initialIP, nil

		default:
			return 0, &CommandError{
				LineNumber: lineNumber,
				Line:       line,
				Message:    fmt.Sprintf("unknown command type: %s", fields[0]),
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("error reading file: %v", err)
	}

	return 0, &CommandError{
		LineNumber: lineNumber,
		Line:       "",
		Message:    "program file ended without 's' command",
	}
}
