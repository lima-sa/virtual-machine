package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

type Command interface {
	Execute(*Processor) error
}

func calculateAddress(p *Processor, bb uint8, address uint16, regIndex uint8) (uint16, error) {
	// Формат BB:
	// Бит 0: Флаг модификации адреса
	// Бит 1: Флаг режима регистра

	effectiveAddr := address // Изначально эффективный адрес равен переданному адресу

	if bb&0x02 != 0 { // Если установлен флаг режима регистра
		regValue, err := p.GetRegister(regIndex) // Получаем значение регистра по индексу
		if err != nil {
			return 0, err // Возвращаем ошибку, если не удалось получить значение регистра
		}
		if bb&0x01 != 0 { // Если установлен флаг модификации адреса
			effectiveAddr = uint16(int32(address) + regValue) // Модифицируем адрес с использованием значения регистра
		} else {
			effectiveAddr = uint16(regValue) // Устанавливаем эффективный адрес равным значению регистра
		}
	} else if bb&0x01 != 0 { // Если установлен флаг модификации адреса без использования регистра
		regValue, err := p.GetRegister(0) // Используем R0 для обратной совместимости
		if err != nil {
			return 0, err // Возвращаем ошибку, если не удалось получить значение регистра R0
		}
		effectiveAddr = uint16(int32(address) + regValue) // Модифицируем адрес с использованием значения регистра R0
	}

	return effectiveAddr, nil // Возвращаем эффективный адрес и nil (без ошибок)
}

// JumpZero реализация команды JumpZero
type JumpZero struct {
	CommandData // Встраиваем структуру CommandData для хранения данных команды
}

// NewJumpZero создает новый экземпляр JumpZero с заданными параметрами
func NewJumpZero(bb uint8, addr1, addr2 uint16) *JumpZero {
	return &JumpZero{CommandData{
		Opcode:   uint8(JZ), // Устанавливаем код операции для JumpZero
		BB:       bb,        // Устанавливаем значение BB
		Address1: addr1,     // Устанавливаем первый адрес
		Address2: addr2,     // Устанавливаем второй адрес
	}}
}

// Execute выполняет команду JumpZero
func (j *JumpZero) Execute(p *Processor) error {
	if p.GetFlags() == 0 { // Проверяем флаги процессора; если они равны 0, условие выполнено
		effectiveAddr, err := calculateAddress(p, j.BB, j.Address1, 0) // Вычисляем эффективный адрес
		if err != nil {
			return err // Возвращаем ошибку, если произошла ошибка при вычислении адреса
		}
		p.psw.IP = effectiveAddr                                                      // Обновляем указатель команд (IP) процессора на эффективный адрес
		p.logMessage(fmt.Sprintf("JumpZero: Jumping to address 0x%X", effectiveAddr)) // Логируем информацию о переходе
	} else {
		p.logMessage("JumpZero: Condition not met, continuing") // Логируем информацию о том, что условие не выполнено
	}
	return nil // Возвращаем nil (без ошибок)
}

// JumpGreater реализация команды JumpGreater
type JumpGreater struct {
	CommandData // Встраиваем структуру CommandData для хранения данных команды
}

// NewJumpGreater создает новый экземпляр JumpGreater с заданными параметрами
func NewJumpGreater(bb uint8, addr1, addr2 uint16) *JumpGreater {
	return &JumpGreater{CommandData{ // Возвращаем новый объект JumpGreater, инициализируя его CommandData
		Opcode:   uint8(JG), // Устанавливаем код операции для JumpGreater
		BB:       bb,        // Устанавливаем значение BB (биты управления)
		Address1: addr1,     // Устанавливаем первый адрес для перехода
		Address2: addr2,     // Устанавливаем второй адрес (может использоваться в других командах)
	}}
}

// Execute выполняет команду JumpGreater
func (j *JumpGreater) Execute(p *Processor) error {
	if p.GetFlags() > 0 { // Проверяем флаги процессора; если они больше 0, условие выполнено
		effectiveAddr, err := calculateAddress(p, j.BB, j.Address1, 0) // Вычисляем эффективный адрес
		if err != nil {
			return err // Возвращаем ошибку, если произошла ошибка при вычислении адреса
		}
		p.psw.IP = effectiveAddr                                                         // Обновляем указатель команд (IP) процессора на эффективный адрес
		p.logMessage(fmt.Sprintf("JumpGreater: Jumping to address 0x%X", effectiveAddr)) // Логируем информацию о переходе
	} else {
		p.logMessage("JumpGreater: Condition not met, continuing") // Логируем информацию о том, что условие не выполнено
	}
	return nil // Возвращаем nil (без ошибок)
}

// JumpLess реализация команды JumpLess
type JumpLess struct {
	CommandData // Встраиваем структуру CommandData для хранения данных команды
}

// NewJumpLess создает новый экземпляр JumpLess с заданными параметрами
func NewJumpLess(bb uint8, addr1, addr2 uint16) *JumpLess {
	return &JumpLess{CommandData{ // Возвращаем новый объект JumpLess, инициализируя его CommandData
		Opcode:   uint8(JL), // Устанавливаем код операции для JumpLess
		BB:       bb,        // Устанавливаем значение BB (биты управления)
		Address1: addr1,     // Устанавливаем первый адрес для перехода
		Address2: addr2,     // Устанавливаем второй адрес (может использоваться в других командах)
	}}
}

// Execute выполняет команду JumpLess
func (j *JumpLess) Execute(p *Processor) error {
	if p.GetFlags() < 0 { // Проверяем флаги процессора; если они меньше 0, условие выполнено
		effectiveAddr, err := calculateAddress(p, j.BB, j.Address1, 0) // Вычисляем эффективный адрес
		if err != nil {
			return err // Возвращаем ошибку, если произошла ошибка при вычислении адреса
		}
		p.psw.IP = effectiveAddr                                                      // Обновляем указатель команд (IP) процессора на эффективный адрес
		p.logMessage(fmt.Sprintf("JumpLess: Jumping to address 0x%X", effectiveAddr)) // Логируем информацию о переходе
	} else {
		p.logMessage("JumpLess: Condition not met, continuing") // Логируем информацию о том, что условие не выполнено
	}
	return nil // Возвращаем nil (без ошибок)
}

// Halt command implementation
type Halt struct {
	CommandData // Встраиваем структуру CommandData для хранения данных команды
}

// NewHalt создает новый экземпляр Halt с заданными параметрами
func NewHalt(bb uint8, addr1, addr2 uint16) *Halt {
	return &Halt{CommandData{ // Возвращаем новый объект Halt, инициализируя его CommandData
		Opcode:   uint8(STOP), // Устанавливаем код операции для команды остановки
		BB:       bb,          // Устанавливаем значение BB (биты управления)
		Address1: addr1,       // Устанавливаем первый адрес (может быть использован в других командах)
		Address2: addr2,       // Устанавливаем второй адрес (может быть использован в других командах)
	}}
}

// Execute выполняет команду Halt
func (h *Halt) Execute(p *Processor) error {
	p.stop = true                            // Устанавливаем флаг остановки процессора в true
	p.logMessage("Halt: Stopping processor") // Логируем сообщение о том, что процессор останавливается
	return nil                               // Возвращаем nil (без ошибок)
}

type AddInt struct {
	CommandData // Встраиваем структуру CommandData для хранения данных команды
}

// NewAddInt создает новый экземпляр AddInt с заданными параметрами
func NewAddInt(bb uint8, addr1, addr2 uint16) *AddInt {
	return &AddInt{CommandData{ // Возвращаем новый объект AddInt, инициализируя его CommandData
		Opcode:   uint8(IADD), // Устанавливаем код операции для сложения целых чисел
		BB:       bb,          // Устанавливаем значение BB (биты управления)
		Address1: addr1,       // Устанавливаем первый адрес для первого операнда
		Address2: addr2,       // Устанавливаем второй адрес для второго операнда
	}}
}

// Execute выполняет команду AddInt
func (a *AddInt) Execute(p *Processor) error {
	// Получаем индекс регистра из младших 3 битов, если в режиме работы с регистрами
	regIndex := uint8(a.Address1 & 0x07)
	// Вычисляем адрес первого операнда
	addr1, err := calculateAddress(p, a.BB, a.Address1, regIndex)
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при вычислении адреса
	}
	// Вычисляем адрес второго операнда
	addr2, err := calculateAddress(p, a.BB, a.Address2, regIndex)
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при вычислении адреса
	}
	// Читаем первое слово из памяти по адресу addr1
	word1, err := p.memory.ReadWord(int(addr1))
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при чтении слова из памяти
	}
	// Читаем второе слово из памяти по адресу addr2
	word2, err := p.memory.ReadWord(int(addr2))
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при чтении слова из памяти
	}
	// Выполняем сложение двух целых чисел
	result := word1.D.I + word2.D.I
	word1.D.I = result // Обновляем первое слово с результатом сложения
	// Записываем обновленное слово обратно в память по адресу addr1
	err = p.memory.WriteWord(int(addr1), word1)
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при записи слова в память
	}
	// Обновляем флаги на основе результата сложения
	hasOverflow := (word1.D.I > 0 && word2.D.I > 0 && result < 0) ||
		(word1.D.I < 0 && word2.D.I < 0 && result > 0) // Проверка на переполнение
	hasCarry := uint32(word1.D.I)+uint32(word2.D.I) > uint32(0x7FFFFFFF) // Проверка на перенос
	p.UpdateArithmeticFlags(result, hasCarry, hasOverflow)               // Обновляем арифметические флаги процессора
	// Логируем информацию о выполненной операции сложения
	p.logMessage(fmt.Sprintf("AddInt: %d + %d = %d", word1.D.I, word2.D.I, result))
	return nil // Возвращаем nil (без ошибок)
}

// SubInt command implementation
type SubInt struct {
	CommandData // Встраиваем структуру CommandData для хранения данных команды
}

// NewSubInt создает новый экземпляр SubInt с заданными параметрами
func NewSubInt(bb uint8, addr1, addr2 uint16) *SubInt {
	return &SubInt{CommandData{ // Возвращаем новый объект SubInt, инициализируя его CommandData
		Opcode:   uint8(ISUB), // Устанавливаем код операции для вычитания целых чисел
		BB:       bb,          // Устанавливаем значение BB (биты управления)
		Address1: addr1,       // Устанавливаем первый адрес для первого операнда
		Address2: addr2,       // Устанавливаем второй адрес для второго операнда
	}}
}

// Execute выполняет команду SubInt
func (s *SubInt) Execute(p *Processor) error {
	regIndex := uint8(s.Address1 & 0x07) // Получаем индекс регистра из младших 3 битов адреса

	// Вычисляем адрес первого операнда
	addr1, err := calculateAddress(p, s.BB, s.Address1, regIndex)
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при вычислении адреса
	}

	// Вычисляем адрес второго операнда
	addr2, err := calculateAddress(p, s.BB, s.Address2, regIndex)
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при вычислении адреса
	}

	// Читаем первое слово из памяти по адресу addr1
	word1, err := p.memory.ReadWord(int(addr1))
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при чтении слова из памяти
	}

	// Читаем второе слово из памяти по адресу addr2
	word2, err := p.memory.ReadWord(int(addr2))
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при чтении слова из памяти
	}

	// Выполняем вычитание двух целых чисел
	result := word1.D.I - word2.D.I
	word1.D.I = result // Обновляем первое слово с результатом вычитания

	// Записываем обновленное слово обратно в память по адресу addr1
	err = p.memory.WriteWord(int(addr1), word1)
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при записи слова в память
	}

	// Обновляем флаги на основе результата вычитания
	hasOverflow := (word1.D.I > 0 && word2.D.I < 0 && result < 0) ||
		(word1.D.I < 0 && word2.D.I > 0 && result > 0) // Проверка на переполнение
	hasCarry := uint32(word1.D.I) < uint32(word2.D.I)      // Проверка на заимствование
	p.UpdateArithmeticFlags(result, hasCarry, hasOverflow) // Обновляем арифметические флаги процессора

	// Логируем информацию о выполненной операции вычитания
	p.logMessage(fmt.Sprintf("SubInt: %d - %d = %d", word1.D.I, word2.D.I, result))
	return nil // Возвращаем nil (без ошибок)
}

type MulInt struct {
	CommandData // Встраиваем структуру CommandData для хранения данных команды
}

// NewMulInt создает новый экземпляр MulInt с заданными параметрами
func NewMulInt(bb uint8, addr1, addr2 uint16) *MulInt {
	return &MulInt{CommandData{ // Возвращаем новый объект MulInt, инициализируя его CommandData
		Opcode:   uint8(IMUL), // Устанавливаем код операции для умножения целых чисел
		BB:       bb,          // Устанавливаем значение BB (биты управления)
		Address1: addr1,       // Устанавливаем первый адрес для первого операнда
		Address2: addr2,       // Устанавливаем второй адрес для второго операнда
	}}
}

// Execute выполняет команду MulInt
func (m *MulInt) Execute(p *Processor) error {
	regIndex := uint8(m.Address1 & 0x07) // Получаем индекс регистра из младших 3 битов адреса

	// Вычисляем адрес первого операнда
	addr1, err := calculateAddress(p, m.BB, m.Address1, regIndex)
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при вычислении адреса
	}

	// Вычисляем адрес второго операнда
	addr2, err := calculateAddress(p, m.BB, m.Address2, regIndex)
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при вычислении адреса
	}

	// Читаем первое слово из памяти по адресу addr1
	word1, err := p.memory.ReadWord(int(addr1))
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при чтении слова из памяти
	}

	// Читаем второе слово из памяти по адресу addr2
	word2, err := p.memory.ReadWord(int(addr2))
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при чтении слова из памяти
	}

	// Выполняем умножение двух целых чисел
	result := word1.D.I * word2.D.I
	word1.D.I = result // Обновляем первое слово с результатом умножения

	// Записываем обновленное слово обратно в память по адресу addr1
	err = p.memory.WriteWord(int(addr1), word1)
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при записи слова в память
	}

	// Обновляем флаги на основе результата умножения
	hasOverflow := result/word2.D.I != word1.D.I           // Проверка на переполнение (если результат делится на второй операнд)
	hasCarry := false                                      // Флаг переноса не имеет смысла для умножения
	p.UpdateArithmeticFlags(result, hasCarry, hasOverflow) // Обновляем арифметические флаги процессора

	// Логируем информацию о выполненной операции умножения
	p.logMessage(fmt.Sprintf("MulInt: %d * %d = %d", word1.D.I, word2.D.I, result))
	return nil // Возвращаем nil (без ошибок)
}

// DivInt command implementation
type DivInt struct {
	CommandData // Встраиваем структуру CommandData для хранения данных команды
}

// NewDivInt создает новый экземпляр DivInt с заданными параметрами
func NewDivInt(bb uint8, addr1, addr2 uint16) *DivInt {
	return &DivInt{CommandData{ // Возвращаем новый объект DivInt, инициализируя его CommandData
		Opcode:   uint8(IDIV), // Устанавливаем код операции для целочисленного деления
		BB:       bb,          // Устанавливаем значение BB (биты управления)
		Address1: addr1,       // Устанавливаем первый адрес для делимого
		Address2: addr2,       // Устанавливаем второй адрес для делителя
	}}
}

// Execute выполняет команду DivInt
func (d *DivInt) Execute(p *Processor) error {
	regIndex := uint8(d.Address1 & 0x07) // Получаем индекс регистра из младших 3 битов адреса

	// Вычисляем адрес первого операнда (делимого)
	addr1, err := calculateAddress(p, d.BB, d.Address1, regIndex)
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при вычислении адреса
	}

	// Вычисляем адрес второго операнда (делителя)
	addr2, err := calculateAddress(p, d.BB, d.Address2, regIndex)
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при вычислении адреса
	}

	// Читаем первое слово из памяти по адресу addr1 (делимое)
	word1, err := p.memory.ReadWord(int(addr1))
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при чтении слова из памяти
	}

	// Читаем второе слово из памяти по адресу addr2 (делитель)
	word2, err := p.memory.ReadWord(int(addr2))
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при чтении слова из памяти
	}

	// Проверяем делитель на ноль
	if word2.D.I == 0 {
		p.error = true                                 // Устанавливаем флаг ошибки в процессоре
		p.logMessage("DivInt: Division by zero error") // Логируем сообщение об ошибке деления на ноль
		return fmt.Errorf("division by zero")          // Возвращаем ошибку деления на ноль
	}

	// Выполняем деление двух целых чисел
	result := word1.D.I / word2.D.I
	word1.D.I = result // Обновляем первое слово с результатом деления

	// Записываем обновленное слово обратно в память по адресу addr1
	err = p.memory.WriteWord(int(addr1), word1)
	if err != nil {
		return err // Возвращаем ошибку, если произошла ошибка при записи слова в память
	}

	// Обновляем флаги на основе результата деления
	hasOverflow := false                                   // Деление не может привести к переполнению в целочисленной арифметике
	hasCarry := false                                      // Флаг переноса не имеет смысла для деления
	p.UpdateArithmeticFlags(result, hasCarry, hasOverflow) // Обновляем арифметические флаги процессора

	// Логируем информацию о выполненной операции деления
	p.logMessage(fmt.Sprintf("DivInt: %d / %d = %d", word1.D.I, word2.D.I, result))
	return nil // Возвращаем nil (без ошибок)
}

// Реализация команды AddFloat
type AddFloat struct {
	CommandData // Встраиваем структуру CommandData, содержащую данные команды
}

// Конструктор для создания нового объекта AddFloat
func NewAddFloat(bb uint8, addr1, addr2 uint16) *AddFloat {
	// Возвращаем указатель на новый объект AddFloat с заданными параметрами
	return &AddFloat{CommandData{
		Opcode:   uint8(RADD), // Устанавливаем опкод для команды RADD (сложение)
		BB:       bb,          // Устанавливаем значение bb (базовый регистр)
		Address1: addr1,       // Устанавливаем адрес первого операнда
		Address2: addr2,       // Устанавливаем адрес второго операнда
	}}
}

// Метод Execute выполняет команду AddFloat
func (a *AddFloat) Execute(p *Processor) error {
	// Получаем индекс регистра из Address1 (нижние 3 бита), если в режиме регистра
	regIndex := uint8(a.Address1 & 0x07)

	// Вычисляем адрес для первого операнда с помощью функции calculateAddress
	addr1, err := calculateAddress(p, a.BB, a.Address1, regIndex)
	if err != nil {
		return err // Возвращаем ошибку, если вычисление адреса не удалось
	}

	// Вычисляем адрес для второго операнда аналогично первому
	addr2, err := calculateAddress(p, a.BB, a.Address2, regIndex)
	if err != nil {
		return err // Возвращаем ошибку, если вычисление адреса не удалось
	}

	// Читаем слово из памяти по адресу addr1
	word1, err := p.memory.ReadWord(int(addr1))
	if err != nil {
		return err // Возвращаем ошибку, если чтение слова не удалось
	}

	// Читаем слово из памяти по адресу addr2
	word2, err := p.memory.ReadWord(int(addr2))
	if err != nil {
		return err // Возвращаем ошибку, если чтение слова не удалось
	}

	// Выполняем сложение значений с плавающей точкой
	result := word1.D.F + word2.D.F
	word1.D.F = result // Обновляем значение первого операнда с результатом сложения

	// Записываем обновленное значение обратно в память по адресу addr1
	err = p.memory.WriteWord(int(addr1), word1)
	if err != nil {
		return err // Возвращаем ошибку, если запись слова не удалась
	}

	// Обновляем флаги процессора на основе результата сложения
	p.UpdateFloatFlags(result)

	// Логируем сообщение о выполнении операции сложения
	p.logMessage(fmt.Sprintf("AddFloat: %f + %f = %f", word1.D.F, word2.D.F, result))
	return nil // Завершаем выполнение функции без ошибок
}

// Реализация команды SubFloat
type SubFloat struct {
	CommandData // Встраиваем структуру CommandData, содержащую данные команды
}

// Конструктор для создания нового объекта SubFloat
func NewSubFloat(bb uint8, addr1, addr2 uint16) *SubFloat {
	// Возвращаем указатель на новый объект SubFloat с заданными параметрами
	return &SubFloat{CommandData{
		Opcode:   uint8(RSUB), // Устанавливаем опкод для команды RSUB (вычитание)
		BB:       bb,          // Устанавливаем значение bb (базовый регистр)
		Address1: addr1,       // Устанавливаем адрес первого операнда
		Address2: addr2,       // Устанавливаем адрес второго операнда
	}}
}

// Метод Execute выполняет команду SubFloat
func (s *SubFloat) Execute(p *Processor) error {
	// Получаем индекс регистра из Address1 (нижние 3 бита), если в режиме регистра
	regIndex := uint8(s.Address1 & 0x07)

	// Вычисляем адрес для первого операнда с помощью функции calculateAddress
	addr1, err := calculateAddress(p, s.BB, s.Address1, regIndex)
	if err != nil {
		return err // Возвращаем ошибку, если вычисление адреса не удалось
	}

	// Вычисляем адрес для второго операнда аналогично первому
	addr2, err := calculateAddress(p, s.BB, s.Address2, regIndex)
	if err != nil {
		return err // Возвращаем ошибку, если вычисление адреса не удалось
	}

	// Читаем слово из памяти по адресу addr1
	word1, err := p.memory.ReadWord(int(addr1))
	if err != nil {
		return err // Возвращаем ошибку, если чтение слова не удалось
	}

	// Читаем слово из памяти по адресу addr2
	word2, err := p.memory.ReadWord(int(addr2))
	if err != nil {
		return err // Возвращаем ошибку, если чтение слова не удалось
	}

	// Выполняем вычитание значений с плавающей точкой
	result := word1.D.F - word2.D.F
	word1.D.F = result // Обновляем значение первого операнда с результатом вычитания

	// Записываем обновленное значение обратно в память по адресу addr1
	err = p.memory.WriteWord(int(addr1), word1)
	if err != nil {
		return err // Возвращаем ошибку, если запись слова не удалась
	}

	// Обновляем флаги процессора на основе результата вычитания
	p.UpdateFloatFlags(result)

	// Логируем сообщение о выполнении операции вычитания
	p.logMessage(fmt.Sprintf("SubFloat: %f - %f = %f", word1.D.F, word2.D.F, result))
	return nil // Завершаем выполнение функции без ошибок
}

// Реализация команды MulFloat
type MulFloat struct {
	CommandData // Встраиваем структуру CommandData, содержащую данные команды
}

// Конструктор для создания нового объекта MulFloat
func NewMulFloat(bb uint8, addr1, addr2 uint16) *MulFloat {
	// Возвращаем указатель на новый объект MulFloat с заданными параметрами
	return &MulFloat{CommandData{
		Opcode:   uint8(RMUL), // Устанавливаем опкод для команды RMUL (умножение)
		BB:       bb,          // Устанавливаем значение bb (базовый регистр)
		Address1: addr1,       // Устанавливаем адрес первого операнда
		Address2: addr2,       // Устанавливаем адрес второго операнда
	}}
}

// Метод Execute выполняет команду MulFloat
func (m *MulFloat) Execute(p *Processor) error {
	// Получаем индекс регистра из Address1 (нижние 3 бита), если в режиме регистра
	regIndex := uint8(m.Address1 & 0x07)

	// Вычисляем адрес для первого операнда с помощью функции calculateAddress
	addr1, err := calculateAddress(p, m.BB, m.Address1, regIndex)
	if err != nil {
		return err // Возвращаем ошибку, если вычисление адреса не удалось
	}

	// Вычисляем адрес для второго операнда аналогично первому
	addr2, err := calculateAddress(p, m.BB, m.Address2, regIndex)
	if err != nil {
		return err // Возвращаем ошибку, если вычисление адреса не удалось
	}

	// Читаем слово из памяти по адресу addr1
	word1, err := p.memory.ReadWord(int(addr1))
	if err != nil {
		return err // Возвращаем ошибку, если чтение слова не удалось
	}

	// Читаем слово из памяти по адресу addr2
	word2, err := p.memory.ReadWord(int(addr2))
	if err != nil {
		return err // Возвращаем ошибку, если чтение слова не удалось
	}

	// Выполняем умножение значений с плавающей точкой
	result := word1.D.F * word2.D.F
	word1.D.F = result // Обновляем значение первого операнда с результатом умножения

	// Записываем обновленное значение обратно в память по адресу addr1
	err = p.memory.WriteWord(int(addr1), word1)
	if err != nil {
		return err // Возвращаем ошибку, если запись слова не удалась
	}

	// Обновляем флаги процессора на основе результата умножения
	p.UpdateFloatFlags(result)

	// Логируем сообщение о выполнении операции умножения
	p.logMessage(fmt.Sprintf("MulFloat: %f * %f = %f", word1.D.F, word2.D.F, result))
	return nil // Завершаем выполнение функции без ошибок
}

// Реализация команды DivFloat
type DivFloat struct {
	CommandData // Встраиваем структуру CommandData, содержащую данные команды
}

// Конструктор для создания нового объекта DivFloat
func NewDivFloat(bb uint8, addr1, addr2 uint16) *DivFloat {
	// Возвращаем указатель на новый объект DivFloat с заданными параметрами
	return &DivFloat{CommandData{
		Opcode:   uint8(RDIV), // Устанавливаем опкод для команды RDIV (деление)
		BB:       bb,          // Устанавливаем значение bb (базовый регистр)
		Address1: addr1,       // Устанавливаем адрес первого операнда
		Address2: addr2,       // Устанавливаем адрес второго операнда
	}}
}

// Метод Execute выполняет команду DivFloat
func (d *DivFloat) Execute(p *Processor) error {
	// Получаем индекс регистра из Address1 (нижние 3 бита), если в режиме регистра
	regIndex := uint8(d.Address1 & 0x07)

	// Вычисляем адрес для первого операнда с помощью функции calculateAddress
	addr1, err := calculateAddress(p, d.BB, d.Address1, regIndex)
	if err != nil {
		return err // Возвращаем ошибку, если вычисление адреса не удалось
	}

	// Вычисляем адрес для второго операнда аналогично первому
	addr2, err := calculateAddress(p, d.BB, d.Address2, regIndex)
	if err != nil {
		return err // Возвращаем ошибку, если вычисление адреса не удалось
	}

	// Читаем слово из памяти по адресу addr1
	word1, err := p.memory.ReadWord(int(addr1))
	if err != nil {
		return err // Возвращаем ошибку, если чтение слова не удалось
	}

	// Читаем слово из памяти по адресу addr2
	word2, err := p.memory.ReadWord(int(addr2))
	if err != nil {
		return err // Возвращаем ошибку, если чтение слова не удалось
	}

	// Проверяем на деление на ноль
	if word2.D.F == 0 {
		p.error = true                                   // Устанавливаем флаг ошибки в процессоре
		p.logMessage("DivFloat: Division by zero error") // Логируем сообщение об ошибке
		return fmt.Errorf("division by zero")            // Возвращаем ошибку деления на ноль
	}

	// Выполняем деление значений с плавающей точкой
	result := word1.D.F / word2.D.F
	word1.D.F = result // Обновляем значение первого операнда с результатом деления

	// Записываем обновленное значение обратно в память по адресу addr1
	err = p.memory.WriteWord(int(addr1), word1)
	if err != nil {
		return err // Возвращаем ошибку, если запись слова не удалась
	}

	// Обновляем флаги процессора на основе результата деления
	p.UpdateFloatFlags(result)

	// Логируем сообщение о выполнении операции деления
	p.logMessage(fmt.Sprintf("DivFloat: %f / %f = %f", word1.D.F, word2.D.F, result))
	return nil // Завершаем выполнение функции без ошибок
}

// Структура InputInt, которая содержит данные команды
type InputInt struct {
	CommandData // Встраиваем структуру CommandData, содержащую данные команды
}

// Конструктор для создания нового объекта InputInt
func NewInputInt(bb uint8, addr1, addr2 uint16) *InputInt {
	// Возвращаем указатель на новый объект InputInt с заданными параметрами
	return &InputInt{CommandData{
		Opcode:   uint8(IIN), // Устанавливаем опкод для команды IIN (ввод целого числа)
		BB:       bb,         // Устанавливаем значение bb (базовый регистр)
		Address1: addr1,      // Устанавливаем адрес первого операнда
		Address2: addr2,      // Устанавливаем адрес второго операнда (не используется)
	}}
}

// Метод Execute выполняет команду InputInt
func (i *InputInt) Execute(p *Processor) error {
	scanner := bufio.NewScanner(os.Stdin)                  // Создаем новый сканер для чтения ввода с клавиатуры
	fmt.Print("Enter integer value: ")                     // Запрашиваем ввод целого числа у пользователя
	scanner.Scan()                                         // Считываем ввод пользователя
	value, err := strconv.ParseInt(scanner.Text(), 10, 32) // Преобразуем введенное значение в целое число
	if err != nil {
		return fmt.Errorf("invalid integer input: %v", err) // Возвращаем ошибку, если ввод некорректен
	}

	// Получаем индекс регистра из Address1 (нижние 3 бита), если в режиме регистра
	regIndex := uint8(i.Address1 & 0x07)

	// Вычисляем адрес для записи значения с помощью функции calculateAddress
	addr1, err := calculateAddress(p, i.BB, i.Address1, regIndex)
	if err != nil {
		return err // Возвращаем ошибку, если вычисление адреса не удалось
	}

	// Создаем новое слово с данными целого числа
	word := Word{D: Data{I: int32(value)}}

	// Записываем слово в память по вычисленному адресу
	err = p.memory.WriteWord(int(addr1), word)
	if err != nil {
		return err // Возвращаем ошибку, если запись слова не удалась
	}

	// Логируем сообщение о введенном значении
	p.logMessage(fmt.Sprintf("InputInt: Read value %d", value))
	return nil // Завершаем выполнение функции без ошибок
}

// Структура OutputInt, которая содержит данные команды
type OutputInt struct {
	CommandData // Встраиваем структуру CommandData, содержащую данные команды
}

// Конструктор для создания нового объекта OutputInt
func NewOutputInt(bb uint8, addr1, addr2 uint16) *OutputInt {
	// Возвращаем указатель на новый объект OutputInt с заданными параметрами
	return &OutputInt{CommandData{
		Opcode:   uint8(IOUT), // Устанавливаем опкод для команды IOUT (вывод целого числа)
		BB:       bb,          // Устанавливаем значение bb (базовый регистр)
		Address1: addr1,       // Устанавливаем адрес первого операнда
		Address2: addr2,       // Устанавливаем адрес второго операнда (не используется)
	}}
}

// Метод Execute выполняет команду OutputInt
func (o *OutputInt) Execute(p *Processor) error {
	// Получаем индекс регистра из Address1 (нижние 3 бита), если в режиме регистра
	regIndex := uint8(o.Address1 & 0x07)

	// Вычисляем адрес для чтения значения с помощью функции calculateAddress
	addr1, err := calculateAddress(p, o.BB, o.Address1, regIndex)
	if err != nil {
		return err // Возвращаем ошибку, если вычисление адреса не удалось
	}

	// Читаем слово из памяти по адресу addr1
	word, err := p.memory.ReadWord(int(addr1))
	if err != nil {
		return err // Возвращаем ошибку, если чтение слова не удалось
	}

	// Выводим значение на экран
	fmt.Printf("Output: %dn", word.D.I)

	// Логируем сообщение о выведенном значении
	p.logMessage(fmt.Sprintf("OutputInt: Value %d", word.D.I))
	return nil // Завершаем выполнение функции без ошибок
}

// Структура InputFloat, которая содержит данные команды
type InputFloat struct {
	CommandData // Встраиваем структуру CommandData, содержащую данные команды
}

// Конструктор для создания нового объекта InputFloat
func NewInputFloat(bb uint8, addr1, addr2 uint16) *InputFloat {
	// Возвращаем указатель на новый объект InputFloat с заданными параметрами
	return &InputFloat{CommandData{
		Opcode:   uint8(RIN), // Устанавливаем опкод для команды RIN (ввод числа с плавающей точкой)
		BB:       bb,         // Устанавливаем значение bb (базовый регистр)
		Address1: addr1,      // Устанавливаем адрес первого операнда
		Address2: addr2,      // Устанавливаем адрес второго операнда (не используется)
	}}
}

// Метод Execute выполняет команду InputFloat
func (i *InputFloat) Execute(p *Processor) error {
	scanner := bufio.NewScanner(os.Stdin)                // Создаем новый сканер для чтения ввода с клавиатуры
	fmt.Print("Enter float value: ")                     // Запрашиваем ввод числа с плавающей точкой у пользователя
	scanner.Scan()                                       // Считываем ввод пользователя
	value, err := strconv.ParseFloat(scanner.Text(), 32) // Преобразуем введенное значение в число с плавающей точкой (32 бита)
	if err != nil {
		return fmt.Errorf("invalid float input: %v", err) // Возвращаем ошибку, если ввод некорректен
	}

	// Получаем индекс регистра из Address1 (нижние 3 бита), если в режиме регистра
	regIndex := uint8(i.Address1 & 0x07)

	// Вычисляем адрес для записи значения с помощью функции calculateAddress
	addr1, err := calculateAddress(p, i.BB, i.Address1, regIndex)
	if err != nil {
		return err // Возвращаем ошибку, если вычисление адреса не удалось
	}

	// Создаем новое слово с данными числа с плавающей точкой
	word := Word{D: Data{F: float32(value)}}   // Преобразуем значение в float32 и оборачиваем в структуру Word
	err = p.memory.WriteWord(int(addr1), word) // Записываем слово в память по вычисленному адресу
	if err != nil {
		return err // Возвращаем ошибку, если запись слова не удалась
	}

	// Логируем сообщение о введенном значении
	p.logMessage(fmt.Sprintf("InputFloat: Read value %f", value))
	return nil // Завершаем выполнение функции без ошибок
}

// Структура OutputFloat, которая содержит данные команды
type OutputFloat struct {
	CommandData // Встраиваем структуру CommandData, содержащую данные команды
}

// Конструктор для создания нового объекта OutputFloat
func NewOutputFloat(bb uint8, addr1, addr2 uint16) *OutputFloat {
	// Возвращаем указатель на новый объект OutputFloat с заданными параметрами
	return &OutputFloat{CommandData{
		Opcode:   uint8(ROUT), // Устанавливаем опкод для команды ROUT (вывод числа с плавающей точкой)
		BB:       bb,          // Устанавливаем значение bb (базовый регистр)
		Address1: addr1,       // Устанавливаем адрес первого операнда
		Address2: addr2,       // Устанавливаем адрес второго операнда (не используется)
	}}
}

// Метод Execute выполняет команду OutputFloat
func (o *OutputFloat) Execute(p *Processor) error {
	// Получаем индекс регистра из Address1 (нижние 3 бита), если в режиме регистра
	regIndex := uint8(o.Address1 & 0x07)

	// Вычисляем адрес для чтения значения с помощью функции calculateAddress
	addr1, err := calculateAddress(p, o.BB, o.Address1, regIndex)
	if err != nil {
		return err // Возвращаем ошибку, если вычисление адреса не удалось
	}

	// Читаем слово из памяти по адресу addr1
	word, err := p.memory.ReadWord(int(addr1))
	if err != nil {
		return err // Возвращаем ошибку, если чтение слова не удалось
	}

	// Выводим значение на экран
	fmt.Printf("Output: %fn", word.D.F)

	// Логируем сообщение о выведенном значении
	p.logMessage(fmt.Sprintf("OutputFloat: Value %f", word.D.F))
	return nil // Завершаем выполнение функции без ошибок
}

// LoadRegister command implementation
type LoadRegister struct {
	CommandData // Встраиваемый тип CommandData, который содержит общие данные команды
}

// NewLoadRegister создает новый экземпляр LoadRegister с заданными параметрами
func NewLoadRegister(bb uint8, addr1, addr2 uint16) *LoadRegister {
	return &LoadRegister{CommandData{
		Opcode:   uint8(LOAD), // Устанавливаем код операции (Opcode) для команды LOAD
		BB:       bb,          // Устанавливаем значение bb (возможно, это флаг или дополнительный байт)
		Address1: addr1,       // Устанавливаем первый адрес (Address1)
		Address2: addr2,       // Устанавливаем второй адрес (Address2)
	}}
}

// Execute выполняет команду LoadRegister, загружая значение из памяти в регистр
func (l *LoadRegister) Execute(p *Processor) error {
	// Получаем индекс регистра из Address1 (используем младшие 3 бита)
	regIndex := uint8(l.Address1 & 0x07)

	// Загружаем слово из памяти по адресу Address2
	word, err := p.memory.ReadWord(int(l.Address2))
	if err != nil {
		return err // Возвращаем ошибку, если чтение из памяти не удалось
	}

	// Устанавливаем значение загруженного слова в указанный регистр
	err = p.SetRegister(regIndex, word.D.I)
	if err != nil {
		return err // Возвращаем ошибку, если установка регистра не удалась
	}

	// Логируем сообщение о загрузке значения в регистр
	p.logMessage(fmt.Sprintf("LoadRegister: R%d = %d", regIndex, word.D.I))
	return nil // Возвращаем nil, указывая на успешное выполнение команды
}

// StoreRegister command implementation
type StoreRegister struct {
	CommandData // Встраиваемый тип CommandData, который содержит общие данные команды
}

// NewStoreRegister создает новый экземпляр StoreRegister с заданными параметрами
func NewStoreRegister(bb uint8, addr1, addr2 uint16) *StoreRegister {
	return &StoreRegister{CommandData{
		Opcode:   uint8(STORE), // Устанавливаем код операции (Opcode) для команды STORE
		BB:       bb,           // Устанавливаем значение bb (возможно, это флаг или дополнительный байт)
		Address1: addr1,        // Устанавливаем адрес для записи (Address1)
		Address2: addr2,        // Устанавливаем адрес для получения индекса регистра (Address2)
	}}
}

// Execute выполняет команду StoreRegister, сохраняя значение из регистра в память
func (s *StoreRegister) Execute(p *Processor) error {
	// Получаем индекс регистра из Address2 (используем младшие 3 бита)
	regIndex := uint8(s.Address2 & 0x07)

	// Получаем значение из указанного регистра
	value, err := p.GetRegister(regIndex)
	if err != nil {
		return err // Возвращаем ошибку, если получение значения из регистра не удалось
	}

	// Создаем объект Word с загружаемым значением
	word := Word{D: Data{I: value}}

	// Записываем значение в память по адресу Address1
	err = p.memory.WriteWord(int(s.Address1), word)
	if err != nil {
		return err // Возвращаем ошибку, если запись в память не удалась
	}

	// Логируем сообщение о сохранении значения в памяти
	p.logMessage(fmt.Sprintf("StoreRegister: [0x%X] = R%d (%d)", s.Address1, regIndex, value))
	return nil // Возвращаем nil, указывая на успешное выполнение команды
}

// AddRegisters command implementation
type AddRegisters struct {
	CommandData // Встраиваемый тип CommandData, который содержит общие данные команды
}

// NewAddRegisters создает новый экземпляр AddRegisters с заданными параметрами
func NewAddRegisters(bb uint8, addr1, addr2 uint16) *AddRegisters {
	return &AddRegisters{CommandData{
		Opcode:   uint8(ADDR), // Устанавливаем код операции (Opcode) для команды ADDR
		BB:       bb,          // Устанавливаем значение bb (возможно, это флаг или дополнительный байт)
		Address1: addr1,       // Устанавливаем адрес для назначения результата (Address1)
		Address2: addr2,       // Устанавливаем адрес источника (Address2)
	}}
}

// Execute выполняет команду AddRegisters, складывая значения из двух регистров
func (a *AddRegisters) Execute(p *Processor) error {
	// Получаем индексы регистров из адресов (используем младшие 3 бита)
	regDest := uint8(a.Address1 & 0x07) // Индекс регистра назначения
	regSrc := uint8(a.Address2 & 0x07)  // Индекс регистра источника

	// Получаем значение из регистра назначения
	val1, err := p.GetRegister(regDest)
	if err != nil {
		return err // Возвращаем ошибку, если получение значения из регистра не удалось
	}

	// Получаем значение из регистра источника
	val2, err := p.GetRegister(regSrc)
	if err != nil {
		return err // Возвращаем ошибку, если получение значения из регистра не удалось
	}

	// Складываем два значения
	result := val1 + val2

	// Устанавливаем результат в регистр назначения
	err = p.SetRegister(regDest, result)
	if err != nil {
		return err // Возвращаем ошибку, если установка значения в регистр не удалась
	}

	// Обновляем флаги арифметических операций
	hasOverflow := (val1 > 0 && val2 > 0 && result < 0) || // Проверка на переполнение
		(val1 < 0 && val2 < 0 && result > 0) // Проверка на переполнение при отрицательных значениях
	hasCarry := uint32(val1)+uint32(val2) > uint32(0x7FFFFFFF) // Проверка на перенос

	// Обновляем флаги в процессоре
	p.UpdateArithmeticFlags(result, hasCarry, hasOverflow)

	// Логируем сообщение о результате сложения
	p.logMessage(fmt.Sprintf("AddRegisters: R%d = R%d + R%d (%d = %d + %d)",
		regDest, regDest, regSrc, result, val1, val2))
	return nil // Возвращаем nil, указывая на успешное выполнение команды
}

// SubtractRegisters command implementation
type SubtractRegisters struct {
	CommandData // Встраиваемый тип CommandData, который содержит общие данные команды
}

// NewSubtractRegisters создает новый экземпляр SubtractRegisters с заданными параметрами
func NewSubtractRegisters(bb uint8, addr1, addr2 uint16) *SubtractRegisters {
	return &SubtractRegisters{CommandData{
		Opcode:   uint8(SUBR), // Устанавливаем код операции (Opcode) для команды SUBR
		BB:       bb,          // Устанавливаем значение bb (возможно, это флаг или дополнительный байт)
		Address1: addr1,       // Устанавливаем адрес для назначения результата (Address1)
		Address2: addr2,       // Устанавливаем адрес источника (Address2)
	}}
}

// Execute выполняет команду SubtractRegisters, вычитая значения из двух регистров
func (s *SubtractRegisters) Execute(p *Processor) error {
	// Получаем индексы регистров из адресов (используем младшие 3 бита)
	regDest := uint8(s.Address1 & 0x07) // Индекс регистра назначения
	regSrc := uint8(s.Address2 & 0x07)  // Индекс регистра источника

	// Получаем значение из регистра назначения
	val1, err := p.GetRegister(regDest)
	if err != nil {
		return err // Возвращаем ошибку, если получение значения из регистра не удалось
	}

	// Получаем значение из регистра источника
	val2, err := p.GetRegister(regSrc)
	if err != nil {
		return err // Возвращаем ошибку, если получение значения из регистра не удалось
	}

	// Вычитаем значение из регистра источника из значения регистра назначения
	result := val1 - val2

	// Устанавливаем результат в регистр назначения
	err = p.SetRegister(regDest, result)
	if err != nil {
		return err // Возвращаем ошибку, если установка значения в регистр не удалась
	}

	// Обновляем флаги арифметических операций
	hasOverflow := (val1 > 0 && val2 < 0 && result < 0) || // Проверка на переполнение
		(val1 < 0 && val2 > 0 && result > 0) // Проверка на переполнение при различных знаках
	hasCarry := uint32(val1) < uint32(val2) // Проверка на заимствование

	// Обновляем флаги в процессоре
	p.UpdateArithmeticFlags(result, hasCarry, hasOverflow)

	// Логируем сообщение о результате вычитания
	p.logMessage(fmt.Sprintf("SubtractRegisters: R%d = R%d - R%d (%d = %d - %d)",
		regDest, regDest, regSrc, result, val1, val2))
	return nil // Возвращаем nil, указывая на успешное выполнение команды
}

// MoveRegister command implementation
type MoveRegister struct {
	CommandData
}

func NewMoveRegister(bb uint8, addr1, addr2 uint16) *MoveRegister {
	return &MoveRegister{CommandData{
		Opcode:   uint8(MOVR),
		BB:       bb,
		Address1: addr1,
		Address2: addr2,
	}}
}

func (m *MoveRegister) Execute(p *Processor) error {
	// Get register indices from addresses (lower 3 bits)
	regDest := uint8(m.Address1 & 0x07)
	regSrc := uint8(m.Address2 & 0x07)

	// Move value from one register to another
	value, err := p.GetRegister(regSrc)
	if err != nil {
		return err
	}

	err = p.SetRegister(regDest, value)
	if err != nil {
		return err
	}

	p.logMessage(fmt.Sprintf("MoveRegister: R%d = R%d (%d)", regDest, regSrc, value))
	return nil
}
