# svc-trading-api
svc-trading-api는 거래 요청(Order) 접수/조회 REST API입니다.
Istio mTLS, JWT 경계 뒤에서 동작하며, Flagger Canary 기반 점진 배포를 사용합니다.

---

## 역할과 흐름

- **Input:** HTTP(S) from svc-gateway (JWT validated)
- **Output**
    - Kafka (MSK) → [topic=orders.in](http://topic=orders.in/) (신규/취소 주문)
    - Aurora → orders 테이블 insert/update (상태 기록)
- **Observability:** /metrics(Prometheus), 구조화 Log(Loki), Trace(Tempo)
    
    ### Quick Start
    
    1. Docker build/push → [ghcr.io/2025-demo-01/svc-trading-api:0.1.0](http://ghcr.io/2025-demo-01/svc-trading-api:0.1.0)
    2. Argo CD sync (platform-argocd)
    3. Health: GET /healthz, Ready: GET /readyz
    4. Orders: POST /api/v1/trade/orders (Idempotency-Key 헤더 지원)

---

## **Namespace: trading**

**Argo CD:** sync-wave=30 (mesh10 → policy20 → services30~60)
**Secrets:** External Secrets Operator(SSM → Secret)로 KAFKA_BROKERS, DB_ENDPOINT 주입

---

## 주요 기능

- Readiness(/readyz): Kafka + DB 둘 다 OK일 때만 200 (진짜 준비상태)
- Idempotency-Key: 중복 주문 방지 (클라이언트 재시도 안전)
- Resilience: RDS Proxy(권장), DestinationRule Circuit-Breaker/Outlier
- HPA: CPU + (선택) KEDA Kafka lag 기반 스케일
- PDB + TopologySpread로 장애 내성 보장
- SLO Alert: P95 latency & 5xx rate
- Supply Chain: SBOM(Trivy) + (옵션) Cosign verify

---
