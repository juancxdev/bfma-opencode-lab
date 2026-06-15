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

## Principio metodológico

Los agentes no devuelven métricas ni logs de memoria. El runner controla el contexto, observa OpenCode y registra eventos experimentales en JSONL.
