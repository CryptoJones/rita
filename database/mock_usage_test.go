package database

import (
	"context"
	"testing"

	clickhouse "github.com/ClickHouse/clickhouse-go/v2"
	"go.uber.org/mock/gomock"
)

// TestMockDatabaseImplementsInterface verifies the generated MockDatabase
// satisfies the Database interface (including its unexported getConn method,
// which is why the mock must live in-package) and works with gomock
// expectations. This is the seam that lets future tests drive BulkWriter without
// a live ClickHouse connection.
func TestMockDatabaseImplementsInterface(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockDatabase(ctrl)

	// compile-time assertion that the mock satisfies the interface
	var db Database = mock

	ctx := context.Background()
	mock.EXPECT().GetContext().Return(ctx)
	mock.EXPECT().QueryParameters(gomock.Any()).Return(ctx)

	if got := db.GetContext(); got != ctx {
		t.Fatalf("GetContext returned unexpected context")
	}
	if got := db.QueryParameters(clickhouse.Parameters{"database": "test"}); got != ctx {
		t.Fatalf("QueryParameters returned unexpected context")
	}
}
