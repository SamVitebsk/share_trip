package service_test

import (
	"context"
	"log"
	"os"
	"testing"

	"share_trip/internal/test"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	testCtx  context.Context
	testPool *pgxpool.Pool
)

func TestMain(m *testing.M) {
	testCtx = context.Background()

	var err error
	var teardown func()

	testPool, _, teardown, err = test.SetupTestDB(testCtx, "../../migrations")
	if err != nil {
		log.Fatalf("failed to setup test database: %v", err)
	}
	defer teardown()

	os.Exit(m.Run())
}
