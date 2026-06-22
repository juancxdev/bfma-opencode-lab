package memory

import "testing"

func TestCalculateScoreSeparatesAntecedentAndBFMAExtension(t *testing.T) {
	weights := ScoreWeights{Relevance: 0.25, Importance: 0.30, Recency: 0.15, Frequency: 0.10, TokenCost: 0.10, Obsolescence: 0.10}
	components := ScoreComponents{Relevance: 0.4, Importance: 0.9, Recency: 0.5, Frequency: 0.2, TokenCost: 0.1, Obsolescence: 0.0}
	got := CalculateScore(components, weights)
	if got.FormulaVersion != BFMAFormulaVersion {
		t.Fatalf("formula version = %q, want %q", got.FormulaVersion, BFMAFormulaVersion)
	}
	if got.AntecedentScore != 0.445 {
		t.Fatalf("antecedent = %.3f, want 0.445", got.AntecedentScore)
	}
	if got.BFMAUtility != 0.455 {
		t.Fatalf("bfma utility = %.3f, want 0.455", got.BFMAUtility)
	}
}

func TestDecideForgetReasons(t *testing.T) {
	tests := []struct {
		name         string
		utility      float64
		usedTokens   int
		cost         int
		budget       int
		obsolescence ObsolescenceAssessment
		wantDecision string
		wantReason   string
	}{
		{name: "obsolete", utility: 0.9, usedTokens: 0, cost: 1, budget: 10, obsolescence: ObsolescenceAssessment{Score: 1}, wantDecision: "discard", wantReason: "obsolete_replaced"},
		{name: "below score", utility: 0.1, usedTokens: 0, cost: 1, budget: 10, wantDecision: "discard", wantReason: "below_min_score"},
		{name: "budget exceeded", utility: 0.5, usedTokens: 9, cost: 2, budget: 10, wantDecision: "discard", wantReason: "token_budget_exceeded"},
		{name: "keep", utility: 0.5, usedTokens: 2, cost: 2, budget: 10, wantDecision: "keep", wantReason: "within_budget"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DecideForget(tt.utility, 0.28, tt.usedTokens, tt.cost, tt.budget, tt.obsolescence)
			if got.Decision != tt.wantDecision || got.Reason != tt.wantReason {
				t.Fatalf("decision = %s/%s, want %s/%s", got.Decision, got.Reason, tt.wantDecision, tt.wantReason)
			}
		})
	}
}

func TestAssessObsolescenceDetectsReplacementByNewerTaggedMemory(t *testing.T) {
	old := MemoryRecord{ID: "mem_001", Content: "La carga horaria pertenece al módulo de docentes.", SourceTurn: 3, Tags: []string{"docentes", "carga horaria"}}
	newer := MemoryRecord{ID: "mem_002", Content: "La carga horaria ya no pertenece al módulo de docentes; se reemplaza por planificación académica.", SourceTurn: 8, Tags: []string{"docentes", "carga horaria"}}
	got := AssessObsolescence(old, []MemoryRecord{old, newer})
	if got.Score != 1.0 || got.Reason != "replaced_by_mem_002" {
		t.Fatalf("obsolescence = %.2f/%q, want 1.0/replaced_by_mem_002", got.Score, got.Reason)
	}
}

func TestAssessObsolescenceReplacementCases(t *testing.T) {
	tests := []struct {
		name       string
		old        MemoryRecord
		newer      MemoryRecord
		wantScore  float64
		wantReason string
	}{
		{
			name:       "valid carga horaria replacement",
			old:        MemoryRecord{ID: "mem_001", Content: "La carga horaria pertenece al módulo de docentes.", SourceTurn: 3, Tags: []string{"docentes", "carga horaria"}},
			newer:      MemoryRecord{ID: "mem_002", Content: "La carga horaria ya no pertenece a docentes; pasa a planificación académica.", SourceTurn: 8, Tags: []string{"docentes", "carga horaria"}},
			wantScore:  1,
			wantReason: "replaced_by_mem_002",
		},
		{
			name:       "valid prerequisitos replacement",
			old:        MemoryRecord{ID: "mem_010", Content: "Los prerrequisitos se validan al registrar notas.", SourceTurn: 2, Tags: []string{"prerrequisitos", "notas"}},
			newer:      MemoryRecord{ID: "mem_011", Content: "La validación de prerrequisitos cambia; ahora debe ejecutarse antes de matrícula.", SourceTurn: 7, Tags: []string{"prerrequisitos", "matricula", "cambio"}},
			wantScore:  1,
			wantReason: "replaced_by_mem_011",
		},
		{
			name:       "generic docente tag is not enough",
			old:        MemoryRecord{ID: "mem_020", Content: "El correo institucional del docente será obligatorio.", SourceTurn: 3, Tags: []string{"docentes", "correo institucional"}},
			newer:      MemoryRecord{ID: "mem_021", Content: "La carga horaria ya no pertenece al módulo de docentes.", SourceTurn: 8, Tags: []string{"docentes", "carga horaria"}},
			wantScore:  0,
			wantReason: "",
		},
		{
			name:       "notas tag with low overlap is not enough",
			old:        MemoryRecord{ID: "mem_030", Content: "El cierre académico bloquea cambios posteriores en el registro de notas.", SourceTurn: 2, Tags: []string{"notas", "cierre academico"}},
			newer:      MemoryRecord{ID: "mem_031", Content: "La regla anterior de validar prerrequisitos recién al registrar notas ya no aplica.", SourceTurn: 7, Tags: []string{"prerrequisitos", "notas", "obsoleto"}},
			wantScore:  0,
			wantReason: "",
		},
		{
			name:       "architecture tag is generic only",
			old:        MemoryRecord{ID: "mem_040", Content: "La arquitectura final debe integrar matrícula y notas.", SourceTurn: 13, Tags: []string{"arquitectura", "modulos"}},
			newer:      MemoryRecord{ID: "mem_041", Content: "La arquitectura académica debe priorizar reglas vigentes sobre decisiones iniciales reemplazadas.", SourceTurn: 16, Tags: []string{"arquitectura", "vigente"}},
			wantScore:  0,
			wantReason: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AssessObsolescence(tt.old, []MemoryRecord{tt.old, tt.newer})
			if got.Score != tt.wantScore || got.Reason != tt.wantReason {
				t.Fatalf("obsolescence = %.2f/%q, want %.2f/%q", got.Score, got.Reason, tt.wantScore, tt.wantReason)
			}
		})
	}
}

func TestAssessObsolescenceSelfMarkedObsolete(t *testing.T) {
	record := MemoryRecord{ID: "mem_001", Content: "La regla anterior ya no aplica para la matrícula académica.", SourceTurn: 4, Tags: []string{"matricula"}}
	got := AssessObsolescence(record, []MemoryRecord{record})
	if got.Score != 0.75 || got.Reason != "self_marked_obsolete" {
		t.Fatalf("obsolescence = %.2f/%q, want 0.75/self_marked_obsolete", got.Score, got.Reason)
	}
}

func TestIsObsoleteByReplacementReasons(t *testing.T) {
	defaultOld := MemoryRecord{ID: "mem_001", Content: "El correo institucional del docente será obligatorio.", SourceTurn: 3, Tags: []string{"docentes", "correo institucional"}}

	tests := []struct {
		name       string
		old        MemoryRecord
		newer      MemoryRecord
		wantReason string
	}{
		{
			name:       "no replacement signal",
			newer:      MemoryRecord{ID: "mem_002", Content: "El correo institucional del docente será obligatorio para reportes.", SourceTurn: 4, Tags: []string{"docentes", "correo institucional"}},
			wantReason: "not_obsolete_no_replacement_signal",
		},
		{
			name:       "generic tag only",
			newer:      MemoryRecord{ID: "mem_003", Content: "La carga horaria ya no pertenece al módulo de docentes.", SourceTurn: 8, Tags: []string{"docentes"}},
			wantReason: "not_obsolete_generic_tag_only",
		},
		{
			name:       "low overlap",
			old:        MemoryRecord{ID: "mem_010", Content: "El cierre académico bloquea cambios posteriores en el registro de notas.", SourceTurn: 2, Tags: []string{"notas", "cierre academico"}},
			newer:      MemoryRecord{ID: "mem_011", Content: "La regla anterior de validar prerrequisitos recién al registrar notas ya no aplica.", SourceTurn: 7, Tags: []string{"prerrequisitos", "notas", "obsoleto"}},
			wantReason: "not_obsolete_low_overlap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := tt.old
			if old.ID == "" {
				old = defaultOld
			}
			obsolete, reason := isObsoleteByReplacement(old, tt.newer)
			if obsolete || reason != tt.wantReason {
				t.Fatalf("obsolete/reason = %v/%q, want false/%q", obsolete, reason, tt.wantReason)
			}
		})
	}
}

func TestReplacementHelpers(t *testing.T) {
	if !replacementSignal("La carga horaria ya no pertenece a docentes") {
		t.Fatal("expected replacement signal")
	}
	if isGenericTag("carga horaria") {
		t.Fatal("carga horaria should be specific")
	}
	if !isGenericTag("docentes") {
		t.Fatal("docentes should be generic")
	}
	shared := sharedSpecificTags([]string{"docentes", "carga horaria"}, []string{"docentes", "carga horaria"})
	if len(shared) != 1 || shared[0] != "carga horaria" {
		t.Fatalf("shared specific tags = %#v, want [carga horaria]", shared)
	}
}
