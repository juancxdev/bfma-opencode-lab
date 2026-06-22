package evaluation

import "testing"

func TestFactualF1(t *testing.T) {
	tests := []struct {
		name      string
		answer    string
		expected  string
		minRecall float64
		wantSubEM bool
	}{
		{
			name:      "exact expected phrase",
			answer:    "La autenticación debe ser con correo institucional.",
			expected:  "correo institucional",
			minRecall: 1,
			wantSubEM: true,
		},
		{
			name:      "partial factual recovery",
			answer:    "La matrícula valida prerrequisitos.",
			expected:  "validación de prerrequisitos antes de matrícula",
			minRecall: 0.5,
			wantSubEM: false,
		},
		{
			name:      "missing fact",
			answer:    "Respuesta sin relación.",
			expected:  "correo institucional",
			minRecall: 0,
			wantSubEM: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, recall, _ := factualF1(tt.answer, tt.expected, 0)
			if recall < tt.minRecall {
				t.Fatalf("recall = %.4f, want >= %.4f", recall, tt.minRecall)
			}
			if got := substringExactMatch(tt.answer, tt.expected); got != tt.wantSubEM {
				t.Fatalf("SubEM = %v, want %v", got, tt.wantSubEM)
			}
		})
	}
}

func TestCountDetectedNormalizesAccents(t *testing.T) {
	answer := "No debe gestionar carga horaria dentro de docentes y no debe usar el inventario de laptops como módulo académico."
	phrases := []string{"gestionar carga horaria dentro de docentes", "inventario de laptops como módulo académico", "garantía de laptops"}
	if got := countDetected(answer, phrases); got != 2 {
		t.Fatalf("countDetected = %d, want 2", got)
	}
}
