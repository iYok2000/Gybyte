// Package logger จัดหา zap structured logger ที่ใช้ร่วมกันทุกเลเยอร์ — Req 15.6
package logger

import (
	"sync"

	"go.uber.org/zap"
)

var (
	// log เป็น instance ร่วม (singleton) ที่ทุกเลเยอร์เรียกใช้ผ่าน L()
	log  *zap.Logger
	once sync.Once
)

// Init สร้าง zap logger หนึ่งครั้ง (idempotent) โดยเลือกโปรไฟล์ตาม environment:
//   - production → JSON structured logger (NewProduction)
//   - อื่น ๆ      → console logger อ่านง่าย (NewDevelopment)
func Init(environment string) *zap.Logger {
	once.Do(func() {
		var (
			l   *zap.Logger
			err error
		)
		if environment == "production" {
			l, err = zap.NewProduction()
		} else {
			l, err = zap.NewDevelopment()
		}
		if err != nil {
			// fallback เป็น no-op logger เพื่อไม่ให้แอปล้มเพียงเพราะ init logger ล้มเหลว
			l = zap.NewNop()
		}
		log = l
	})
	return log
}

// L คืน logger ร่วม; ถ้ายังไม่ได้ Init จะคืน no-op logger เพื่อความปลอดภัย
func L() *zap.Logger {
	if log == nil {
		return zap.NewNop()
	}
	return log
}

// Sync flush buffer ที่ค้างของ logger (ควรเรียกตอนปิดโปรแกรม)
func Sync() {
	if log != nil {
		_ = log.Sync()
	}
}
