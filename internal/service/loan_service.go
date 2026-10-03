package service

import "math"

// LoanCalculation hasil perhitungan kredit kendaraan.
type LoanCalculation struct {
	OTR                int64   `json:"otr"`
	DP                 int64   `json:"dp"`
	Principal          int64   `json:"principal"`
	TenorMonths        int     `json:"tenor_months"`
	InterestRate       float64 `json:"interest_rate"`
	InterestAmount     int64   `json:"interest_amount"`
	TotalAmount        int64   `json:"total_amount"`
	MonthlyInstallment int64   `json:"monthly_installment"`
}

// DetermineInterestRate menentukan bunga berdasarkan tenor.
//
//	<= 12 bulan  : 12%
//	<= 24 bulan  : 14%
//	>  24 bulan  : 16.5%
func DetermineInterestRate(tenorMonths int) float64 {
	switch {
	case tenorMonths <= 12:
		return 0.12
	case tenorMonths <= 24:
		return 0.14
	default:
		return 0.165
	}
}

// CalculateLoan menghitung seluruh komponen kredit berdasarkan OTR, DP%, dan tenor.
// Nilai uang dikembalikan dalam satuan rupiah (int64).
// Angsuran dibulatkan ke ribuan terdekat.
func CalculateLoan(otr int64, dpPercentage float64, tenorMonths int) LoanCalculation {
	// DP = OTR × DP%
	dp := roundToThousand(float64(otr) * dpPercentage / 100)

	// Pokok Utang = OTR - DP
	principal := otr - dp

	// Bunga berdasarkan tenor
	interestRate := DetermineInterestRate(tenorMonths)

	// Total Bunga = Pokok × Bunga × (Tenor / 12)
	interestAmount := roundToThousand(float64(principal) * interestRate * float64(tenorMonths) / 12)

	// Total Kredit = Pokok + Bunga
	totalAmount := principal + interestAmount

	// Angsuran = Total Kredit / Tenor (dibulatkan ke ribuan)
	monthlyInstallment := roundToThousand(float64(totalAmount) / float64(tenorMonths))

	return LoanCalculation{
		OTR:                otr,
		DP:                 dp,
		Principal:          principal,
		TenorMonths:        tenorMonths,
		InterestRate:       interestRate,
		InterestAmount:     interestAmount,
		TotalAmount:        totalAmount,
		MonthlyInstallment: monthlyInstallment,
	}
}

// CalculatePenalty menghitung denda keterlambatan.
//
//	Denda = Sisa Angsuran × 0.001 × Hari Keterlambatan
func CalculatePenalty(remainingAmount int64, daysLate int) int64 {
	return int64(float64(remainingAmount) * 0.001 * float64(daysLate))
}

// roundToThousand membulatkan nilai ke ribuan terdekat.
func roundToThousand(v float64) int64 {
	return int64(math.Round(v/1000) * 1000)
}
