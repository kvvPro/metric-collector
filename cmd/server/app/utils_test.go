package app

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func Test_isValidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "1",
			url:  "/update/gauge/mm/1.09",
			want: true,
		},
		{
			name: "2",
			url:  "/update/counter/",
			want: false,
		},
		{
			name: "3",
			url:  "/update/ffff/ri/3",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidURL(tt.url); got != tt.want {
				t.Errorf("isValidURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isNameMissing(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "1",
			url:  "/update/gauge/mm/1.09",
			want: false,
		},
		{
			name: "2",
			url:  "/update/counter/",
			want: true,
		},
		{
			name: "3",
			url:  "/update/counter/",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNameMissing(tt.url); got != tt.want {
				t.Errorf("isNameMissing() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isValidType(t *testing.T) {
	tests := []struct {
		name string
		t    string
		want bool
	}{
		{
			name: "1",
			t:    "gauge",
			want: true,
		},
		{
			name: "2",
			t:    "counter",
			want: true,
		},
		{
			name: "3",
			t:    "counterfff",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidType(tt.t); got != tt.want {
				t.Errorf("isValidType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isValidValue(t *testing.T) {
	tests := []struct {
		name string
		v    string
		want bool
	}{
		{
			name: "1",
			v:    "ss",
			want: false,
		},
		{
			name: "2",
			v:    "34",
			want: true,
		},
		{
			name: "3",
			v:    "14.5",
			want: true,
		},
		{
			name: "4",
			v:    "55g",
			want: false,
		},
		{
			name: "5",
			v:    "12.s",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidValue(tt.v); got != tt.want {
				t.Errorf("isValidValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isValidParams(t *testing.T) {
	tests := []struct {
		name string
		url  string
		res  []string
		want bool
	}{
		{
			name: "1",
			url:  "/update/counter/metric1/9",
			res: []string{
				"",
				"update",
				"counter",
				"metric1",
				"9",
			},
			want: true,
		},
		{
			name: "2",
			url:  "/update/gauge/opa/3.2",
			res: []string{
				"",
				"update",
				"gauge",
				"opa",
				"3.2",
			},
			want: true,
		},
		{
			name: "3",
			url:  "/update/counter/",
			res:  nil,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.url, nil)
			// создаём новый Recorder
			writer := httptest.NewRecorder()

			res, got := isValidParams(request, writer)
			if !reflect.DeepEqual(res, tt.res) {
				t.Errorf("isValidParams() got = %v, want %v", res, tt.res)
			}
			if got != tt.want {
				t.Errorf("isValidParams() got1 = %v, want %v", got, tt.want)
			}
		})
	}
}
