package logger

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm/logger"
)

func TestGormToELK(t *testing.T) {
	elkLogger := NewGormToELK(WithLevel("debug"))
	if elkLogger == nil || elkLogger.Log == nil {
		t.Fatal("expected non-nil GormToELK and Log")
	}

	// Test LogMode
	newLogger := elkLogger.LogMode(logger.Warn)
	if newLogger == nil {
		t.Fatal("expected non-nil logger from LogMode")
	}

	ctx := context.Background()

	// Test Info, Warn, Error methods
	elkLogger.Info(ctx, "info test %s", "arg")
	elkLogger.Warn(ctx, "warn test %s", "arg")
	elkLogger.Error(ctx, "error test %s", "arg")

	// Test Trace - SQL_OK
	fc := func() (string, int64) {
		return "SELECT * FROM users WHERE id = 1", 1
	}
	elkLogger.Trace(ctx, time.Now(), fc, nil)

	// Test Trace - SLOW_SQL
	fcSlow := func() (string, int64) {
		return "SELECT * FROM large_table", 1000
	}
	elkLogger.Trace(ctx, time.Now().Add(-300*time.Millisecond), fcSlow, nil)

	// Test Trace - SQL_ERROR
	fcErr := func() (string, int64) {
		return "SELECT * FROM non_existing_table", 0
	}
	elkLogger.Trace(ctx, time.Now(), fcErr, errors.New("table not found"))

	// Test Trace - Truncated SQL (> 500 chars default)
	fcLongSQL := func() (string, int64) {
		return "SELECT id, username, email, created_at, updated_at FROM users WHERE active = true AND deleted_at IS NULL ORDER BY created_at DESC", 10
	}
	elkLogger.Trace(ctx, time.Now(), fcLongSQL, nil)

	// Test Trace with custom MaxSQLLength via SetMaxSQLLength and UTF-8 characters
	customLenELK := NewGormToELK().SetMaxSQLLength(10)
	fcUTF8 := func() (string, int64) {
		return "SELECT * FROM 用户表 WHERE 名字 = '张三丰'", 1
	}
	customLenELK.Trace(ctx, time.Now(), fcUTF8, nil)

	// Test NewGormToELK with direct GormOption and LoggerOption
	directOptELK := NewGormToELK(WithMaxSQLLength(250), WithLevel("debug"), WithSlowThreshold(100*time.Millisecond))
	if directOptELK.MaxSQLLength != 250 {
		t.Errorf("expected MaxSQLLength 250, got %d", directOptELK.MaxSQLLength)
	}
	if directOptELK.SlowThreshold != 100*time.Millisecond {
		t.Errorf("expected SlowThreshold 100ms, got %v", directOptELK.SlowThreshold)
	}

	// Test Trace with unlimited SQL length (-1)
	unlimitedELK := NewGormToELK()
	unlimitedELK.MaxSQLLength = -1
	unlimitedELK.Trace(ctx, time.Now(), fcLongSQL, nil)

	// Test Trace with Silent LogLevel (fc should not panic or execute unnecessary work)
	silentCalled := false
	fcSilent := func() (string, int64) {
		silentCalled = true
		return "SELECT 1", 1
	}
	silentELK := NewGormToELK()
	silentELK.LogLevel = logger.Silent
	silentELK.Trace(ctx, time.Now(), fcSilent, nil)
	if silentCalled {
		t.Error("expected fc not to be called when LogLevel is Silent")
	}

	// Test Trace with Caller context key
	ctxWithCaller := context.WithValue(ctx, Caller, "TestCustomCallerFunc")
	elkLogger.Trace(ctxWithCaller, time.Now(), fc, nil)

	// Test getLogger fallback constructor
	gl := getLogger()
	if gl == nil || gl.Log == nil {
		t.Fatal("expected non-nil logger from getLogger()")
	}

	// Test Nil Log fallback handling in Trace
	nilLogELK := &GormToELK{
		Config: logger.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      logger.Info,
		},
	}
	nilLogELK.Info(ctx, "info test")
	nilLogELK.Warn(ctx, "warn test")
	nilLogELK.Error(ctx, "error test")
	nilLogELK.Trace(ctx, time.Now(), fc, nil)
}

