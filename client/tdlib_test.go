package client

import (
	"math"
	"testing"
)

func TestJsonInt64(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    JsonInt64
		wantErr bool
	}{
		{"positive_number", []byte("123"), 123, false},
		{"negative_number", []byte("-123"), -123, false},
		{"string_positive", []byte(`"123"`), 123, false},
		{"string_negative", []byte(`"-123"`), -123, false},
		{"zero", []byte("0"), 0, false},
		{"string_zero", []byte(`"0"`), 0, false},
		{"empty_string#1", []byte(`""`), 0, true},
		{"empty_string#2", []byte(``), 0, true},
		{"invalid_string", []byte(`"abc"`), 0, true},
		{"max_int64", []byte("9223372036854775807"), JsonInt64(math.MaxInt64), false},
		{"min_int64", []byte("-9223372036854775808"), JsonInt64(math.MinInt64), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got JsonInt64
			err := got.UnmarshalJSON(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("UnmarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJsonInt64_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		value   JsonInt64
		want    string
		wantErr bool
	}{
		{"positive", 123, `"123"`, false},
		{"negative", -123, `"-123"`, false},
		{"zero", 0, `"0"`, false},
		{"max_int64", JsonInt64(math.MaxInt64), `"9223372036854775807"`, false},
		{"min_int64", JsonInt64(math.MinInt64), `"-9223372036854775808"`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.value.MarshalJSON()

			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && string(got) != tt.want {
				t.Errorf("MarshalJSON() = %v, want %v", string(got), tt.want)
			}
		})
	}
}
