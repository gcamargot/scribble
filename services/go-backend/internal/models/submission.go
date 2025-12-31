package models

import (
	"time"
)

// Submission represents a code submission by a user
type Submission struct {
	ID              string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID          string    `json:"user_id" gorm:"not null;index"`
	ProblemID       string    `json:"problem_id" gorm:"not null;index"`
	Language        string    `json:"language" gorm:"not null"`
	Code            string    `json:"code" gorm:"type:text;not null"`
	Status          string    `json:"status" gorm:"not null;default:'pending'"` // pending, running, accepted, wrong_answer, runtime_error, time_limit, memory_limit, compilation_error
	ExecutionTimeMs int64     `json:"execution_time_ms"`
	MemoryUsedKB    int64     `json:"memory_used_kb"`
	TestsPassed     int       `json:"tests_passed"`
	TestsTotal      int       `json:"tests_total"`
	ErrorMessage    string    `json:"error_message,omitempty" gorm:"type:text"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// SubmissionStatus constants
const (
	StatusPending          = "pending"
	StatusRunning          = "running"
	StatusAccepted         = "accepted"
	StatusWrongAnswer      = "wrong_answer"
	StatusRuntimeError     = "runtime_error"
	StatusTimeLimit        = "time_limit"
	StatusMemoryLimit      = "memory_limit"
	StatusCompilationError = "compilation_error"
)

// Language constants
const (
	LangPython     = "python"
	LangJavaScript = "javascript"
	LangJava       = "java"
	LangCPP        = "cpp"
	LangRust       = "rust"
	LangGo         = "go"
)

// ValidLanguages lists all supported programming languages
var ValidLanguages = []string{
	LangPython,
	LangJavaScript,
	LangJava,
	LangCPP,
	LangRust,
	LangGo,
}

// IsValidLanguage checks if a language is supported
func IsValidLanguage(lang string) bool {
	for _, l := range ValidLanguages {
		if l == lang {
			return true
		}
	}
	return false
}
