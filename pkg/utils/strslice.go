package utils

func StrSlice(s string, idx int) string {
	// Преобразуем строку в срез рун для корректной работы с Unicode
	runes := []rune(s)

	// Если длина меньше или равна 1 - возвращаем пустую строку
	if len(runes) <= 1 {
		return ""
	}

	// Возвращаем подстроку начиная с индекса 1
	return string(runes[idx:])
}
