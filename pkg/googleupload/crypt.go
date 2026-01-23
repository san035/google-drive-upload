package googleupload

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/billgraziano/dpapi"
)

var (
	// EntropyBytes - энтропия для DPAPI шифрования
	EntropyBytes = []byte("solt-fr-Ht-15!")
)

const (
	// EncryptedMarker - маркер для определения зашифрованного файла
	EncryptedMarker = "DPAPI_ENCRYPTED:"
)

// EncryptBytesWithDPAPI шифрует данные с использованием Windows DPAPI
func EncryptBytesWithDPAPI(data []byte) ([]byte, error) {
	return dpapi.EncryptBytesEntropy(data, EntropyBytes)
}

// DecryptBytesWithDPAPI дешифрует данные с использованием Windows DPAPI
func DecryptBytesWithDPAPI(data []byte) ([]byte, error) {
	return dpapi.DecryptBytesEntropy(data, EntropyBytes)
}

// IsFileEncrypted проверяет, зашифрован ли файл по содержимому
func IsFileEncrypted(content []byte) bool {
	return strings.HasPrefix(string(content), EncryptedMarker)
}

// EncryptFile шифрует файл с использованием DPAPI
func EncryptFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("ошибка чтения файла %s: %v", filePath, err)
	}

	if len(content) == 0 {
		return fmt.Errorf("файл пуст: %s", filePath)
	}

	// Проверяем, не зашифрован ли уже файл
	if IsFileEncrypted(content) {
		return nil // файл уже зашифрован
	}

	err = EncryptContentAndSaveToFile(filePath, content)
	return err
}

func EncryptContentAndSaveToFile(filePath string, content []byte) error {
	// Шифруем данные
	encryptedData, err := EncryptBytesWithDPAPI(content)
	if err != nil {
		return err
	}

	// Добавляем маркер в начало зашифрованных данных
	// Используем base64 для безопасного хранения бинарных данных в файле
	encodedData := EncryptedMarker + base64.StdEncoding.EncodeToString(encryptedData)

	// Записываем зашифрованные данные обратно в файл
	err = os.WriteFile(filePath, []byte(encodedData), 0600)
	return err
}

// DecryptFile дешифрует файл с использованием DPAPI
// Если файл не зашифрован, он будет зашифрован и сохранён
func DecryptFile(filePath string) ([]byte, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла %s: %v", filePath, err)
	}

	if len(content) == 0 {
		return content, nil
	}

	// Проверяем, зашифрован ли файл
	if !IsFileEncrypted(content) {
		// Файл не зашифрован - шифруем и сохраняем
		encryptedData, err := EncryptBytesWithDPAPI(content)
		if err != nil {
			return nil, fmt.Errorf("ошибка шифрования файла: %v", err)
		}

		encodedData := EncryptedMarker + base64.StdEncoding.EncodeToString(encryptedData)
		if err := os.WriteFile(filePath, []byte(encodedData), 0600); err != nil {
			return nil, fmt.Errorf("ошибка записи зашифрованного файла: %v", err)
		}

		// Возвращаем оригинальные (незашифрованные) данные
		return content, nil
	}

	// Извлекаем base64-данные (убираем маркер)
	lenEM := len(EncryptedMarker)
	if len(content) < lenEM {
		return nil, fmt.Errorf("маркер зашифрованного файла не найден: %s", filePath)
	}
	encodedContent := string(content[lenEM:])

	// Декодируем base64
	encryptedData, err := base64.StdEncoding.DecodeString(encodedContent)
	if err != nil {
		return nil, fmt.Errorf("ошибка декодирования base64: %v", err)
	}

	// Дешифруем данные
	decryptedData, err := DecryptBytesWithDPAPI(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("ошибка дешифрования файла: %s error: %v", filePath, err)
	}

	return decryptedData, nil
}
