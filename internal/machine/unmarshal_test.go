package machine

import (
	"os"
	"testing"
	"time"

	"github.com/leg100/pug/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshal(t *testing.T) {
	f, err := os.Open("./testdata/plan_with_changes.json")
	require.NoError(t, err)

	got, err := Unmarshal(f)
	require.NoError(t, err)

	assert.Len(t, got, 6)
	assert.Equal(t, &VersionMsg{
		Common: Common{
			Type:      MessageVersion,
			Level:     "info",
			Message:   "Terraform 1.9.8",
			Module:    "terraform.ui",
			TimeStamp: testutils.ParseTime(t, time.RFC3339, "2025-01-12T09:58:25.947800Z"),
		},
		Terraform: "1.9.8",
		UI:        "1.2",
	}, got[0])
	//assert.Equal(t, &PlannedChangeMsg{
	//	Common: Common{
	//		TypeMsg:   TypeMsg{Type: "planned_change"},
	//		Level:     "info",
	//		Message:   "null_resource.null2: Plan to create",
	//		Module:    "terraform.ui",
	//		TimeStamp: testutils.ParseTime(t, time.RFC3339, "2025-01-12T09:58:26.502560Z"),
	//	},
	//	Change: &ResourceInstanceChange{},
	//}, got[1])
}
