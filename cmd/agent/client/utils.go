package client

import (
	"reflect"

	"github.com/kvvPro/metric-collector/internal/metrics"
)

func DeepFields(iface interface{}) []Metric {
	fields := make([]Metric, 0)
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)

	for i := 0; i < ift.NumField(); i++ {
		v := ifv.Field(i)
		field := ift.Field(i)
		ftype := field.Type.Name()

		switch v.Kind() {
		case reflect.Struct:
			fields = append(fields, DeepFields(v.Interface())...)
		default:
			ival := reflect.ValueOf(ifv.Field(i).Interface())
			c := NewMetric(field.Name, ftype, ival)
			if c != nil {
				fields = append(fields, c)
			}
		}
	}

	return fields
}

func NewMetric(mname string, mtype string, ival reflect.Value) Metric {
	switch mtype {
	case "float64":
		val := ival.Float()
		c := metrics.NewGauge(mname, mtype, val)
		return c
	case "uint64":
		val := ival.Uint()
		c := metrics.NewCounter(mname, mtype, int64(val))
		return c
	case "int64":
		val := ival.Int()
		c := metrics.NewCounter(mname, mtype, val)
		return c
	default:
		return nil
	}
}
