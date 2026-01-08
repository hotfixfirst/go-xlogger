package xlogger

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestField_Constructors tests all field constructor functions
func TestField_Constructors(t *testing.T) {
	t.Run("should create string field", func(t *testing.T) {
		field := String("key", "value")

		assert.Equal(t, "key", field.Key())
		assert.Equal(t, "value", field.Value())
		assert.Equal(t, StringType, field.Type())
	})

	t.Run("should create int field", func(t *testing.T) {
		field := Int("count", 42)

		assert.Equal(t, "count", field.Key())
		assert.Equal(t, 42, field.Value())
		assert.Equal(t, IntType, field.Type())
	})

	t.Run("should create int64 field", func(t *testing.T) {
		field := Int64("id", 123456789)

		assert.Equal(t, "id", field.Key())
		assert.Equal(t, int64(123456789), field.Value())
		assert.Equal(t, IntType, field.Type()) // Int64 uses IntType
	})

	t.Run("should create float64 field", func(t *testing.T) {
		field := Float64("price", 19.99)

		assert.Equal(t, "price", field.Key())
		assert.Equal(t, 19.99, field.Value())
		assert.Equal(t, Float64Type, field.Type())
	})

	t.Run("should create bool field", func(t *testing.T) {
		field := Bool("active", true)

		assert.Equal(t, "active", field.Key())
		assert.Equal(t, true, field.Value())
		assert.Equal(t, BoolType, field.Type())
	})

	t.Run("should create error field", func(t *testing.T) {
		err := errors.New("test error")
		field := Error(err)

		assert.Equal(t, "error", field.Key())
		assert.Equal(t, err, field.Value())
		assert.Equal(t, ErrorType, field.Type())
	})

	t.Run("should create error field with nil", func(t *testing.T) {
		field := Error(nil)

		assert.Equal(t, "error", field.Key())
		assert.Nil(t, field.Value())
		assert.Equal(t, ErrorType, field.Type())
	})

	t.Run("should create duration field", func(t *testing.T) {
		duration := 5 * time.Second
		field := Duration("timeout", duration)

		assert.Equal(t, "timeout", field.Key())
		assert.Equal(t, duration, field.Value())
		assert.Equal(t, DurationType, field.Type())
	})

	t.Run("should create time field", func(t *testing.T) {
		now := time.Now()
		field := Time("created_at", now)

		assert.Equal(t, "created_at", field.Key())
		assert.Equal(t, now, field.Value())
		assert.Equal(t, TimeType, field.Type())
	})

	t.Run("should create any field", func(t *testing.T) {
		value := map[string]interface{}{"nested": "value"}
		field := Any("metadata", value)

		assert.Equal(t, "metadata", field.Key())
		assert.Equal(t, value, field.Value())
		assert.Equal(t, AnyType, field.Type())
	})
}

// TestField_EdgeCases tests edge cases for field constructors
func TestField_EdgeCases(t *testing.T) {
	t.Run("should handle empty string key", func(t *testing.T) {
		field := String("", "value")

		assert.Equal(t, "", field.Key())
		assert.Equal(t, "value", field.Value())
		assert.Equal(t, StringType, field.Type())
	})

	t.Run("should handle empty string value", func(t *testing.T) {
		field := String("key", "")

		assert.Equal(t, "key", field.Key())
		assert.Equal(t, "", field.Value())
		assert.Equal(t, StringType, field.Type())
	})

	t.Run("should handle zero values", func(t *testing.T) {
		tests := []struct {
			name  string
			field Field
			typ   FieldType
		}{
			{"zero int", Int("zero", 0), IntType},
			{"zero int64", Int64("zero", 0), IntType},
			{"zero float", Float64("zero", 0.0), Float64Type},
			{"false bool", Bool("zero", false), BoolType},
			{"zero duration", Duration("zero", 0), DurationType},
			{"zero time", Time("zero", time.Time{}), TimeType},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				assert.Equal(t, "zero", test.field.Key())
				assert.Equal(t, test.typ, test.field.Type())
			})
		}
	})

	t.Run("should handle negative values", func(t *testing.T) {
		intField := Int("negative", -42)
		assert.Equal(t, -42, intField.Value())

		int64Field := Int64("negative", -123456789)
		assert.Equal(t, int64(-123456789), int64Field.Value())

		floatField := Float64("negative", -19.99)
		assert.Equal(t, -19.99, floatField.Value())

		durationField := Duration("negative", -5*time.Second)
		assert.Equal(t, -5*time.Second, durationField.Value())
	})

	t.Run("should handle nil values in Any field", func(t *testing.T) {
		field := Any("nil_value", nil)

		assert.Equal(t, "nil_value", field.Key())
		assert.Nil(t, field.Value())
		assert.Equal(t, AnyType, field.Type())
	})
}

// TestFieldType_Constants tests field type constants
func TestFieldType_Constants(t *testing.T) {
	t.Run("should have unique field type values", func(t *testing.T) {
		types := []FieldType{
			StringType,
			IntType,
			Float64Type,
			BoolType,
			ErrorType,
			DurationType,
			TimeType,
			AnyType,
		}

		// Check that all types are unique
		typeMap := make(map[FieldType]bool)
		for _, typ := range types {
			assert.False(t, typeMap[typ], "FieldType %d should be unique", typ)
			typeMap[typ] = true
		}

		// Check expected values (iota starts from 0)
		assert.Equal(t, FieldType(0), StringType)
		assert.Equal(t, FieldType(1), IntType)
		assert.Equal(t, FieldType(2), Float64Type)
		assert.Equal(t, FieldType(3), BoolType)
		assert.Equal(t, FieldType(4), ErrorType)
		assert.Equal(t, FieldType(5), DurationType)
		assert.Equal(t, FieldType(6), TimeType)
		assert.Equal(t, FieldType(7), AnyType)
	})
}

// TestField_GetterMethods tests Field getter methods
func TestField_GetterMethods(t *testing.T) {
	t.Run("should return correct values from getters", func(t *testing.T) {
		tests := []struct {
			name      string
			field     Field
			key       string
			value     interface{}
			fieldType FieldType
		}{
			{
				name:      "string field",
				field:     String("name", "test"),
				key:       "name",
				value:     "test",
				fieldType: StringType,
			},
			{
				name:      "int field",
				field:     Int("count", 42),
				key:       "count",
				value:     42,
				fieldType: IntType,
			},
			{
				name:      "bool field",
				field:     Bool("active", true),
				key:       "active",
				value:     true,
				fieldType: BoolType,
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				assert.Equal(t, test.key, test.field.Key())
				assert.Equal(t, test.value, test.field.Value())
				assert.Equal(t, test.fieldType, test.field.Type())
			})
		}
	})

	t.Run("getters should be consistent", func(t *testing.T) {
		field := String("test", "value")

		// Call getters multiple times to ensure consistency
		for i := 0; i < 5; i++ {
			assert.Equal(t, "test", field.Key())
			assert.Equal(t, "value", field.Value())
			assert.Equal(t, StringType, field.Type())
		}
	})
}

// TestField_TypeSafety tests type safety of field operations
func TestField_TypeSafety(t *testing.T) {
	t.Run("should maintain type information", func(t *testing.T) {
		stringField := String("str", "value")
		intField := Int("num", 42)
		boolField := Bool("flag", true)

		// Type information should be preserved
		assert.Equal(t, StringType, stringField.Type())
		assert.Equal(t, IntType, intField.Type())
		assert.Equal(t, BoolType, boolField.Type())

		// Values should maintain their types
		assert.IsType(t, "", stringField.Value())
		assert.IsType(t, 0, intField.Value())
		assert.IsType(t, true, boolField.Value())
	})

	t.Run("should handle interface{} values correctly", func(t *testing.T) {
		complexValue := map[string]interface{}{
			"nested": map[string]string{
				"key": "value",
			},
		}

		anyField := Any("complex", complexValue)

		assert.Equal(t, AnyType, anyField.Type())
		assert.Equal(t, complexValue, anyField.Value())

		// Should be able to type assert back
		if value, ok := anyField.Value().(map[string]interface{}); ok {
			if nested, ok := value["nested"].(map[string]string); ok {
				assert.Equal(t, "value", nested["key"])
			} else {
				t.Error("Failed to type assert nested value")
			}
		} else {
			t.Error("Failed to type assert complex value")
		}
	})
}
