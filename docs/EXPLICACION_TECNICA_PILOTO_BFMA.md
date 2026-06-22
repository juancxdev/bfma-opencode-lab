# Explicación técnica del piloto BFMA OpenCode Lab

## 1. Propósito del piloto

Este piloto implementa un **demo técnico controlado** para comparar tres modalidades de gestión de memoria en agentes autónomos basados en LLM:

| Grupo | Modalidad | Rol metodológico |
|---|---|---|
| G1 | Arquitectura basada en resumen incremental | Comparador convencional basado en compresión progresiva del historial. |
| G2 | Arquitectura de memoria persistente acumulativa | Comparador convencional que conserva memorias y recupera por relevancia textual. |
| G3 | Arquitectura BFMA con olvido presupuestado | Modalidad experimental central. Selecciona memorias bajo una función de utilidad y un presupuesto de tokens. |

El piloto **no debe presentarse como contraste confirmatorio de la hipótesis**. Su función actual es demostrar que el sistema puede:

1. ejecutar las tres arquitecturas bajo el mismo escenario;
2. aislar el comportamiento de cada grupo;
3. registrar trazas por turno;
4. observar decisiones internas de memoria;
5. generar métricas exploratorias alineadas con la matriz de consistencia;
6. producir evidencia técnica exportable para anexos, revisión del asesor o preparación del experimento final.

La hipótesis formal de la tesis exige posteriormente múltiples escenarios, réplicas emparejadas, orden aleatorizado y el modelo mixto:

```text
F1 ~ arquitectura + orden + (1 | escenario/réplica)
```

---

## 2. Comando de ejecución del piloto

El flujo base se inicia con:

```bash
go run ./cmd/bfma-pilot \
  --scenario scenario_01 \
  --groups g1,g2,g3 \
  --reps 1 \
  --model opencode-go/deepseek-v4-flash
```

Este comando ejecuta el programa definido en:

```text
cmd/bfma-pilot/main.go
```

Su responsabilidad es leer parámetros de línea de comandos, construir una configuración experimental y delegar la ejecución al runner.

### 2.1. Parámetros principales

| Parámetro | Significado |
|---|---|
| `--scenario scenario_01` | Carga el archivo `scenarios/scenario_01.json`. |
| `--groups g1,g2,g3` | Ejecuta las tres arquitecturas comparadoras. |
| `--reps 1` | Ejecuta una réplica por grupo. |
| `--model opencode-go/deepseek-v4-flash` | Usa ese modelo en OpenCode para generar respuestas reales. |
| `--simulate-llm` | Modo simulado; valida plomería, pero no produce evidencia experimental válida. |
| `--timeout` | Tiempo máximo por turno de OpenCode. |
| `--retries` | Reintentos por turno fallido. |
| `--from-turn` | Permite retomar desde un turno reconstruyendo memoria previa sin volver a llamar al LLM para turnos anteriores. |

---

## 3. Construcción de la configuración interna

`cmd/bfma-pilot/main.go` transforma los flags en un objeto `runner.Config`:

```go
cfg := runner.Config{
    ProjectDir:  projectDir,
    ScenarioDir: filepath.Join(projectDir, "scenarios"),
    LogDir:      filepath.Join(projectDir, "logs"),
    DataDir:     filepath.Join(projectDir, "data"),
    ScenarioID:  *scenarioID,
    Groups:      groups,
    Reps:        *reps,
    Model:       *model,
    DryRun:      *simulateLLM || *dryRunAlias,
    Timeout:     *timeout,
    Retries:     *retries,
    FromTurn:    *fromTurn,
}
```

Luego crea un runner:

```go
r := runner.New(cfg, opencode.NewCLIClient("opencode"))
```

El runner recibe dos cosas:

1. **La configuración experimental**: escenario, grupos, réplicas, modelo, logs, etc.
2. **Un cliente OpenCode**: responsable de invocar el modelo real mediante la CLI de OpenCode.

---

## 4. Carga del escenario

El runner carga el escenario desde:

```text
scenarios/<scenario_id>.json
```

Por ejemplo:

```text
scenarios/scenario_01.json
```

Un escenario contiene:

```json
{
  "scenario_id": "scenario_01",
  "title": "Sistema académico de matrícula",
  "description": "...",
  "turns": [...],
  "final_questions": [...],
  "ground_truth": {...}
}
```

### 4.1. Estructura de un turno

Cada turno contiene:

```json
{
  "turn": 1,
  "user_prompt": "...",
  "memory_hints": [
    {
      "content": "El sistema debe autenticar con correo institucional.",
      "type": "constraint",
      "importance": 0.95,
      "tags": ["autenticación", "correo institucional"]
    }
  ]
}
```

El campo `user_prompt` es lo que recibirá el agente como turno actual.

El campo `memory_hints` es fundamental: representa unidades informativas que el runner usa para actualizar la memoria interna después de cada turno. Estas unidades son la materia prima para G1, G2 y G3.

### 4.2. Ground truth

El escenario también define:

```json
"ground_truth": {
  "critical_facts": [...],
  "expected_constraints": [...],
  "expected_contradictions": [...]
}
```

Esto sirve para evaluación posterior. No controla directamente la respuesta del agente durante la ejecución. Se usa para medir si la respuesta final recupera hechos esperados, introduce contradicciones o arrastra información que debería quedar fuera.

---

## 5. Bucle experimental del runner

La ejecución principal ocurre en:

```text
internal/runner/runner.go
```

El método central es:

```go
func (r *Runner) Run(ctx context.Context) (Result, error)
```

Este método hace:

1. cargar escenario;
2. crear un `run_id` con timestamp;
3. iterar por cada grupo solicitado;
4. iterar por cada réplica;
5. ejecutar `runGroupRep` para cada combinación grupo-réplica.

Conceptualmente:

```text
for group in [g1, g2, g3]:
  for rep in [1..reps]:
    runGroupRep(group, rep)
```

Si se ejecuta:

```bash
--groups g1,g2,g3 --reps 1
```

entonces se generan tres ejecuciones:

```text
g1 / réplica 1
g2 / réplica 1
g3 / réplica 1
```

---

## 6. Selección del administrador de memoria por grupo

Antes de ejecutar los turnos de un grupo, el runner llama:

```go
r.newManager(runID, group)
```

Esto instancia una arquitectura distinta según el grupo.

### 6.1. G1: resumen incremental

```go
return memory.NewIncrementalSummary(8), "g1-summary-agent", nil
```

G1 usa:

```text
internal/memory/summary.go
```

Características:

- mantiene un resumen acumulado;
- incorpora `memory_hints` y parte de la respuesta del agente;
- conserva como máximo 8 ítems;
- si supera el límite, descarta los ítems más antiguos;
- representa el riesgo de pérdida factual por compresión sucesiva.

El contexto que recibe el agente se parece a:

```text
Resumen acumulado de la conversación:
- ...
- ...
```

### 6.2. G2: memoria persistente acumulativa

```go
store := filepath.Join(r.cfg.DataDir, runID, string(group), "memories.jsonl")
return memory.NewPersistent(store, 8), "g2-persistent-agent", nil
```

G2 usa:

```text
internal/memory/persistent.go
```

Características:

- guarda cada `memory_hint` como un registro independiente;
- persiste memorias en `data/<run_id>/g2/memories.jsonl`;
- recupera hasta 8 memorias por relevancia textual;
- representa una arquitectura acumulativa con riesgo de arrastrar ruido u obsolescencia.

El contexto que recibe el agente se parece a:

```text
Memorias persistentes recuperadas:
- [mem_001] ...
- [mem_002] ...
```

### 6.3. G3: arquitectura BFMA

```go
return memory.NewBFMA(220, 0.28), "g3-bfma-agent", nil
```

G3 usa:

```text
internal/memory/bfma.go
```

Parámetros actuales:

| Parámetro | Valor | Significado |
|---|---:|---|
| `tokenBudget` | 220 | Presupuesto máximo aproximado de tokens para contexto seleccionado. |
| `keepMinScore` | 0.28 | Umbral mínimo de utilidad para que una memoria pueda entrar al contexto. |

BFMA guarda memorias como registros, calcula utilidad para cada una en cada turno y decide si entran o no al contexto activo.

Importante: en la implementación actual, **BFMA no borra permanentemente la memoria del almacén**. El `discard` significa:

> exclusión del contexto activo bajo presupuesto de tokens.

No significa eliminación física definitiva.

---

## 7. Construcción del prompt enviado al agente

En cada turno, el runner hace:

```go
memoryBefore := manager.Snapshot()
memCtx, err := manager.BuildContext(turn)
prompt := BuildPrompt(string(group), memCtx.Text, turn.UserPrompt)
```

El prompt final enviado al agente tiene esta estructura:

```text
EXPERIMENT_GROUP: g1|g2|g3

MEMORY_CONTEXT:
<memoria construida por la arquitectura correspondiente>

CURRENT_USER_TURN:
<prompt del turno actual>

INSTRUCTIONS:
Responde únicamente con la información disponible en MEMORY_CONTEXT y CURRENT_USER_TURN. No incluyas métricas, JSON experimental ni trazas internas.
```

Esto es importante metodológicamente: el agente no decide qué memorias recuperar. La recuperación la controla el runner. El agente solo recibe el contexto ya preparado.

---

## 8. Ejecución con OpenCode

La llamada real al LLM ocurre en:

```text
internal/opencode/client.go
```

El cliente arma un comando parecido a:

```bash
opencode run \
  --format json \
  --pure \
  --agent <agent> \
  --dir <projectDir> \
  --model opencode-go/deepseek-v4-flash \
  <prompt>
```

También establece:

```text
OPENCODE_DISABLE_AUTOCOMPACT=true
```

Esto evita que OpenCode agregue compactación automática externa que contamine el comportamiento experimental.

### 8.1. Agentes OpenCode

Los agentes están definidos en:

```text
.opencode/agents/g1-summary-agent.md
.opencode/agents/g2-persistent-agent.md
.opencode/agents/g3-bfma-agent.md
```

Cada agente tiene reglas que restringen su comportamiento. Por ejemplo:

- no usar herramientas externas;
- no inventar información fuera del `MEMORY_CONTEXT` y el turno actual;
- no devolver métricas ni trazas internas;
- responder de forma natural.

El archivo `.opencode/opencode.json` desactiva herramientas como `bash`, `edit`, `write` y `read` para los agentes experimentales.

---

## 9. Actualización de memoria después de cada respuesta

Cuando OpenCode devuelve una respuesta, el runner ejecuta:

```go
postEvents, err := manager.Observe(turn, memory.AgentResponse{Text: resp.Text})
```

Esto actualiza la memoria de acuerdo con la arquitectura.

### 9.1. En G1

G1 agrega al resumen:

- los `memory_hints` del turno;
- una frase resumida de la respuesta del agente;
- conserva solo los últimos ítems hasta `maxItems`.

Esto simula una memoria comprimida por resumen incremental.

### 9.2. En G2

G2 guarda cada `memory_hint` como memoria persistente:

```json
{
  "id": "mem_001",
  "content": "...",
  "type": "constraint",
  "source_turn": 1,
  "importance": 0.95,
  "tags": [...]
}
```

La memoria crece acumulativamente.

### 9.3. En G3

G3 también guarda memorias, pero en cada `BuildContext` decide cuáles entran al contexto mediante BFMA.

La implementación refactorizada separa el cálculo en dos etapas:

```text
AntecedentScore =
  wR * relevance
+ wI * importance
+ wC * recency

BFMAUtility =
  AntecedentScore
+ wF * frequency
- wT * token_cost
- wO * obsolescence
```

La versión de fórmula registrada en los logs es `bfma_base_extension_v1`.

Donde:

| Componente | Descripción |
|---|---|
| `importance` | Importancia asignada en el escenario. |
| `relevance` | Solapamiento textual entre prompt actual y contenido de la memoria. |
| `recency` | `1 / (edad + 1)`, favorece memorias recientes. |
| `frequency` | Frecuencia normalizada de aparición/uso. |
| `token_cost` | Penalización por costo aproximado de tokens. |
| `obsolescence` | Penalización por relación de reemplazo detectada entre una memoria anterior y otra posterior con etiquetas compartidas. |

La decisión es:

```text
si utility >= keepMinScore y todavía cabe en tokenBudget:
    keep
si no:
    discard
```

Eventos generados:

```json
{
  "type": "bfma_decision",
  "memory_id": "mem_001",
  "decision": "keep",
  "utility_score": 0.742,
  "components": {
    "importance": 0.95,
    "relevance": 0.31,
    "recency": 0.50,
    "frequency": 0.20,
    "token_cost": 0.08,
    "obsolescence": 0.00
  },
  "reason": "within_budget"
}
```

Razones posibles:

| Razón | Significado |
|---|---|
| `within_budget` | La memoria supera el umbral y entra en el presupuesto. |
| `below_min_score` | La utilidad no supera el umbral mínimo. |
| `token_budget_exceeded` | La utilidad es suficiente, pero no cabe en el presupuesto restante. |

---


### 9.4. Configuración de pesos BFMA desde CLI

El runner recibe una `BFMAConfig` con presupuesto, umbral y pesos. Los valores por defecto son:

| Parámetro | Valor |
|---|---:|
| `--bfma-token-budget` | `220` |
| `--bfma-keep-min-score` | `0.28` |
| `--bfma-weight-relevance` | `0.25` |
| `--bfma-weight-importance` | `0.30` |
| `--bfma-weight-recency` | `0.15` |
| `--bfma-weight-frequency` | `0.10` |
| `--bfma-weight-token-cost` | `0.10` |
| `--bfma-weight-obsolescence` | `0.10` |

Esto permite documentar y replicar la configuración usada en cada corrida. Los pesos también quedan registrados en cada evento `bfma_decision` dentro de los logs JSONL.

## 10. Registro de logs JSONL

Cada turno genera una línea JSON en archivos como:

```text
logs/<run_id>/<group>/<scenario_id>_rep_01.jsonl
```

Ejemplo:

```text
logs/run_20260615_172800/g3/scenario_03_gestion_academica_longitudinal_rep_01.jsonl
```

Cada línea contiene una estructura `TurnLog`:

```go
type TurnLog struct {
    RunID                 string
    Status                string
    Group                 string
    ScenarioID            string
    Rep                   int
    Turn                  int
    Attempt               int
    MaxAttempts           int
    Agent                 string
    MemoryContextInjected string
    UserPrompt            string
    AssistantResponse     string
    Error                 string
    TimeoutMS             int64
    LatencyMS             int64
    OpenCodeEvents        []map[string]any
    MemoryBefore          memory.Snapshot
    MemoryAfter           memory.Snapshot
    MemoryEvents          []memory.Event
    MemoryPreEvents       []memory.Event
    MemoryPostEvents      []memory.Event
    FinalQuestions        []scenario.FinalQuestion
}
```

Estos logs son la fuente principal para:

1. el reporte HTML;
2. el evaluador exploratorio;
3. auditoría de memoria;
4. trazabilidad del demo.

---

## 11. Reintentos y fallos

Si un turno falla, el runner registra un log con:

```json
"status": "failed"
```

También conserva:

- intento actual;
- total de intentos;
- diagnóstico del error;
- memoria antes/después sin mutar.

Esto es importante porque las ejecuciones inválidas no desaparecen. Quedan trazadas.

---

## 12. Uso de `--from-turn`

Para escenarios largos, se puede retomar desde un turno:

```bash
go run ./cmd/bfma-pilot \
  --scenario scenario_03_gestion_academica_longitudinal \
  --groups g1,g2,g3 \
  --from-turn 11 \
  --reps 1 \
  --model opencode-go/deepseek-v4-flash \
  --retries 2 \
  --timeout 10m
```

Internamente, el runner ejecuta:

```go
warmMemoryUntilTurn(sc, manager)
```

Esto reconstruye la memoria previa hasta el turno anterior usando los `memory_hints`, pero **sin llamar al LLM ni generar logs de esos turnos previos**.

Así se puede retomar desde turnos avanzados sin pagar todo el costo del escenario completo.

Limitación metodológica: una corrida con `--from-turn` es útil para demo técnico, pero debe documentarse porque no contiene todas las interacciones reales desde el turno 1 en el log final.

---

## 13. Reporte HTML original

El comando:

```bash
go run ./cmd/bfma-report \
  --run logs/run_20260615_172800 \
  --out reports/run_20260615_172800/index.html
```

usa:

```text
cmd/bfma-report/main.go
internal/report/report.go
internal/report/template.go
```

El reporte carga los logs JSONL y genera un HTML autocontenido.

### 13.1. Qué muestra el reporte

El reporte visualiza:

- grupos ejecutados;
- turnos exitosos/fallidos;
- latencia promedio;
- memoria final;
- decisiones BFMA `keep/discard`;
- crecimiento de memoria;
- razones de descarte;
- contexto final inyectado;
- respuesta final por grupo;
- cobertura automática estimada.

### 13.2. Cobertura automática estimada

Antes de agregar el evaluador, el reporte calculaba una cobertura simple mediante:

```go
coverageFor(...)
expectedAnswerDetected(...)
```

Este cálculo buscaba coincidencias textuales entre la respuesta final y las respuestas esperadas.

Esa cobertura sirve como ayuda visual, pero no debe tratarse como métrica confirmatoria.

---

## 14. Capa nueva agregada: `bfma-evaluate`

Para alinear el demo con la matriz de consistencia, se agregó un comando nuevo:

```bash
go run ./cmd/bfma-evaluate \
  --run logs/run_20260615_172800 \
  --out results/run_20260615_172800
```

Archivos principales:

```text
cmd/bfma-evaluate/main.go
internal/evaluation/evaluation.go
internal/evaluation/evaluation_test.go
```

Esta capa no reemplaza el runner. Lee los logs ya generados y produce métricas exploratorias.

---

## 15. Artefactos generados por `bfma-evaluate`

El evaluador genera:

```text
results/run_20260615_172800/metrics.csv
results/run_20260615_172800/summary.json
results/run_20260615_172800/conclusion.md
```

### 15.1. `metrics.csv`

Dataset tabular por:

- arquitectura;
- réplica;
- turno final;
- pregunta final;
- métricas calculadas.

Columnas principales:

```text
run_id
scenario_id
architecture
replica
turn
question_id
expected_answer
f1
precision
recall
subem
subem_hits
subem_total
false_memory_hits
false_memory_total
contradiction_hits
contradiction_total
memory_tokens
output_tokens
storage_items
latency_ms
valid_execution
exclusion_reason
```

### 15.2. `summary.json`

Resumen agregado por arquitectura:

```json
{
  "architecture": "g3",
  "avg_f1": 0.7148,
  "avg_recall": 0.6205,
  "subem_hits": 6,
  "subem_total": 10,
  "false_memory_hits": 0,
  "false_memory_total": 5,
  "contradiction_hits": 0,
  "contradiction_total": 5,
  "avg_memory_tokens": 262,
  "avg_output_tokens": 535,
  "avg_latency_ms": 21071
}
```

### 15.3. `conclusion.md`

Informe corto para leer o anexar:

```md
# Evaluación exploratoria del demo técnico BFMA

## Lectura metodológica
...

## Resumen por arquitectura
...
```

---

## 16. Métricas exploratorias implementadas

### 16.1. F1 factual

El F1 factual compara la respuesta del agente con cada respuesta esperada.

El cálculo usa tokens normalizados:

1. normaliza acentos y puntuación;
2. elimina palabras muy cortas y stopwords básicas;
3. cuenta tokens esperados presentes en la respuesta;
4. calcula precisión, recall y F1.

Para respuestas arquitectónicas largas, no se penaliza cada palabra extra como falso positivo. Los falsos positivos se operacionalizan mediante:

- falsas memorias detectadas;
- contradicciones detectadas.

Esto evita que una respuesta larga y correcta sea castigada solo por tener explicación adicional.

### 16.2. Recall factual

Mide qué proporción de los tokens esperados fue recuperada.

Ejemplo:

```text
expected_answer: "Antes de permitir o registrar la matrícula académica."
respuesta: "La validación de prerrequisitos ocurre antes de la matrícula."
```

Aunque no sea frase exacta, puede recuperar tokens centrales y obtener recall parcial.

### 16.3. SubEM por entidades críticas

Inicialmente SubEM comparaba la respuesta esperada completa como substring exacto. Eso producía cero para todos los grupos, porque las respuestas del agente podían expresar lo correcto con otro orden o redacción.

Se corrigió para medir entidades críticas definidas en el escenario:

```json
"subem_entities": [
  "Sistema de Gestión Académica Universitaria para la FIIS",
  "registro de notas",
  "docentes",
  "estudiantes",
  "plan de estudios",
  "matrícula académica",
  "correo institucional",
  "auditoría común",
  "prerrequisitos antes de matrícula",
  "rectificación aprobada por coordinador"
]
```

Ahora el resumen muestra, por ejemplo:

```text
g3: 6/10 (0.6000)
```

Eso significa que BFMA recuperó textualmente 6 de 10 entidades críticas definidas para el escenario.

### 16.4. Falsas memorias

El escenario define trampas explícitas:

```json
"false_memory_traps": [
  "inventario de laptops como módulo académico",
  "código patrimonial como entidad académica",
  "garantía de laptops dentro del sistema académico",
  "proveedor de laptops dentro del sistema académico",
  "colores del inventario para definir la arquitectura académica"
]
```

Si la respuesta final contiene alguna de estas frases o equivalentes textuales simples, se cuenta como falsa memoria detectada.

El resumen muestra:

```text
0/5 (0.0000)
```

Esto significa:

> Se buscaron 5 trampas explícitas y no se detectó ninguna.

No significa que el detector sea perfecto. Significa que no hubo coincidencia contra las trampas definidas.

### 16.5. Contradicciones

Se usan las contradicciones prohibidas del `ground_truth`:

```json
"expected_contradictions": [
  "validar prerrequisitos recién al registrar notas",
  "gestionar carga horaria dentro de docentes",
  "permitir edición libre de notas después del cierre académico",
  "usar colores o reglas del inventario de laptops para definir la arquitectura académica",
  "tratar el inventario de laptops como módulo académico"
]
```

El resumen muestra:

```text
0/5 (0.0000)
```

Esto significa:

> Se buscaron 5 contradicciones explícitas y no se detectó ninguna.

### 16.6. Tokens de memoria

Se estima el tamaño del `MEMORY_CONTEXT` inyectado al agente mediante una aproximación basada en palabras.

No es tokenización exacta del proveedor, pero sirve como métrica exploratoria y comparable entre grupos.

### 16.7. Tokens de salida

Se estima la longitud de la respuesta final del agente.

Esto permite observar si una arquitectura produce respuestas más largas o más compactas.

### 16.8. Almacenamiento

Se mide como cantidad de ítems almacenados o representados en memoria:

- para G1: líneas del resumen;
- para G2/G3: cantidad de registros de memoria.

### 16.9. Latencia

Se toma desde la respuesta del cliente OpenCode:

```go
Latency: time.Since(started)
```

Se reporta en milisegundos.

---

## 17. Integración del evaluador con el reporte HTML

El reporte HTML fue actualizado para llamar internamente al evaluador cuando carga un run:

```go
if evalResult, evalErr := evaluation.Evaluate(evaluation.Options{RunDir: runDir}); evalErr == nil {
    data.EvaluationAvailable = true
    data.EvaluationGroups = evalResult.Groups
    data.EvaluationConclusion = evalResult.Conclusion
    data.EvaluationWarnings = evalResult.Warnings
}
```

Gracias a eso, el HTML ahora incluye una sección:

```text
Métricas exploratorias alineadas a la matriz de consistencia
```

Esa sección muestra:

- F1;
- SubEM por entidades;
- falsas memorias detectadas;
- contradicciones detectadas;
- tokens de memoria;
- almacenamiento;
- latencia;
- lectura metodológica.

---

## 18. Resultado actual del run `run_20260615_172800`

El resumen actual es:

| Arquitectura | F1 | Recall | SubEM entidades | Falsas memorias detectadas | Contradicciones detectadas | Tokens memoria | Tokens salida | Latencia ms |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| g1 | 0.6974 | 0.5746 | 5/10 (0.5000) | 0/5 (0.0000) | 0/5 (0.0000) | 153.0 | 793.0 | 28695.0 |
| g2 | 0.6878 | 0.5621 | 7/10 (0.7000) | 0/5 (0.0000) | 0/5 (0.0000) | 163.0 | 1236.0 | 41349.0 |
| g3 | 0.7148 | 0.6205 | 6/10 (0.6000) | 0/5 (0.0000) | 0/5 (0.0000) | 262.0 | 535.0 | 21071.0 |

Lectura técnica prudente:

- BFMA obtiene el mayor F1 exploratorio en esta corrida.
- BFMA obtiene el mayor recall factual en esta corrida.
- G2 obtiene mayor SubEM de entidades exactas, probablemente por acumular más información textual.
- BFMA produce una respuesta final más compacta que G1 y G2.
- G2 tiene mayor latencia y mayor longitud de salida.
- No se detectaron falsas memorias ni contradicciones contra las trampas explícitas definidas.

Lectura metodológica correcta:

> Este resultado valida que el pipeline puede generar métricas alineadas a la tesis, pero no permite concluir superioridad estadística de BFMA porque corresponde a un solo escenario y una réplica.

---

## 19. Relación con la matriz de consistencia de la tesis

| Elemento del plan de tesis | Implementación actual en el piloto |
|---|---|
| Variable independiente: arquitectura de gestión de memoria | `group`: `g1`, `g2`, `g3`. |
| Modalidad G1 | `IncrementalSummary`. |
| Modalidad G2 | `Persistent`. |
| Modalidad G3 | `BFMA`. |
| Resultado primario: F1 factual | `internal/evaluation/factualF1`. |
| SubEM | Entidades críticas en `evaluation.subem_entities`. |
| Falsas memorias | `evaluation.false_memory_traps`. |
| Contradicciones | `ground_truth.expected_contradictions`. |
| Tokens | Estimación de tokens de memoria y salida. |
| Almacenamiento | Ítems de resumen o registros de memoria. |
| Latencia | `LatencyMS` registrado por OpenCode client. |
| Trazabilidad | Logs JSONL por turno, grupo y réplica. |
| Modelo mixto | Aún no implementado; requiere múltiples escenarios/réplicas. |

---

## 20. Limitaciones actuales

Este piloto todavía tiene limitaciones importantes:

1. **Un solo escenario no permite inferencia estadística.**
2. **Una réplica no permite estimar variabilidad real.**
3. **No hay aleatorización formal del orden de ejecución.**
4. **F1, SubEM, falsas memorias y contradicciones son métricas exploratorias automáticas.**
5. **No hay juicio de expertos ni V de Aiken sobre escenarios/preguntas/ground truth.**
6. **No hay evaluación ciega humana para falsas memorias o contradicciones semánticas.**
7. **El token count es aproximado, no tokenización exacta del modelo.**
8. **BFMA actualmente descarta del contexto activo, no elimina memorias permanentemente.**

Estas limitaciones no invalidan el demo técnico. Simplemente delimitan su alcance.

---

## 21. Qué demuestra y qué no demuestra

### 21.1. Sí demuestra

El piloto sí demuestra que:

- las tres arquitecturas pueden ejecutarse bajo un mismo escenario;
- el runner controla el contexto entregado al agente;
- BFMA calcula utilidad por memoria;
- BFMA toma decisiones `keep/discard` bajo presupuesto;
- las decisiones quedan trazadas;
- las respuestas finales pueden evaluarse contra ground truth;
- se puede generar un dataset tabular para análisis posterior.

### 21.2. No demuestra todavía

El piloto todavía no demuestra que:

- BFMA sea estadísticamente superior;
- BFMA reduzca la deriva de resumen en población de escenarios;
- el efecto sea significativo;
- el efecto tenga tamaño suficiente;
- las métricas estén validadas por expertos;
- el resultado sea generalizable.

---

## 22. Narrativa recomendada para defensa

Una forma correcta de presentarlo sería:

> Este demo técnico implementa un entorno controlado para comparar tres modalidades de gestión de memoria: resumen incremental, memoria persistente acumulativa y arquitectura BFMA. El runner controla el contexto que recibe cada agente, registra trazas por turno y permite observar cómo BFMA selecciona memorias mediante una función de utilidad bajo presupuesto de tokens. La capa de evaluación agregada convierte los logs en métricas exploratorias alineadas con la matriz de consistencia, como F1 factual, SubEM por entidades críticas, falsas memorias, contradicciones, tokens, almacenamiento y latencia. Estos resultados validan la viabilidad técnica e instrumental del experimento, pero no constituyen todavía contraste confirmatorio de la hipótesis, porque se requiere ampliar a múltiples escenarios, réplicas emparejadas, orden aleatorizado y modelo mixto.

---

## 23. Flujo completo recomendado

Para ejecutar el demo técnico longitudinal:

```bash
go run ./cmd/bfma-pilot \
  --scenario scenario_03_gestion_academica_longitudinal \
  --groups g1,g2,g3 \
  --from-turn 11 \
  --reps 1 \
  --model opencode-go/deepseek-v4-flash \
  --retries 2 \
  --timeout 10m
```

Para generar métricas exploratorias:

```bash
go run ./cmd/bfma-evaluate \
  --run logs/run_20260615_172800 \
  --out results/run_20260615_172800
```

Para generar reporte HTML:

```bash
go run ./cmd/bfma-report \
  --run logs/run_20260615_172800 \
  --out reports/run_20260615_172800/index.html
```

---

## 24. Próximo paso para convertirlo en piloto metodológico

Para pasar de demo técnico a piloto metodológico, habría que agregar:

1. varios escenarios longitudinales;
2. banco de preguntas por escenario;
3. ground truth validado por expertos;
4. V de Aiken para escenarios, preguntas y respuestas esperadas;
5. réplicas emparejadas;
6. aleatorización de orden de arquitecturas;
7. exportación a dataset final por `escenario-arquitectura-réplica`;
8. script de análisis estadístico;
9. diagnóstico de modelo mixto;
10. reporte de contrastes BFMA vs G1 y BFMA vs G2.

Solo con esa etapa se puede avanzar hacia evidencia confirmatoria de la hipótesis.
