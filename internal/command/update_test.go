package command

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDaysFromStartOfDayUntilEndOfSunday(t *testing.T) {
	assert := assert.New(t)

	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		panic(err)
	}

	sunday := time.Date(2024, 6, 2, 18, 7, 0, 0, loc)
	monday := time.Date(2024, 6, 3, 18, 7, 0, 0, loc)
	tuesday := time.Date(2024, 6, 4, 18, 7, 0, 0, loc)
	wednesday := time.Date(2024, 6, 5, 18, 7, 0, 0, loc)
	thursday := time.Date(2024, 6, 6, 18, 7, 0, 0, loc)
	friday := time.Date(2024, 6, 7, 18, 7, 0, 0, loc)
	saturday := time.Date(2024, 6, 8, 18, 7, 0, 0, loc)

	assert.Equal(1, daysFromStartOfDayUntilEndOfSunday(sunday))
	assert.Equal(7, daysFromStartOfDayUntilEndOfSunday(monday))
	assert.Equal(6, daysFromStartOfDayUntilEndOfSunday(tuesday))
	assert.Equal(5, daysFromStartOfDayUntilEndOfSunday(wednesday))
	assert.Equal(4, daysFromStartOfDayUntilEndOfSunday(thursday))
	assert.Equal(3, daysFromStartOfDayUntilEndOfSunday(friday))
	assert.Equal(2, daysFromStartOfDayUntilEndOfSunday(saturday))
}
