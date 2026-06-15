# BFMA OpenCode Lab

Piloto técnico para comparar tres estrategias de memoria en agentes ejecutados con OpenCode:

- G1: resumen incremental tradicional.
- G2: memoria persistente con store aislada.
- G3: BFMA con utilidad, relevancia y olvido presupuestado.

## Ejecución

```bash
go run ./cmd/bfma-pilot --scenario scenario_01 --groups g1,g2,g3 --reps 2
```

Para validar la plomería sin llamar a OpenCode/LLM. Este modo NO produce resultados experimentales válidos:

```bash
go run ./cmd/bfma-pilot --scenario scenario_01 --groups g1,g2,g3 --reps 1 --simulate-llm
```


Para ejecutar el piloto real con el modelo recomendado:

```bash
go run ./cmd/bfma-pilot --scenario scenario_01 --groups g1,g2,g3 --reps 1 --model opencode-go/deepseek-v4-flash
```

Para el escenario longitudinal de estrés:

```bash
go run ./cmd/bfma-pilot \
  --scenario scenario_03_gestion_academica_longitudinal \
  --groups g1,g2,g3 \
  --reps 1 \
  --model opencode-go/deepseek-v4-flash \
  --retries 2 \
  --timeout 10m
```

Si OpenCode o el proveedor cortan una corrida, podés retomar un grupo desde un turno reconstruyendo la memoria previa sin llamar al LLM:

```bash
go run ./cmd/bfma-pilot \
  --scenario scenario_03_gestion_academica_longitudinal \
  --groups g3 \
  --from-turn 11 \
  --reps 1 \
  --model opencode-go/deepseek-v4-flash \
  --retries 2 \
  --timeout 10m
```

## Reporte visual para asesor/anexos

Para generar un dashboard HTML autocontenido desde los logs JSONL:

```bash
go run ./cmd/bfma-report \
  --run logs/run_20260615_122925 \
  --out reports/run_20260615_122925/index.html
```

Para un reporte comparativo G1/G2/G3, usar un run que contenga los tres grupos:

```bash
go run ./cmd/bfma-report \
  --run logs/run_20260615_172800 \
  --out reports/run_20260615_172800/index.html
```

El reporte se puede abrir directamente en el navegador y permite imprimir/guardar como PDF o tomar screenshots para anexos de tesis.

## Principio metodológico

Los agentes no devuelven métricas ni logs de memoria. El runner controla el contexto, observa OpenCode y registra eventos experimentales en JSONL.
