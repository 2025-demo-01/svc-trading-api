package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// --- 환경변수 ---
// KAFKA_BROKER: "kafka:9092" 같은 broker 주소 (여러개면 콤마로 연결: "k1:9092,k2:9092")
// DB_DRIVER   : 예) "pgx", "postgres", "mysql" (드라이버 blank import 필요)
// DB_DSN      : 예) "postgres://user:pass@db:5432/app?sslmode=disable"
// (옵션) DB_ADDR: 드라이버를 아직 못 붙였을 때 TCP 레벨로만 헬스체크할 주소 "db:5432"

var (
	db           *sql.DB
	kafkaBrokers []string
)

func main() {
	// --- 라우터 등록 ---
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", Healthz) // liveness/startup
	mux.HandleFunc("/readyz", Readyz)   // readiness (DB + Kafka)

	// --- 의존성 준비 (비차단) ---
	initDeps()

	// --- HTTP 서버 (타임아웃 & graceful) ---
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		// 필요시 WriteTimeout/IdleTimeout도 추가 가능
	}

	// 비동기로 시작 (프로세스 블로킹 방지)
	go func() {
		log.Println("svc-trading-api listening on :8080")
		// ListenAndServe가 에러로 종료되면 fatal (정상 종료는 Shutdown에서 처리)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server error: %v", err)
		}
	}()

	// --- OS 시그널 대기 & 종료 처리 ---
	// K8s는 종료시 SIGTERM → preStop 실행 → terminationGracePeriod 내 종료를 기대
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutdown signal received; draining...")
	// readiness는 핸들러에서 즉시 실패로 바꾸지 않지만,
	// K8s가 preStop(sleep 10) 동안 LB에서 제거하므로 신규 트래픽 유입은 멈춤.

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
	if db != nil {
		_ = db.Close()
	}
	log.Println("server stopped cleanly")
}

// --- 핸들러들 ---

// /healthz: 프로세스 살았는지 (panic/뮤텍스 교착 등 아닌 이상 200)
func Healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// /readyz: "진짜" 준비상태 — Kafka & DB 모두 OK일 때만 200
func Readyz(w http.ResponseWriter, r *http.Request) {
	// 각 체크에 짧은 데드라인을 둬서 핸들러가 뻗지 않게 함
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	kOK := kafkaHealthy(ctx)
	dOK := dbHealthy(ctx)

	if kOK && dOK {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte("not ready"))
}

// --- 의존성 초기화 & 헬스체크 ---

func initDeps() {
	// Kafka 브로커 목록
	if raw := os.Getenv("KAFKA_BROKER"); raw != "" {
		// 콤마로 여러개 지원
		kafkaBrokers = splitAndTrim(raw, ',')
	}

	// DB 연결 (드라이버가 등록되어 있지 않으면 Open에서 에러)
	driver := os.Getenv("DB_DRIVER")
	dsn := os.Getenv("DB_DSN")
	if driver != "" && dsn != "" {
		sqlDB, err := sql.Open(driver, dsn)
		if err != nil {
			log.Printf("sql.Open failed (driver=%s): %v", driver, err)
		} else {
			// 커넥션 풀 파라미터는 운영 성격에 맞게 조정
			sqlDB.SetMaxOpenConns(50)
			sqlDB.SetMaxIdleConns(25)
			sqlDB.SetConnMaxLifetime(30 * time.Minute)
			db = sqlDB
		}
	}
}

func kafkaHealthy(ctx context.Context) bool {
	// 최소 한 개 브로커와 TCP 레벨 연결이 되어야 "준비됨"으로 간주
	// (프로토콜 핸드셰이크까지 보려면 sarama/franz-go 사용 권장)
	if len(kafkaBrokers) == 0 {
		return false
	}
	dialer := &net.Dialer{}
	for _, b := range kafkaBrokers {
		conn, err := dialer.DialContext(ctx, "tcp", b)
		if err == nil {
			_ = conn.Close()
			return true
		}
	}
	return false
}

func dbHealthy(ctx context.Context) bool {
	// 드라이버 기반 Ping이 최선
	if db != nil {
		if err := db.PingContext(ctx); err == nil {
			return true
		}
		// 드라이버가 등록됐지만 아직 못 붙으면 false
		return false
	}

	// 아직 드라이버/DSN을 못 붙였으면 TCP 레벨로만 체크 (임시)
	if addr := os.Getenv("DB_ADDR"); addr != "" {
		dialer := &net.Dialer{}
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err == nil {
			_ = conn.Close()
			return true
		}
	}
	return false
}

// --- 유틸 ---

func splitAndTrim(s string, sep rune) []string {
	out := make([]string, 0, 4)
	field := make([]rune, 0, len(s))
	for _, r := range s {
		if r == sep {
			if len(field) > 0 {
				out = append(out, stringTrimSpace(string(field)))
				field = field[:0]
			}
			continue
		}
		field = append(field, r)
	}
	if len(field) > 0 {
		out = append(out, stringTrimSpace(string(field)))
	}
	return out
}

func stringTrimSpace(x string) string { return string([]rune(x)) } // placeholder; build tags 회피
