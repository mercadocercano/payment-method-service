//go:build integration

// Test de integración de la semántica de is_active en los tres planos (MC-E37 T8).
//
// Verifica el doble filtro de habilitación del POS (decisión de diseño del gate T1 / T6):
//   - is_active es PER-TENANT: deshabilitar la fila del tenant A NO afecta la del tenant B
//     ni el catálogo global.
//   - is_active GLOBAL manda: un método retirado a nivel global desaparece para todos los
//     tenants, aunque el tenant lo tuviera habilitado.
//
// Corre contra una DB Postgres real con el esquema de MC-E37 aplicado (los tres planos). Usa
// datos throwaway (un global de prueba + dos tenants random) y los limpia al terminar — no toca
// los 9 globales ni el backfill del lab. La DSN se lee de PAYMENT_TEST_DSN (superuser, para poder
// sembrar el global de prueba); si no está seteada, el test se omite.
//
// Para correrlo (dentro del contenedor, contra lab-postgres):
//
//	docker exec -e PAYMENT_TEST_DSN="postgres://postgres:postgres@lab-postgres:5432/payment_method_db?sslmode=disable" \
//	  mc-payment-method-service sh -c 'cd /app && go test -tags=integration ./test/... -run IsActive -v'
package persistence_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"payment_method/src/payment_method/domain/entity"
	"payment_method/src/payment_method/infrastructure/persistence/repository"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("PAYMENT_TEST_DSN")
	if dsn == "" {
		t.Skip("PAYMENT_TEST_DSN no seteado — test de integración omitido")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("ping: %v", err)
	}
	return db
}

// hasCode indica si la lista contiene un método con ese code (code viene del global vía JOIN).
func hasCode(list []*entity.PaymentMethod, code string) bool {
	for _, pm := range list {
		if pm.Code == code {
			return true
		}
	}
	return false
}

func TestPaymentMethods_IsActive_PerTenantAndGlobalOverride(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	// --- Setup: un global throwaway + un global de control (cash) + dos tenants ---
	gTest := uuid.New()
	if _, err := db.Exec(
		`INSERT INTO global_payment_methods (id, code, name, is_active) VALUES ($1, $2, $3, true)`,
		gTest, "zzz_t8_"+gTest.String()[:8], "T8 Test Method",
	); err != nil {
		t.Fatalf("seed global throwaway: %v", err)
	}

	var gCash uuid.UUID
	if err := db.QueryRow(`SELECT id FROM global_payment_methods WHERE code = 'cash'`).Scan(&gCash); err != nil {
		t.Fatalf("global cash de control: %v", err)
	}

	tenantA := uuid.New()
	tenantB := uuid.New()

	if _, err := db.Exec(`
		INSERT INTO payment_methods (tenant_id, global_payment_method_id, is_active) VALUES
			($1, $2, true),   -- A × gTest (activo)
			($1, $3, true),   -- A × cash  (control, siempre visible)
			($4, $2, true)    -- B × gTest (activo)
	`, tenantA, gTest, gCash, tenantB); err != nil {
		t.Fatalf("seed payment_methods: %v", err)
	}

	// Cleanup: primero las filas tenant (FK), después el global throwaway.
	defer func() {
		if _, err := db.Exec(`DELETE FROM payment_methods WHERE tenant_id = ANY($1)`,
			"{"+tenantA.String()+","+tenantB.String()+"}"); err != nil {
			t.Errorf("cleanup payment_methods: %v", err)
		}
		if _, err := db.Exec(`DELETE FROM global_payment_methods WHERE id = $1`, gTest); err != nil {
			t.Errorf("cleanup global throwaway: %v", err)
		}
	}()

	repo := repository.NewPostgresPaymentMethodRepository(db)
	codeTest := "zzz_t8_" + gTest.String()[:8]

	// --- 1) is_active PER-TENANT: deshabilitar la fila de A no toca a B ni al global ---
	if _, err := db.Exec(
		`UPDATE payment_methods SET is_active = false WHERE tenant_id = $1 AND global_payment_method_id = $2`,
		tenantA, gTest,
	); err != nil {
		t.Fatalf("disable A×gTest: %v", err)
	}

	aList, err := repo.FindAll(tenantA, true)
	if err != nil {
		t.Fatalf("FindAll(A): %v", err)
	}
	if hasCode(aList, codeTest) {
		t.Errorf("A: gTest deshabilitado por el tenant NO debería listarse")
	}
	if !hasCode(aList, "cash") {
		t.Errorf("A: cash (habilitado) debería seguir listándose")
	}

	bList, err := repo.FindAll(tenantB, true)
	if err != nil {
		t.Fatalf("FindAll(B): %v", err)
	}
	if !hasCode(bList, codeTest) {
		t.Errorf("B: gTest debería seguir visible — deshabilitarlo en A no puede afectar a B")
	}

	var globalActive bool
	if err := db.QueryRow(`SELECT is_active FROM global_payment_methods WHERE id = $1`, gTest).Scan(&globalActive); err != nil {
		t.Fatalf("check global is_active: %v", err)
	}
	if !globalActive {
		t.Errorf("el global NO debería haberse tocado al deshabilitar la fila del tenant A")
	}

	// --- 2) is_active GLOBAL manda: retirarlo del global lo saca para todos ---
	if _, err := db.Exec(
		`UPDATE payment_methods SET is_active = true WHERE tenant_id = $1 AND global_payment_method_id = $2`,
		tenantA, gTest,
	); err != nil {
		t.Fatalf("re-enable A×gTest: %v", err)
	}
	if _, err := db.Exec(`UPDATE global_payment_methods SET is_active = false WHERE id = $1`, gTest); err != nil {
		t.Fatalf("disable gTest global: %v", err)
	}

	bList2, err := repo.FindAll(tenantB, true)
	if err != nil {
		t.Fatalf("FindAll(B) tras retiro global: %v", err)
	}
	if hasCode(bList2, codeTest) {
		t.Errorf("gTest retirado a nivel global NO debería listarse para ningún tenant (aunque la fila del tenant esté activa)")
	}
}
