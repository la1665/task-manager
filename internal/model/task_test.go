package model

import "testing"

func TestTaskStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status TaskStatus
		want   bool
	}{
		{"pending is valid", TaskStatusPending, true},
		{"in_progress is valid", TaskStatusInProgress, true},
		{"done is valid", TaskStatusDone, true},
		{"unknown value is invalid", TaskStatus("unknown"), false},
		{"empty value is invalid", TaskStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskPriority_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		priority TaskPriority
		want     bool
	}{
		{"low is valid", TaskPriorityLow, true},
		{"medium is valid", TaskPriorityMedium, true},
		{"high is valid", TaskPriorityHigh, true},
		{"unknown value is invalid", TaskPriority("urgent"), false},
		{"empty value is invalid", TaskPriority(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.priority.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
