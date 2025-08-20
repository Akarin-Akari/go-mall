package auth

import (
	"testing"
)

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		errType  error
	}{
		{
			name:     "有效密码",
			password: "SecurePass123!",
			wantErr:  false,
		},
		{
			name:     "密码太短",
			password: "Short1!",
			wantErr:  true,
			errType:  ErrPasswordTooShort,
		},
		{
			name:     "缺少大写字母",
			password: "securepass123!",
			wantErr:  true,
			errType:  ErrPasswordTooWeak,
		},
		{
			name:     "缺少小写字母",
			password: "SECUREPASS123!",
			wantErr:  true,
			errType:  ErrPasswordTooWeak,
		},
		{
			name:     "缺少数字",
			password: "SecurePass!",
			wantErr:  true,
			errType:  ErrPasswordTooWeak,
		},
		{
			name:     "缺少特殊字符",
			password: "SecurePass123",
			wantErr:  true,
			errType:  ErrPasswordTooWeak,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePasswordStrength() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errType != nil && err != tt.errType {
				t.Errorf("ValidatePasswordStrength() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	password := "SecurePass123!"
	
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	
	if hash == "" {
		t.Error("HashPassword() returned empty hash")
	}
	
	if hash == password {
		t.Error("HashPassword() returned plaintext password")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "SecurePass123!"
	wrongPassword := "WrongPass123!"
	
	// 生成密码哈希
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	
	// 测试正确密码
	err = VerifyPassword(hash, password)
	if err != nil {
		t.Errorf("VerifyPassword() with correct password error = %v", err)
	}
	
	// 测试错误密码
	err = VerifyPassword(hash, wrongPassword)
	if err == nil {
		t.Error("VerifyPassword() with wrong password should return error")
	}
}

func TestIsPasswordValid(t *testing.T) {
	password := "SecurePass123!"
	wrongPassword := "WrongPass123!"
	
	// 生成密码哈希
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	
	// 测试正确密码
	if !IsPasswordValid(hash, password) {
		t.Error("IsPasswordValid() with correct password should return true")
	}
	
	// 测试错误密码
	if IsPasswordValid(hash, wrongPassword) {
		t.Error("IsPasswordValid() with wrong password should return false")
	}
}

// TODO: 密码强度评分测试在MVP阶段暂不需要
// 遵循YAGNI原则，如果将来需要可以再添加

/*
func TestPasswordStrengthScore(t *testing.T) {
	// 测试实现留待将来需要时添加
}

func TestGetPasswordStrengthLevel(t *testing.T) {
	// 测试实现留待将来需要时添加
}
*/

func BenchmarkHashPassword(b *testing.B) {
	password := "SecurePass123!"
	
	for i := 0; i < b.N; i++ {
		_, err := HashPassword(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	password := "SecurePass123!"
	hash, err := HashPassword(password)
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := VerifyPassword(hash, password)
		if err != nil {
			b.Fatal(err)
		}
	}
}
