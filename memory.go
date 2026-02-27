package main

import (
	"encoding/binary"
	"unsafe" // Added import for unsafe package
)

// Memory представляет память виртуальной машины
type Memory struct {
	data        []byte // Массив байтов для хранения данных памяти
	size        int    // Размер памяти в байтах
	errorCount  int    // Счетчик ошибок при доступе к памяти
	accessCount int    // Счетчик обращений к памяти
	initialized bool   // Флаг, указывающий, инициализирована ли память
}

// NewMemory создает новый экземпляр Memory с заданным размером
func NewMemory(size int) *Memory {
	// Проверяем, является ли размер памяти допустимым (больше 0)
	if size <= 0 {
		panic("attempted to create memory with invalid size") // Вызываем панику при недопустимом размере
	}
	return &Memory{
		data:        make([]byte, size), // Инициализируем массив байтов заданного размера
		size:        size,               // Устанавливаем размер памяти
		initialized: true,               // Устанавливаем флаг инициализации в true
	}
}

// Size возвращает размер памяти в байтах
func (m *Memory) Size() int {
	return m.size // Возвращаем размер памяти
}

// IsValidAddress проверяет, находится ли адрес в пределах допустимого диапазона
func (m *Memory) IsValidAddress(address int) bool {
	return address >= 0 && address < m.size // Проверяем, что адрес не отрицательный и меньше размера памяти
}

// isWordAligned проверяет, выровнен ли адрес по границе слова (4 байта)
func (m *Memory) isWordAligned(address int) bool {
	return address%4 == 0 // Проверяем, делится ли адрес на 4 без остатка
}

// WriteWord записывает слово в память по заданному адресу с проверкой границ
func (m *Memory) WriteWord(address int, word Word) error {
	// Преобразуем слово в массив байтов
	var bytes [4]byte
	if word.Cmd.Opcode > 0 { // Если это команда
		binary.LittleEndian.PutUint32(bytes[:], uint32( // Преобразуем команду в байты
			uint32(word.Cmd.Opcode)<<24| // Сдвигаем код операции на 24 бита
				uint32(word.Cmd.BB)<<22| // Сдвигаем BB на 22 бита
				uint32(word.Cmd.Address1)<<10| // Сдвигаем Address1 на 10 бит
				uint32(word.Cmd.Address2))) // Добавляем Address2
	} else { // Если это данные
		binary.LittleEndian.PutUint32(bytes[:], *(*uint32)(unsafe.Pointer(&word.D.I))) // Преобразуем данные в байты
	}

	// Записываем байты в память
	copy(m.data[address:address+4], bytes[:]) // Копируем 4 байта по указанному адресу
	m.accessCount++                           // Увеличиваем счетчик обращений к памяти
	return nil                                // Возвращаем nil, если ошибок не было
}

// ReadWord читает слово из памяти по заданному адресу с проверкой границ
func (m *Memory) ReadWord(address int) (Word, error) {
	// Читаем 4 байта из памяти
	var bytes [4]byte
	copy(bytes[:], m.data[address:address+4]) // Копируем 4 байта из памяти по указанному адресу

	// Преобразуем байты в слово
	var word Word
	rawValue := binary.LittleEndian.Uint32(bytes[:]) // Преобразуем байты в целое число

	// Проверяем, является ли это командой (код операции в старшем байте)
	if bytes[3] > 0 { // Если это команда
		word.Cmd.Opcode = uint8(rawValue >> 24)              // Извлекаем код операции
		word.Cmd.BB = uint8((rawValue >> 22) & 0x03)         // Извлекаем BB
		word.Cmd.Address1 = uint16((rawValue >> 10) & 0xFFF) // Извлекаем Address1
		word.Cmd.Address2 = uint16(rawValue & 0x3FF)         // Извлекаем Address2
	} else { // Если это данные
		word.D.I = *(*int32)(unsafe.Pointer(&rawValue)) // Преобразуем целое число обратно в данные
	}
	return word, nil // Возвращаем считанное слово и nil, если ошибок не было
}

// WriteByte записывает один байт в память по заданному адресу
func (m *Memory) WriteByte(address int, value byte) error {
	m.data[address] = value // Записываем значение байта по указанному адресу в массив данных
	m.accessCount++         // Увеличиваем счетчик обращений к памяти
	return nil              // Возвращаем nil, если ошибок не было
}

// ReadByte считывает один байт из памяти по заданному адресу
func (m *Memory) ReadByte(address int) (byte, error) {
	m.accessCount++             // Увеличиваем счетчик обращений к памяти
	return m.data[address], nil // Возвращаем считанный байт из массива данных и nil, если ошибок не было
}

// Clear сбрасывает все ячейки памяти в ноль
func (m *Memory) Clear() {
	for i := range m.data { // Проходим по всем элементам массива данных
		m.data[i] = 0 // Устанавливаем значение каждого элемента в 0
	}
	m.accessCount = 0 // Сбрасываем счетчик обращений к памяти
	m.errorCount = 0  // Сбрасываем счетчик ошибок
}

// GetAccessCount возвращает общее количество обращений к памяти
func (m *Memory) GetAccessCount() int {
	return m.accessCount // Возвращаем текущее значение счетчика обращений к памяти
}

// GetErrorCount возвращает общее количество ошибок при доступе к памяти
func (m *Memory) GetErrorCount() int {
	return m.errorCount // Возвращаем текущее значение счетчика ошибок
}

// Close корректно закрывает ресурсы памяти
func (m *Memory) Close() {
	m.initialized = false // Устанавливаем флаг инициализации в false, чтобы указать, что память больше не используется
}
