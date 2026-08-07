package tools

import (
	"math"
	"testing"
)

type customInt int

func TestNumberToString(t *testing.T) {
	t.Run("int types", func(t *testing.T) {
		res, err := NumberToString(int(123))
		if err != nil || res != "123" {
			t.Errorf("int failed: got %q, err %v", res, err)
		}

		res, err = NumberToString(int8(-12))
		if err != nil || res != "-12" {
			t.Errorf("int8 failed: got %q, err %v", res, err)
		}

		res, err = NumberToString(int16(1234))
		if err != nil || res != "1234" {
			t.Errorf("int16 failed: got %q, err %v", res, err)
		}

		res, err = NumberToString(int32(-123456))
		if err != nil || res != "-123456" {
			t.Errorf("int32 failed: got %q, err %v", res, err)
		}

		res, err = NumberToString(int64(9223372036854775807))
		if err != nil || res != "9223372036854775807" {
			t.Errorf("int64 failed: got %q, err %v", res, err)
		}
	})

	t.Run("uint types", func(t *testing.T) {
		res, err := NumberToString(uint(456))
		if err != nil || res != "456" {
			t.Errorf("uint failed: got %q, err %v", res, err)
		}

		res, err = NumberToString(uint8(255))
		if err != nil || res != "255" {
			t.Errorf("uint8 failed: got %q, err %v", res, err)
		}

		res, err = NumberToString(uint16(65535))
		if err != nil || res != "65535" {
			t.Errorf("uint16 failed: got %q, err %v", res, err)
		}

		res, err = NumberToString(uint32(4294967295))
		if err != nil || res != "4294967295" {
			t.Errorf("uint32 failed: got %q, err %v", res, err)
		}

		res, err = NumberToString(uint64(18446744073709551615))
		if err != nil || res != "18446744073709551615" {
			t.Errorf("uint64 failed: got %q, err %v", res, err)
		}
	})

	t.Run("float types", func(t *testing.T) {
		res, err := NumberToString(float32(3.14))
		if err != nil || res != "3.14" {
			t.Errorf("float32 failed: got %q, err %v", res, err)
		}

		res, err = NumberToString(float64(2.718281828459045))
		if err != nil || res != "2.718281828459045" {
			t.Errorf("float64 failed: got %q, err %v", res, err)
		}
	})

	t.Run("complex types", func(t *testing.T) {
		res, err := NumberToString(complex64(1 + 2i))
		if err != nil || res != "(1+2i)" {
			t.Errorf("complex64 failed: got %q, err %v", res, err)
		}

		res, err = NumberToString(complex128(3.5 + 4.5i))
		if err != nil || res != "(3.5+4.5i)" {
			t.Errorf("complex128 failed: got %q, err %v", res, err)
		}
	})

	t.Run("unsupported type default branch", func(t *testing.T) {
		res, err := NumberToString(customInt(10))
		if err == nil {
			t.Errorf("expected error for customInt, got res %q", res)
		}
	})
}

func TestNumberToStringIgnoreError(t *testing.T) {
	if got := NumberToStringIgnoreError(int(100)); got != "100" {
		t.Errorf("expected '100', got %q", got)
	}

	if got := NumberToStringIgnoreError(customInt(100)); got != "" {
		t.Errorf("expected empty string for unsupported type, got %q", got)
	}
}

func TestStringToNumber(t *testing.T) {
	t.Run("int types success", func(t *testing.T) {
		vInt, err := StringToNumber[int]("123")
		if err != nil || vInt != 123 {
			t.Errorf("int failed: %v, err: %v", vInt, err)
		}

		vInt8, err := StringToNumber[int8]("-12")
		if err != nil || vInt8 != -12 {
			t.Errorf("int8 failed: %v, err: %v", vInt8, err)
		}

		vInt16, err := StringToNumber[int16]("1234")
		if err != nil || vInt16 != 1234 {
			t.Errorf("int16 failed: %v, err: %v", vInt16, err)
		}

		vInt32, err := StringToNumber[int32]("-123456")
		if err != nil || vInt32 != -123456 {
			t.Errorf("int32 failed: %v, err: %v", vInt32, err)
		}

		vInt64, err := StringToNumber[int64]("9223372036854775807")
		if err != nil || vInt64 != 9223372036854775807 {
			t.Errorf("int64 failed: %v, err: %v", vInt64, err)
		}
	})

	t.Run("uint types success", func(t *testing.T) {
		vUint, err := StringToNumber[uint]("456")
		if err != nil || vUint != 456 {
			t.Errorf("uint failed: %v, err: %v", vUint, err)
		}

		vUint8, err := StringToNumber[uint8]("255")
		if err != nil || vUint8 != 255 {
			t.Errorf("uint8 failed: %v, err: %v", vUint8, err)
		}

		vUint16, err := StringToNumber[uint16]("65535")
		if err != nil || vUint16 != 65535 {
			t.Errorf("uint16 failed: %v, err: %v", vUint16, err)
		}

		vUint32, err := StringToNumber[uint32]("4294967295")
		if err != nil || vUint32 != 4294967295 {
			t.Errorf("uint32 failed: %v, err: %v", vUint32, err)
		}

		vUint64, err := StringToNumber[uint64]("18446744073709551615")
		if err != nil || vUint64 != 18446744073709551615 {
			t.Errorf("uint64 failed: %v, err: %v", vUint64, err)
		}
	})

	t.Run("float types success", func(t *testing.T) {
		vFloat32, err := StringToNumber[float32]("3.14")
		if err != nil || math.Abs(float64(vFloat32-3.14)) > 1e-5 {
			t.Errorf("float32 failed: %v, err: %v", vFloat32, err)
		}

		vFloat64, err := StringToNumber[float64]("2.718281828459045")
		if err != nil || vFloat64 != 2.718281828459045 {
			t.Errorf("float64 failed: %v, err: %v", vFloat64, err)
		}
	})

	t.Run("complex types success", func(t *testing.T) {
		vComplex64, err := StringToNumber[complex64]("(1+2i)")
		if err != nil || vComplex64 != complex(1, 2) {
			t.Errorf("complex64 failed: %v, err: %v", vComplex64, err)
		}

		vComplex128, err := StringToNumber[complex128]("(3.5+4.5i)")
		if err != nil || vComplex128 != complex(3.5, 4.5) {
			t.Errorf("complex128 failed: %v, err: %v", vComplex128, err)
		}
	})

	t.Run("invalid string error paths", func(t *testing.T) {
		invalid := "abc"

		if _, err := StringToNumber[int](invalid); err == nil {
			t.Error("expected error for int")
		}
		if _, err := StringToNumber[int8](invalid); err == nil {
			t.Error("expected error for int8")
		}
		if _, err := StringToNumber[int16](invalid); err == nil {
			t.Error("expected error for int16")
		}
		if _, err := StringToNumber[int32](invalid); err == nil {
			t.Error("expected error for int32")
		}
		if _, err := StringToNumber[int64](invalid); err == nil {
			t.Error("expected error for int64")
		}
		if _, err := StringToNumber[uint](invalid); err == nil {
			t.Error("expected error for uint")
		}
		if _, err := StringToNumber[uint8](invalid); err == nil {
			t.Error("expected error for uint8")
		}
		if _, err := StringToNumber[uint16](invalid); err == nil {
			t.Error("expected error for uint16")
		}
		if _, err := StringToNumber[uint32](invalid); err == nil {
			t.Error("expected error for uint32")
		}
		if _, err := StringToNumber[uint64](invalid); err == nil {
			t.Error("expected error for uint64")
		}
		if _, err := StringToNumber[float32](invalid); err == nil {
			t.Error("expected error for float32")
		}
		if _, err := StringToNumber[float64](invalid); err == nil {
			t.Error("expected error for float64")
		}
		if _, err := StringToNumber[complex64](invalid); err == nil {
			t.Error("expected error for complex64")
		}
		if _, err := StringToNumber[complex128](invalid); err == nil {
			t.Error("expected error for complex128")
		}
	})

	t.Run("unsupported type default branch", func(t *testing.T) {
		if _, err := StringToNumber[customInt]("123"); err == nil {
			t.Error("expected error for customInt")
		}
	})
}

func TestStringToNumberIgnoreError(t *testing.T) {
	if got := StringToNumberIgnoreError[int]("123"); got != 123 {
		t.Errorf("expected 123, got %v", got)
	}

	if got := StringToNumberIgnoreError[int]("invalid"); got != 0 {
		t.Errorf("expected 0 for invalid string, got %v", got)
	}

	if got := StringToNumberIgnoreError[customInt]("123"); got != 0 {
		t.Errorf("expected 0 for unsupported type, got %v", got)
	}
}
