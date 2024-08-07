package regularity

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExtractRegularity(t *testing.T) {
	cases := map[string]time.Duration{
		"1 день":     day(1),
		"2 дня":      day(2),
		"10 дней":    day(10),
		"1 неделя":   week(1),
		"2 недели":   week(2),
		"10 недель":  week(10),
		"1 месяц":    month(1),
		"2 месяца":   month(2),
		"10 месяцев": month(10),
	}
	for key, value := range cases {
		t.Run(fmt.Sprintf("test %s", key), func(t *testing.T) {
			res, err := ExtractRegularity(key)
			assert.NoError(t, err)
			assert.Equal(t, res, value)
		})
	}
}
