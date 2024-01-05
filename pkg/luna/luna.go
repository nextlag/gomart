package luna

import (
	"strconv"
)

// CheckValidOrder проверяет, проходит ли номер заказа алгоритм Луна.
func CheckValidOrder(orderNumber string) bool {
	digits, err := toDigits(orderNumber)
	if err != nil {
		return false
	}

	// Удвоить значения каждого второго числа.
	for i := len(digits) - 2; i >= 0; i -= 2 {
		digits[i] *= 2
		if digits[i] > 9 {
			digits[i] -= 9
		}
	}

	// Считаем сумму цифр в последовательности.
	sum := 0
	for _, digit := range digits {
		sum += digit
	}

	// Проверяем, делится ли сумма на 10.
	return sum%10 == 0
}

// toDigits преобразует строку с номером заказа в слайс цифр.
func toDigits(orderNumber string) ([]int, error) {
	var digits []int
	for _, char := range orderNumber {
		digit, err := strconv.Atoi(string(char))
		if err != nil {
			return nil, err
		}
		digits = append(digits, digit)
	}
	return digits, nil
}
