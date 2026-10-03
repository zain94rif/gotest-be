package service

import "testing"

func TestDetermineInterestRate(t *testing.T) {
	tests := []struct {
		tenor    int
		expected float64
	}{
		{12, 0.12},
		{18, 0.14},
		{24, 0.14},
		{25, 0.165},
		{36, 0.165},
	}

	for _, tt := range tests {
		got := DetermineInterestRate(tt.tenor)
		if got != tt.expected {
			t.Errorf("DetermineInterestRate(%d) = %.4f, want %.4f", tt.tenor, got, tt.expected)
		}
	}
}

func TestCalculateLoan(t *testing.T) {
	// Data Pak Sugus: OTR=240jt, DP=20%, tenor=18
	result := CalculateLoan(240_000_000, 20, 18)

	if result.DP != 48_000_000 {
		t.Errorf("DP = %d, want 48000000", result.DP)
	}
	if result.Principal != 192_000_000 {
		t.Errorf("Principal = %d, want 192000000", result.Principal)
	}
	if result.InterestRate != 0.14 {
		t.Errorf("InterestRate = %.4f, want 0.14", result.InterestRate)
	}
	if result.InterestAmount != 40_320_000 {
		t.Errorf("InterestAmount = %d, want 40320000", result.InterestAmount)
	}
	if result.TotalAmount != 232_320_000 {
		t.Errorf("TotalAmount = %d, want 232320000", result.TotalAmount)
	}
	if result.MonthlyInstallment != 12_907_000 {
		t.Errorf("MonthlyInstallment = %d, want 12907000", result.MonthlyInstallment)
	}
}

func TestCalculatePenalty(t *testing.T) {
	// 12.907.000 × 0.001 × 50 = 645.350
	got := CalculatePenalty(12_907_000, 50)
	if got != 645_350 {
		t.Errorf("CalculatePenalty(12907000, 50) = %d, want 645350", got)
	}

	// 12.907.000 × 0.001 × 20 = 258.140
	got2 := CalculatePenalty(12_907_000, 20)
	if got2 != 258_140 {
		t.Errorf("CalculatePenalty(12907000, 20) = %d, want 258140", got2)
	}
}
