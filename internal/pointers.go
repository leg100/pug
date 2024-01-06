package internal

import "time"

func String(str string) *string   { return &str }
func Int(i int) *int              { return &i }
func Int64(i int64) *int64        { return &i }
func UInt(i uint) *uint           { return &i }
func Bool(b bool) *bool           { return &b }
func Time(t time.Time) *time.Time { return &t }
