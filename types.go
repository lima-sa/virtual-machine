package main

// Data представляет структуру, подобную объединению, для хранения различных типов данных
type Data struct {
	I int32   // Целочисленное значение (32-битное знаковое целое)
	F float32 // Значение с плавающей запятой (32-битное)
}

// CommandData представляет структуру команды
type CommandData struct {
	Opcode   uint8  // Код операции (6 бит)
	BB       uint8  // 2 бита для BB (включает режим регистра)
	Address1 uint16 // Первый адрес/индекс регистра (12 бит)
	Address2 uint16 // Второй адрес/индекс регистра (12 бит)
}

// Word представляет объединение Data и CommandData
type Word struct {
	D   Data        // Поле для хранения данных типа Data
	Cmd CommandData // Поле для хранения данных типа CommandData
}

// MemoryError представляет ошибки доступа к памяти
type MemoryError struct {
	Operation string // Описание операции, вызвавшей ошибку
	Address   int    // Адрес, по которому произошла ошибка
	Message   string // Сообщение об ошибке
}
