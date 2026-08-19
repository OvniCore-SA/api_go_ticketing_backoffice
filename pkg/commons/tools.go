package commons

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword genera un hash de la contraseña usando bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash compara una contraseña con su hash.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func StringIsEmpty(e string) bool {
	return len(strings.TrimSpace(e)) == 0
}

func IsEmailValid(e string) bool {
	pattern := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	email := strings.TrimSpace(e)
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	return pattern.MatchString(email)
}

func IsNameValid(name string) error {
	if len(name) < 3 {
		return errors.New("El nombre y el apellido deben tener al menos 3 letras")
	}
	for _, char := range name {
		if !(unicode.IsLetter(char) || unicode.IsSpace(char)) {
			return errors.New("El nombre y el apellido solo pueden contener letras y espacios")
		}
	}
	return nil
}

func IsDniValid(dni string) bool {
	// Expresión regular para validar un DNI que tenga entre 7 y 8 dígitos
	pattern := regexp.MustCompile(`^\d{7,8}$`)
	return pattern.MatchString(dni)
}

func IsAdult(birthdate string) (birth time.Time, erro error) {
	layout := "02-01-2006" // Formato de fecha esperado
	birth, err := time.Parse(layout, birthdate)
	if err != nil {
		erro = errors.New("Debes ingresar una fecha válida")
		return
	}

	today := time.Now()
	age := today.Year() - birth.Year()

	// Si aún no ha pasado su cumpleaños este año, restar un año a la edad
	if today.YearDay() < birth.YearDay() {
		age--
	}

	if age < 18 {
		erro = errors.New("Debes tener al menos 18 años")
		return
	}
	return birth, nil
}

func IsValidPhoneNumber(phoneNumber string) error {
	for _, char := range phoneNumber {
		if !unicode.IsDigit(char) {
			return errors.New("El número de teléfono solo puede contener caracteres numéricos")
		}
	}
	if len(phoneNumber) < 10 {
		return errors.New("El número de teléfono debe tener al menos 10 caracteres")
	}
	return nil
}

// ContainsSpaces verifica si un string contiene espacios en blanco
func ContainsSpaces(str string) bool {
	return strings.Contains(str, " ")
}

func FormatPhoneNumber(phoneNumber string) string {
	phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "+", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")
	if len(phoneNumber) > 0 && (phoneNumber[0] == '9' || phoneNumber[0] == '0') {
		phoneNumber = phoneNumber[1:]
	}
	if strings.HasPrefix(phoneNumber, "549") {
		phoneNumber = phoneNumber[3:]
	}
	if strings.HasPrefix(phoneNumber, "54") {
		phoneNumber = phoneNumber[2:]
	}
	return phoneNumber
}

// Formateo primer letra de un Nombre y Apellido
func FormatNombre(name string) string {
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return name
	}
	return strings.ToUpper(string(name[0])) + name[1:]
}

func Difference(slice1 []string, slice2 []string) []string {
	var diff []string

	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				if len(s1) > 0 {
					diff = append(diff, s1)
				}
			}
		}
		// Swap the slices, only if it was the first loop
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return diff
}

/*
funciones que se pueden utilizar para quitar caracteres especiales en una cadena
*/
func SpaceStringsBuilder(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	for _, ch := range str {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func StripSpaces(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			// if the character is a space, drop it
			return -1
		}
		// else keep it in the string
		return r
	}, str)
}

/*
remplazar un caracter por otro
*/
func ReplaceCharacters(str, valorBuscar, valorReemplazar string) string {
	resultadoString := strings.Replace(str, valorBuscar, valorReemplazar, -1)
	return resultadoString
}

// devuelve una fecha string en formato ISO 8601 con la HH:mm:ss finales del dia.
// Uso: comparar limites de fechas
func GetDateFirstMoment(fecha time.Time) (fechaISO string) {
	year, month, day := fecha.Date()
	t := time.Date(year, month, day, 00, 00, 00, 1, fecha.Location())
	return t.Format(time.RFC3339)
}

// devuelve una fecha string en formato ISO 8601 con la HH:mm:ss finales del dia.
// Uso: comparar limites de fechas
func GetDateLastMoment(fecha time.Time) (fechaISO string) {
	year, month, day := fecha.Date()
	t := time.Date(year, month, day, 23, 59, 59, 999, fecha.Location())
	return t.Format(time.RFC3339)
}

// CalcularEdad recibe una fecha de nacimiento y calcula la edad actual.
func CalcularEdad(fechaNacimiento time.Time) int64 {
	fechaActual := time.Now()
	// Calcular la diferencia en años
	edad := fechaActual.Year() - fechaNacimiento.Year()
	// Ajustar la edad si no ha cumplido años este año
	if fechaActual.YearDay() < fechaNacimiento.YearDay() {
		edad--
	}
	return int64(edad)
}

// ValidatePassword valida que la contraseña cumpla con ciertos requisitos
func ValidatePassword(password string) error {
	if StringIsEmpty(password) {
		return errors.New("Debes enviar una contraseña válida")
	}
	if len(password) < 8 {
		return errors.New("La contraseña debe tener al menos 8 caracteres")
	}

	// Verificar que no tenga espacios en blanco
	if ContainsSpaces(password) {
		return errors.New("No se permiten espacios en blanco en las contraseñas")
	}

	// Verificar que tenga al menos un número
	match, err := regexp.MatchString("[0-9]", password)
	if err != nil {
		return err
	}
	if !match {
		return errors.New("La contraseña debe contener al menos un número")
	}

	return nil
}

// IsAlphanumeric verifica si un string solo contiene letras y números
func IsAlphanumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}

// GenerateToken genera un token seguro aleatorio
func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Retorna la hora actual en la zona horaria de Argentina
func TimeNowArgentina() time.Time {
	loc, _ := time.LoadLocation("America/Argentina/Buenos_Aires")
	return time.Now().In(loc)
}

func GeneratePaymentReferenceID(membershipID uint) *string {
	reference := fmt.Sprintf("M%d-%d", membershipID, time.Now().UnixNano())
	return &reference
}

// retorna time con el ultimo momento de la fecha
func GetDateLastMomentTime(fecha time.Time) (lastMomentDate time.Time) {
	year, month, day := fecha.Date()
	return time.Date(year, month, day, 23, 59, 59, 999, fecha.Location())
}

// Función para obtener el primer momento de un día en time.Time.
func GetDateFirstMomentTime(fecha time.Time) time.Time {
	year, month, day := fecha.Date()
	t := time.Date(year, month, day, 00, 00, 00, 0, time.UTC)
	return t
}
